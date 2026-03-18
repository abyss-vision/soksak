package gemini_test

import (
	"testing"

	"soksak/internal/adapters/gemini"
	"soksak/pkg/adapter"
)

func TestGeminiAdapter_Name(t *testing.T) {
	if gemini.New().Name() != "gemini_local" {
		t.Error("wrong name")
	}
}

func TestGeminiAdapter_BuildCommand(t *testing.T) {
	spec, err := gemini.New().BuildCommand(adapter.AdapterConfig{Model: "gemini-2.5-pro"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Command != "gemini" {
		t.Errorf("got command %q, want gemini", spec.Command)
	}
	hasModel := false
	for i, a := range spec.Args {
		if a == "--model" && i+1 < len(spec.Args) && spec.Args[i+1] == "gemini-2.5-pro" {
			hasModel = true
		}
	}
	if !hasModel {
		t.Errorf("model not in args: %v", spec.Args)
	}
}

func TestParseGeminiJSONL_WithAssistantAndResult(t *testing.T) {
	// When assistant events produce text, that text wins in the summary
	// (result text is only used as fallback when messages is empty).
	stdout := `{"type":"assistant","session_id":"g1","message":{"content":[{"type":"text","text":"Hello Gemini"}]}}
{"type":"result","session_id":"g1","result":"Done","usage":{"input_tokens":15,"output_tokens":8}}`

	p := gemini.ParseGeminiJSONL(stdout)
	if p.SessionID != "g1" {
		t.Errorf("got session %q want g1", p.SessionID)
	}
	// assistant text wins over result fallback
	if p.Summary != "Hello Gemini" {
		t.Errorf("got summary %q want 'Hello Gemini'", p.Summary)
	}
	if p.InputTokens != 15 {
		t.Errorf("got input_tokens %d want 15", p.InputTokens)
	}
}

func TestParseGeminiJSONL_ResultOnlyFallback(t *testing.T) {
	// When there are no assistant messages, result text is used.
	stdout := `{"type":"result","session_id":"g2","result":"FallbackSummary","usage":{"output_tokens":3}}`

	p := gemini.ParseGeminiJSONL(stdout)
	if p.Summary != "FallbackSummary" {
		t.Errorf("got summary %q want FallbackSummary", p.Summary)
	}
}

func TestDetectGeminiAuthRequired(t *testing.T) {
	if !gemini.DetectAuthRequired("", "authentication required") {
		t.Error("expected auth required")
	}
	if gemini.DetectAuthRequired("all good", "") {
		t.Error("expected no auth required")
	}
}
