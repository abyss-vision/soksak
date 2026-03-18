package gemini

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ParsedGeminiOutput holds the aggregated result of a Gemini JSONL stream.
type ParsedGeminiOutput struct {
	SessionID         string
	Summary           string
	InputTokens       int
	CachedInputTokens int
	OutputTokens      int
	CostUSD           float64
	ErrorMessage      string
}

var geminiAuthRe = regexp.MustCompile(
	`(?i)(?:not\s+authenticated|please\s+authenticate|api[_ ]?key\s+(?:required|missing|invalid)|authentication\s+required|unauthorized|invalid\s+credentials|not\s+logged\s+in|login\s+required|run\s+` + "`?" + `gemini\s+auth(?:\s+login)?` + "`?" + `\s+first)`,
)
var geminiUnknownSessionRe = regexp.MustCompile(
	`(?i)unknown\s+session|session\s+.*\s+not\s+found|resume\s+.*\s+not\s+found|checkpoint\s+.*\s+not\s+found|cannot\s+resume|failed\s+to\s+resume`,
)

// ParseGeminiJSONL scans JSONL output from the Gemini CLI.
func ParseGeminiJSONL(stdout string) ParsedGeminiOutput {
	var result ParsedGeminiOutput
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

		for _, key := range []string{"session_id", "sessionId", "sessionID", "checkpoint_id", "thread_id"} {
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
			accumulateUsage(&result, firstOf(ev, "usage", "usageMetadata"))

		case "result":
			accumulateUsage(&result, firstOf(ev, "usage", "usageMetadata"))
			result.CostUSD += asFloat(firstOf(ev, "total_cost_usd", "cost_usd", "cost"))
			r := strings.TrimSpace(anyString(firstOf(ev, "result", "text", "response")))
			if r != "" && len(messages) == 0 {
				messages = append(messages, r)
			}
			isError := ev["is_error"] == true
			if sub, _ := ev["subtype"].(string); strings.ToLower(sub) == "error" {
				isError = true
			}
			if isError {
				result.ErrorMessage = extractErrorText(firstOf(ev, "error", "message", "result"))
			}

		case "error":
			result.ErrorMessage = extractErrorText(firstOf(ev, "error", "message", "detail"))

		case "system":
			if sub, _ := ev["subtype"].(string); strings.ToLower(sub) == "error" {
				result.ErrorMessage = extractErrorText(firstOf(ev, "error", "message", "detail"))
			}

		case "text":
			if part, ok := ev["part"].(map[string]interface{}); ok {
				if t, _ := part["text"].(string); t != "" {
					messages = append(messages, strings.TrimSpace(t))
				}
			}

		case "step_finish":
			accumulateUsage(&result, firstOf(ev, "usage", "usageMetadata"))
			result.CostUSD += asFloat(firstOf(ev, "total_cost_usd", "cost_usd", "cost"))

		default:
			// accumulate usage fields if present on unknown event types
			if u := firstOf(ev, "usage", "usageMetadata"); u != nil {
				accumulateUsage(&result, u)
			}
		}
	}

	result.Summary = strings.TrimSpace(strings.Join(messages, "\n\n"))
	return result
}

// DetectAuthRequired checks stdout/stderr for Gemini authentication errors.
func DetectAuthRequired(stdout, stderr string) bool {
	combined := stdout + "\n" + stderr
	for _, line := range strings.Split(combined, "\n") {
		if geminiAuthRe.MatchString(strings.TrimSpace(line)) {
			return true
		}
	}
	return false
}

// IsUnknownSessionError returns true when Gemini cannot resume the given session.
func IsUnknownSessionError(stdout, stderr string) bool {
	return geminiUnknownSessionRe.MatchString(stdout + "\n" + stderr)
}

func accumulateUsage(r *ParsedGeminiOutput, raw interface{}) {
	if raw == nil {
		return
	}
	usage, _ := raw.(map[string]interface{})
	if usage == nil {
		return
	}
	meta, _ := usage["usageMetadata"].(map[string]interface{})
	src := usage
	if len(meta) > 0 {
		src = meta
	}
	r.InputTokens += asInt(firstOf(src, "input_tokens", "inputTokens", "promptTokenCount"))
	r.CachedInputTokens += asInt(firstOf(src, "cached_input_tokens", "cachedInputTokens", "cachedContentTokenCount"))
	r.OutputTokens += asInt(firstOf(src, "output_tokens", "outputTokens", "candidatesTokenCount"))
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
		if pt == "output_text" || pt == "text" || pt == "content" {
			t := anyString(firstOf(part, "text", "content"))
			if t != "" {
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

func firstOf(m map[string]interface{}, keys ...string) interface{} {
	for _, k := range keys {
		if v, ok := m[k]; ok && v != nil {
			return v
		}
	}
	return nil
}

func anyString(v interface{}) string {
	s, _ := v.(string)
	return s
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
