package pi

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ParsedPiOutput holds the aggregated result of a Pi JSONL stream.
type ParsedPiOutput struct {
	SessionID         string
	Summary           string
	FinalMessage      string
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
	CostUSD           float64
	ErrorMessage      string
}

var piUnknownSessionRe = regexp.MustCompile(
	`(?i)unknown\s+session|session\s+not\s+found|session\s+.*\s+not\s+found|no\s+session`,
)

// ParsePiJSONL scans JSONL output from the Pi CLI.
func ParsePiJSONL(stdout string) ParsedPiOutput {
	var result ParsedPiOutput
	var messages []string

	for _, rawLine := range strings.Split(stdout, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		var ev map[string]interface{}
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}

		evType, _ := ev["type"].(string)

		// Skip RPC / internal protocol frames.
		switch evType {
		case "response", "extension_ui_request", "extension_ui_response", "extension_error", "agent_start":
			continue
		}

		switch evType {
		case "agent_end":
			msgs, _ := ev["messages"].([]interface{})
			if len(msgs) > 0 {
				last, ok := msgs[len(msgs)-1].(map[string]interface{})
				if ok && last["role"] == "assistant" {
					if text := extractTextContent(last["content"]); text != "" {
						result.FinalMessage = text
					}
				}
			}

		case "turn_end":
			if msg, ok := ev["message"].(map[string]interface{}); ok {
				text := extractTextContent(msg["content"])
				if text != "" {
					result.FinalMessage = text
					messages = append(messages, text)
				}
				if usage, ok := msg["usage"].(map[string]interface{}); ok {
					result.InputTokens += asInt(usage["input"])
					result.OutputTokens += asInt(usage["output"])
					result.CachedInputTokens += asInt(usage["cacheRead"])
					if cost, ok := usage["cost"].(map[string]interface{}); ok {
						result.CostUSD += asFloat(cost["total"])
					}
				}
			}

		case "message_update":
			if ae, ok := ev["assistantMessageEvent"].(map[string]interface{}); ok {
				if t, _ := ae["type"].(string); t == "text_delta" {
					if delta, _ := ae["delta"].(string); delta != "" {
						if len(messages) == 0 {
							messages = append(messages, delta)
						} else {
							messages[len(messages)-1] += delta
						}
					}
				}
			}

		case "usage":
			if usage, ok := ev["usage"].(map[string]interface{}); ok {
				result.InputTokens += asInt(firstOf(usage, "inputTokens", "input"))
				result.OutputTokens += asInt(firstOf(usage, "outputTokens", "output"))
				result.CachedInputTokens += asInt(firstOf(usage, "cachedInputTokens", "cacheRead"))
				if cost, ok := usage["cost"].(map[string]interface{}); ok {
					result.CostUSD += asFloat(firstOf(cost, "total", "costUsd"))
				} else {
					result.CostUSD += asFloat(usage["costUsd"])
				}
			}
		}
	}

	result.Summary = strings.TrimSpace(strings.Join(messages, "\n\n"))
	return result
}

// IsUnknownSessionError returns true when Pi cannot resume the given session.
func IsUnknownSessionError(stdout, stderr string) bool {
	return piUnknownSessionRe.MatchString(stdout + "\n" + stderr)
}

func extractTextContent(content interface{}) string {
	if s, ok := content.(string); ok {
		return strings.TrimSpace(s)
	}
	items, ok := content.([]interface{})
	if !ok {
		return ""
	}
	var parts []string
	for _, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		if t, _ := item["type"].(string); t == "text" {
			if text, _ := item["text"].(string); text != "" {
				parts = append(parts, text)
			}
		}
	}
	return strings.TrimSpace(strings.Join(parts, ""))
}

func firstOf(m map[string]interface{}, keys ...string) interface{} {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func asInt(v interface{}) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}

func asFloat(v interface{}) float64 {
	if f, ok := v.(float64); ok {
		return f
	}
	return 0
}
