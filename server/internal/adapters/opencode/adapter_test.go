package opencode_test

import (
	"testing"

	"abyss-view/internal/adapters/opencode"
	"abyss-view/pkg/adapter"
)

func TestOpenCodeAdapter_Name(t *testing.T) {
	if opencode.New().Name() != "opencode_local" {
		t.Error("wrong name")
	}
}

func TestOpenCodeAdapter_BuildCommand(t *testing.T) {
	spec, err := opencode.New().BuildCommand(adapter.AdapterConfig{Model: "anthropic/claude-sonnet-4-5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Command != "opencode" {
		t.Errorf("got command %q, want opencode", spec.Command)
	}
}

func TestParseOpenCodeJSONL(t *testing.T) {
	stdout := `{"type":"text","sessionID":"oc1","part":{"text":"Response"}}
{"type":"step_finish","sessionID":"oc1","part":{"tokens":{"input":5,"output":3},"cost":0.01}}`

	p := opencode.ParseOpenCodeJSONL(stdout)
	if p.SessionID != "oc1" {
		t.Errorf("got session %q want oc1", p.SessionID)
	}
	if p.Summary != "Response" {
		t.Errorf("got summary %q want Response", p.Summary)
	}
	if p.InputTokens != 5 {
		t.Errorf("got input_tokens %d want 5", p.InputTokens)
	}
}
