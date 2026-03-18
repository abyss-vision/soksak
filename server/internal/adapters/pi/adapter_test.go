package pi_test

import (
	"testing"

	"abyss-view/internal/adapters/pi"
	"abyss-view/pkg/adapter"
)

func TestPiAdapter_Name(t *testing.T) {
	if pi.New().Name() != "pi_local" {
		t.Error("wrong name")
	}
}

func TestPiAdapter_BuildCommand(t *testing.T) {
	spec, err := pi.New().BuildCommand(adapter.AdapterConfig{WorkDir: "/home"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Command != "pi" {
		t.Errorf("got command %q, want pi", spec.Command)
	}
}

func TestParsePiJSONL(t *testing.T) {
	stdout := `{"type":"turn_end","message":{"content":"Final answer","usage":{"input":8,"output":4,"cost":{"total":0.005}}}}`

	p := pi.ParsePiJSONL(stdout)
	if p.Summary != "Final answer" {
		t.Errorf("got summary %q want 'Final answer'", p.Summary)
	}
	if p.InputTokens != 8 {
		t.Errorf("got input_tokens %d want 8", p.InputTokens)
	}
}

func TestPiAdapter_ParseOutput_SkipsInternalFrames(t *testing.T) {
	ev, err := pi.New().ParseOutput([]byte(`{"type":"response","id":"x"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev != nil {
		t.Error("expected nil event for internal protocol frame")
	}
}
