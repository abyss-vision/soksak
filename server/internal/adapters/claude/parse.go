package claude

import (
	"encoding/json"
	"regexp"
	"strings"
)

var claudeAuthRe = regexp.MustCompile(
	`(?i)(?:not\s+logged\s+in|please\s+log\s+in|please\s+run\s+` + "`?" + `claude\s+login` + "`?" + `|login\s+required|requires\s+login|unauthorized|authentication\s+required)`,
)
var urlRe = regexp.MustCompile(`https?://[^\s'"` + "`" + `<>()\[\]{};,!?]+[^\s'"` + "`" + `<>()\[\]{};,!.?:]`)

// ParsedStream holds the aggregated result of scanning a Claude stream-json stdout.
type ParsedStream struct {
	SessionID string
	Model     string
	Summary   string
	CostUSD   float64
	InputTokens        int
	CachedInputTokens  int
	OutputTokens       int
	ResultJSON         map[string]interface{}
}

// ParseClaudeStreamJSON scans newline-delimited JSON from Claude --output-format stream-json.
func ParseClaudeStreamJSON(stdout string) ParsedStream {
	var result ParsedStream
	var assistantTexts []string
	var finalEvent map[string]interface{}

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
		case "system":
			if sub, _ := ev["subtype"].(string); sub == "init" {
				if sid, _ := ev["session_id"].(string); sid != "" {
					result.SessionID = sid
				}
				if m, _ := ev["model"].(string); m != "" {
					result.Model = m
				}
			}

		case "assistant":
			if sid, _ := ev["session_id"].(string); sid != "" {
				result.SessionID = sid
			}
			if msg, ok := ev["message"].(map[string]interface{}); ok {
				if content, ok := msg["content"].([]interface{}); ok {
					for _, raw := range content {
						block, ok := raw.(map[string]interface{})
						if !ok {
							continue
						}
						if bt, _ := block["type"].(string); bt == "text" {
							if t, _ := block["text"].(string); t != "" {
								assistantTexts = append(assistantTexts, t)
							}
						}
					}
				}
			}

		case "result":
			finalEvent = ev
			if sid, _ := ev["session_id"].(string); sid != "" {
				result.SessionID = sid
			}
		}
	}

	if finalEvent == nil {
		result.Summary = strings.TrimSpace(strings.Join(assistantTexts, "\n\n"))
		return result
	}

	result.ResultJSON = finalEvent

	if usage, ok := finalEvent["usage"].(map[string]interface{}); ok {
		result.InputTokens = asInt(usage["input_tokens"])
		result.CachedInputTokens = asInt(usage["cache_read_input_tokens"])
		result.OutputTokens = asInt(usage["output_tokens"])
	}
	if c, ok := finalEvent["total_cost_usd"].(float64); ok {
		result.CostUSD = c
	}
	if s, _ := finalEvent["result"].(string); s != "" {
		result.Summary = strings.TrimSpace(s)
	} else {
		result.Summary = strings.TrimSpace(strings.Join(assistantTexts, "\n\n"))
	}

	return result
}

// DetectAuthRequired checks stdout/stderr for Claude login-required signals.
func DetectAuthRequired(stdout, stderr string) (required bool, loginURL string) {
	combined := stdout + "\n" + stderr
	for _, line := range strings.Split(combined, "\n") {
		line = strings.TrimSpace(line)
		if claudeAuthRe.MatchString(line) {
			required = true
		}
	}
	loginURL = extractLoginURL(stdout + "\n" + stderr)
	return
}

// IsMaxTurnsResult returns true when Claude stopped due to turn limit.
func IsMaxTurnsResult(parsed map[string]interface{}) bool {
	if parsed == nil {
		return false
	}
	if sub, _ := parsed["subtype"].(string); strings.ToLower(sub) == "error_max_turns" {
		return true
	}
	if sr, _ := parsed["stop_reason"].(string); strings.ToLower(sr) == "max_turns" {
		return true
	}
	if r, _ := parsed["result"].(string); regexp.MustCompile(`(?i)max(?:imum)?\s+turns?`).MatchString(r) {
		return true
	}
	return false
}

// IsUnknownSessionError returns true when Claude cannot resume the given session.
func IsUnknownSessionError(parsed map[string]interface{}) bool {
	if parsed == nil {
		return false
	}
	re := regexp.MustCompile(`(?i)no conversation found with session id|unknown session|session .* not found`)
	msgs := collectClaudeErrorMessages(parsed)
	if r, _ := parsed["result"].(string); r != "" {
		msgs = append(msgs, r)
	}
	for _, m := range msgs {
		if re.MatchString(m) {
			return true
		}
	}
	return false
}

func collectClaudeErrorMessages(parsed map[string]interface{}) []string {
	raw, _ := parsed["errors"].([]interface{})
	var out []string
	for _, entry := range raw {
		switch v := entry.(type) {
		case string:
			if v != "" {
				out = append(out, v)
			}
		case map[string]interface{}:
			for _, key := range []string{"message", "error", "code"} {
				if s, _ := v[key].(string); s != "" {
					out = append(out, s)
					break
				}
			}
		}
	}
	return out
}

func extractLoginURL(text string) string {
	matches := urlRe.FindAllString(text, -1)
	for _, raw := range matches {
		// strip trailing punctuation
		cleaned := strings.TrimRight(raw, `])}.!,?;:'"`)
		lower := strings.ToLower(cleaned)
		if strings.Contains(lower, "claude") || strings.Contains(lower, "anthropic") || strings.Contains(lower, "auth") {
			return cleaned
		}
	}
	if len(matches) > 0 {
		return strings.TrimRight(matches[0], `])}.!,?;:'"`)
	}
	return ""
}

func asInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	}
	return 0
}
