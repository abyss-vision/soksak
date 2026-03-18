package opencode

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ParsedOpenCodeOutput holds the aggregated result of an OpenCode JSONL stream.
type ParsedOpenCodeOutput struct {
	SessionID         string
	Summary           string
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
	CostUSD           float64
	ErrorMessage      string
}

var openCodeUnknownSessionRe = regexp.MustCompile(
	`(?i)unknown\s+session|session\b.*\bnot\s+found|resource\s+not\s+found:.*[/\\]session[/\\].*\.json|notfounderror|no session`,
)

// ParseOpenCodeJSONL scans JSONL output from the opencode CLI.
func ParseOpenCodeJSONL(stdout string) ParsedOpenCodeOutput {
	var result ParsedOpenCodeOutput
	var messages []string
	var errs []string

	for _, rawLine := range strings.Split(stdout, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		var ev map[string]interface{}
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}

		if sid, _ := ev["sessionID"].(string); sid != "" {
			result.SessionID = sid
		}

		evType, _ := ev["type"].(string)

		switch evType {
		case "text":
			if part, ok := ev["part"].(map[string]interface{}); ok {
				if t, _ := part["text"].(string); t != "" {
					messages = append(messages, strings.TrimSpace(t))
				}
			}

		case "step_finish":
			if part, ok := ev["part"].(map[string]interface{}); ok {
				if tokens, ok := part["tokens"].(map[string]interface{}); ok {
					result.InputTokens += asInt(tokens["input"])
					if cache, ok := tokens["cache"].(map[string]interface{}); ok {
						result.CachedInputTokens += asInt(cache["read"])
					}
					result.OutputTokens += asInt(tokens["output"])
					result.OutputTokens += asInt(tokens["reasoning"])
				}
				result.CostUSD += asFloat(part["cost"])
			}

		case "tool_use":
			if part, ok := ev["part"].(map[string]interface{}); ok {
				if state, ok := part["state"].(map[string]interface{}); ok {
					if s, _ := state["status"].(string); s == "error" {
						if t, _ := state["error"].(string); t != "" {
							errs = append(errs, strings.TrimSpace(t))
						}
					}
				}
			}

		case "error":
			if text := errorText(firstOf(ev, "error", "message")); text != "" {
				errs = append(errs, text)
			}
		}
	}

	result.Summary = strings.TrimSpace(strings.Join(messages, "\n\n"))
	if len(errs) > 0 {
		result.ErrorMessage = strings.Join(errs, "\n")
	}
	return result
}

// IsUnknownSessionError returns true when OpenCode cannot resume the given session.
func IsUnknownSessionError(stdout, stderr string) bool {
	return openCodeUnknownSessionRe.MatchString(stdout + "\n" + stderr)
}

func errorText(v interface{}) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	rec, ok := v.(map[string]interface{})
	if !ok {
		return ""
	}
	for _, key := range []string{"message", "data", "name", "code"} {
		if key == "data" {
			if d, ok := rec["data"].(map[string]interface{}); ok {
				if msg, _ := d["message"].(string); msg != "" {
					return strings.TrimSpace(msg)
				}
			}
			continue
		}
		if s, _ := rec[key].(string); s != "" {
			return strings.TrimSpace(s)
		}
	}
	b, _ := json.Marshal(rec)
	return string(b)
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
