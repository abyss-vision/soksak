package cursor

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ParsedCursorOutput holds the aggregated result of a Cursor JSONL stream.
type ParsedCursorOutput struct {
	SessionID         string
	Summary           string
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
	CostUSD           float64
	ErrorMessage      string
}

var cursorUnknownSessionRe = regexp.MustCompile(
	`(?i)unknown\s+(session|chat)|session\s+.*\s+not\s+found|chat\s+.*\s+not\s+found|resume\s+.*\s+not\s+found|could\s+not\s+resume`,
)

// ParseCursorJSONL scans JSONL output from the Cursor background agent CLI.
func ParseCursorJSONL(stdout string) ParsedCursorOutput {
	var result ParsedCursorOutput
	var messages []string

	for _, rawLine := range strings.Split(stdout, "\n") {
		line := strings.TrimSpace(normalizeLine(rawLine))
		if line == "" {
			continue
		}

		var ev map[string]interface{}
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}

		// Track session id from any event that carries it.
		for _, key := range []string{"session_id", "sessionId", "sessionID"} {
			if s, _ := ev[key].(string); s != "" {
				result.SessionID = s
				break
			}
		}

		evType, _ := ev["type"].(string)

		switch evType {
		case "assistant":
			texts := collectMessageText(ev["message"])
			messages = append(messages, texts...)

		case "result":
			if usage, ok := ev["usage"].(map[string]interface{}); ok {
				result.InputTokens += asInt(firstOf(usage, "input_tokens", "inputTokens"))
				result.CachedInputTokens += asInt(firstOf(usage, "cached_input_tokens", "cachedInputTokens", "cache_read_input_tokens"))
				result.OutputTokens += asInt(firstOf(usage, "output_tokens", "outputTokens"))
			}
			result.CostUSD += asFloat(firstOf(ev, "total_cost_usd", "cost_usd", "cost"))
			isError := ev["is_error"] == true
			if sub, _ := ev["subtype"].(string); strings.ToLower(sub) == "error" {
				isError = true
			}
			if isError {
				result.ErrorMessage = extractErrorText(firstOf(ev, "error", "message", "result"))
			} else if len(messages) == 0 {
				if r, _ := ev["result"].(string); r != "" {
					messages = append(messages, r)
				}
			}

		case "error":
			result.ErrorMessage = extractErrorText(firstOf(ev, "message", "error", "detail"))

		case "system":
			if sub, _ := ev["subtype"].(string); strings.ToLower(sub) == "error" {
				result.ErrorMessage = extractErrorText(firstOf(ev, "message", "error", "detail"))
			}

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
				}
				result.CostUSD += asFloat(part["cost"])
			}
		}
	}

	result.Summary = strings.TrimSpace(strings.Join(messages, "\n\n"))
	return result
}

// IsUnknownSessionError returns true when Cursor cannot resume the given session.
func IsUnknownSessionError(stdout, stderr string) bool {
	return cursorUnknownSessionRe.MatchString(stdout + "\n" + stderr)
}

// normalizeLine strips any SSE "data: " prefix that Cursor may emit.
func normalizeLine(line string) string {
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "data: ") {
		return strings.TrimPrefix(line, "data: ")
	}
	return line
}

func collectMessageText(message interface{}) []string {
	if s, ok := message.(string); ok {
		if t := strings.TrimSpace(s); t != "" {
			return []string{t}
		}
		return nil
	}
	rec, ok := message.(map[string]interface{})
	if !ok {
		return nil
	}
	var out []string
	if t, _ := rec["text"].(string); t != "" {
		out = append(out, strings.TrimSpace(t))
	}
	content, _ := rec["content"].([]interface{})
	for _, partRaw := range content {
		part, ok := partRaw.(map[string]interface{})
		if !ok {
			continue
		}
		pt, _ := part["type"].(string)
		if pt == "output_text" || pt == "text" {
			if t, _ := part["text"].(string); t != "" {
				out = append(out, strings.TrimSpace(t))
			}
		}
	}
	return out
}

func extractErrorText(v interface{}) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	rec, ok := v.(map[string]interface{})
	if !ok {
		return ""
	}
	for _, key := range []string{"message", "error", "code", "detail"} {
		if s, _ := rec[key].(string); s != "" {
			return strings.TrimSpace(s)
		}
	}
	b, _ := json.Marshal(rec)
	return string(b)
}

// firstOf returns the first non-nil value found among the given keys.
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
