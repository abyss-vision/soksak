package cursor_test

import (
	"testing"

	"abyss-view/internal/adapters/cursor"
	"abyss-view/pkg/adapter"
)

func TestCursorAdapter_Name(t *testing.T) {
	if cursor.New().Name() != "cursor" {
		t.Error("wrong name")
	}
}

func TestCursorAdapter_BuildCommand(t *testing.T) {
	spec, err := cursor.New().BuildCommand(adapter.AdapterConfig{Model: "gpt-4o"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Command != "cursor" {
		t.Errorf("got command %q, want cursor", spec.Command)
	}
	hasModel := false
	for i, a := range spec.Args {
		if a == "--model" && i+1 < len(spec.Args) && spec.Args[i+1] == "gpt-4o" {
			hasModel = true
		}
	}
	if !hasModel {
		t.Errorf("model not in args: %v", spec.Args)
	}
}

func TestCursorAdapter_ParseOutput_Assistant(t *testing.T) {
	line := []byte(`{"type":"assistant","message":{"content":[{"type":"text","text":"hi"}]}}`)
	ev, err := cursor.New().ParseOutput(line)
	if err != nil || ev == nil {
		t.Fatalf("err=%v ev=%v", err, ev)
	}
	if ev.Content != "hi" {
		t.Errorf("got content %q want hi", ev.Content)
	}
}

func TestParseCursorJSONL(t *testing.T) {
	stdout := `{"type":"assistant","session_id":"c1","message":{"content":[{"type":"text","text":"Response"}]}}
{"type":"result","session_id":"c1","usage":{"input_tokens":12,"output_tokens":6}}`

	p := cursor.ParseCursorJSONL(stdout)
	if p.SessionID != "c1" {
		t.Errorf("got session %q want c1", p.SessionID)
	}
	if p.Summary != "Response" {
		t.Errorf("got summary %q want Response", p.Summary)
	}
	if p.InputTokens != 12 {
		t.Errorf("got input_tokens %d want 12", p.InputTokens)
	}
}
