package codex

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ParsedCodexOutput holds the aggregated result of a Codex JSONL stream.
type ParsedCodexOutput struct {
	SessionID         string
	Summary           string
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
	ErrorMessage      string
}

var codexUnknownSessionRe = regexp.MustCompile(
	`(?i)unknown (session|thread)|session .* not found|thread .* not found|conversation .* not found|missing rollout path for thread|state db missing rollout path`,
)

// ParseCodexJSONL scans JSONL output from the codex CLI.
func ParseCodexJSONL(stdout string) ParsedCodexOutput {
	var result ParsedCodexOutput
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

		switch evType {
		case "thread.started":
			if sid, _ := ev["thread_id"].(string); sid != "" {
				result.SessionID = sid
			}

		case "error":
			if msg, _ := ev["message"].(string); msg != "" {
				result.ErrorMessage = strings.TrimSpace(msg)
			}

		case "item.completed":
			if item, ok := ev["item"].(map[string]interface{}); ok {
				if t, _ := item["type"].(string); t == "agent_message" {
					if text, _ := item["text"].(string); text != "" {
						messages = append(messages, text)
					}
				}
			}

		case "turn.completed":
			if usage, ok := ev["usage"].(map[string]interface{}); ok {
				result.InputTokens = asInt(usage["input_tokens"])
				result.CachedInputTokens = asInt(usage["cached_input_tokens"])
				result.OutputTokens = asInt(usage["output_tokens"])
			}

		case "turn.failed":
			if errObj, ok := ev["error"].(map[string]interface{}); ok {
				if msg, _ := errObj["message"].(string); msg != "" {
					result.ErrorMessage = strings.TrimSpace(msg)
				}
			}
		}
	}

	result.Summary = strings.TrimSpace(strings.Join(messages, "\n\n"))
	return result
}

// IsUnknownSessionError returns true when Codex cannot resume the given session.
func IsUnknownSessionError(stdout, stderr string) bool {
	return codexUnknownSessionRe.MatchString(stdout + "\n" + stderr)
}

func asInt(v interface{}) int {
	if f, ok := v.(float64); ok {
		return int(f)
	}
	return 0
}
