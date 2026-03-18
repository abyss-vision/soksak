package claude_test

import (
	"testing"

	"abyss-view/internal/adapters/claude"
)

func TestParseClaudeStreamJSON_Basic(t *testing.T) {
	stdout := `{"type":"system","subtype":"init","session_id":"sess1","model":"claude-opus-4-5"}
{"type":"assistant","session_id":"sess1","message":{"content":[{"type":"text","text":"Hello!"}]}}
{"type":"result","session_id":"sess1","result":"Done","usage":{"input_tokens":10,"output_tokens":5}}`

	p := claude.ParseClaudeStreamJSON(stdout)
	if p.SessionID != "sess1" {
		t.Errorf("got session %q, want %q", p.SessionID, "sess1")
	}
	if p.Model != "claude-opus-4-5" {
		t.Errorf("got model %q, want %q", p.Model, "claude-opus-4-5")
	}
	if p.Summary != "Done" {
		t.Errorf("got summary %q, want %q", p.Summary, "Done")
	}
	if p.InputTokens != 10 {
		t.Errorf("got input_tokens %d, want 10", p.InputTokens)
	}
	if p.OutputTokens != 5 {
		t.Errorf("got output_tokens %d, want 5", p.OutputTokens)
	}
}

func TestParseClaudeStreamJSON_NoResult_FallbackToAssistantText(t *testing.T) {
	stdout := `{"type":"assistant","session_id":"s2","message":{"content":[{"type":"text","text":"hi there"}]}}`

	p := claude.ParseClaudeStreamJSON(stdout)
	if p.Summary != "hi there" {
		t.Errorf("got summary %q, want %q", p.Summary, "hi there")
	}
}

func TestDetectAuthRequired(t *testing.T) {
	required, _ := claude.DetectAuthRequired("", "please log in to continue")
	if !required {
		t.Error("expected auth required")
	}
	required2, _ := claude.DetectAuthRequired("normal output", "")
	if required2 {
		t.Error("expected no auth required for normal output")
	}
}

func TestIsMaxTurnsResult(t *testing.T) {
	if !claude.IsMaxTurnsResult(map[string]interface{}{"subtype": "error_max_turns"}) {
		t.Error("expected max turns for error_max_turns subtype")
	}
	if claude.IsMaxTurnsResult(map[string]interface{}{"subtype": "success"}) {
		t.Error("expected no max turns for success subtype")
	}
}
