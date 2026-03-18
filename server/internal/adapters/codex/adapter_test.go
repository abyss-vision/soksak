package codex_test

import (
	"testing"

	"soksak/internal/adapters/codex"
	"soksak/pkg/adapter"
)

func TestCodexAdapter_Name(t *testing.T) {
	if codex.New().Name() != "codex_local" {
		t.Error("wrong name")
	}
}

func TestCodexAdapter_BuildCommand(t *testing.T) {
	spec, err := codex.New().BuildCommand(adapter.AdapterConfig{Model: "o4-mini", WorkDir: "/repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Command != "codex" {
		t.Errorf("got command %q, want codex", spec.Command)
	}
	found := false
	for i, arg := range spec.Args {
		if arg == "--model" && i+1 < len(spec.Args) && spec.Args[i+1] == "o4-mini" {
			found = true
		}
	}
	if !found {
		t.Errorf("model not in args: %v", spec.Args)
	}
}

func TestCodexAdapter_ParseOutput_ItemCompleted(t *testing.T) {
	line := []byte(`{"type":"item.completed","item":{"type":"agent_message","text":"result"}}`)
	ev, err := codex.New().ParseOutput(line)
	if err != nil || ev == nil {
		t.Fatalf("err=%v ev=%v", err, ev)
	}
	if ev.Content != "result" {
		t.Errorf("got %q want %q", ev.Content, "result")
	}
}

func TestCodexAdapter_ParseOutput_Empty(t *testing.T) {
	ev, err := codex.New().ParseOutput([]byte(""))
	if err != nil || ev != nil {
		t.Errorf("expected nil event for empty line")
	}
}

func TestParseCodexJSONL(t *testing.T) {
	stdout := `{"type":"thread.started","thread_id":"t1"}
{"type":"item.completed","item":{"type":"agent_message","text":"Answer"}}
{"type":"turn.completed","usage":{"input_tokens":20,"output_tokens":10}}`

	p := codex.ParseCodexJSONL(stdout)
	if p.SessionID != "t1" {
		t.Errorf("got session %q, want t1", p.SessionID)
	}
	if p.Summary != "Answer" {
		t.Errorf("got summary %q, want Answer", p.Summary)
	}
	if p.InputTokens != 20 {
		t.Errorf("got input_tokens %d, want 20", p.InputTokens)
	}
}
