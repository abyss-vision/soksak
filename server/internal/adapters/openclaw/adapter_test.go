package openclaw_test

import (
	"testing"

	"abyss-view/internal/adapters/openclaw"
	"abyss-view/pkg/adapter"
)

func TestOpenClawAdapter_Name(t *testing.T) {
	if openclaw.New().Name() != "openclaw_gateway" {
		t.Error("wrong name")
	}
}

func TestOpenClawAdapter_BuildCommandErrors(t *testing.T) {
	_, err := openclaw.New().BuildCommand(adapter.AdapterConfig{})
	if err == nil {
		t.Error("expected error: openclaw is not a process adapter")
	}
}

func TestOpenClawAdapter_BuildRequestPayload_IncludesLanguage(t *testing.T) {
	a := openclaw.New()
	cfg := adapter.AdapterConfig{
		Prompt:                "run tests",
		Model:                 "claude-3-7-sonnet",
		CommunicationLanguage: "ja",
	}
	body := a.BuildRequestPayload(cfg)
	lang, ok := body["language"]
	if !ok {
		t.Fatal("expected 'language' field in request payload")
	}
	if lang != "ja" {
		t.Errorf("got language %q, want %q", lang, "ja")
	}
}

func TestOpenClawAdapter_BuildRequestPayload_NoLanguageWhenEmpty(t *testing.T) {
	a := openclaw.New()
	cfg := adapter.AdapterConfig{Prompt: "hello"}
	body := a.BuildRequestPayload(cfg)
	if _, ok := body["language"]; ok {
		t.Error("language field should not be present when CommunicationLanguage is empty")
	}
}

func TestOpenClawAdapter_ParseOutput_SSE(t *testing.T) {
	a := openclaw.New()

	ev, err := a.ParseOutput([]byte(`data: {"type":"message","content":"hello"}`))
	if err != nil || ev == nil {
		t.Fatalf("err=%v ev=%v", err, ev)
	}
	if ev.Content != "hello" {
		t.Errorf("got content %q want hello", ev.Content)
	}

	done, _ := a.ParseOutput([]byte("data: [DONE]"))
	if done == nil || done.Type != "done" {
		t.Errorf("expected done event, got %+v", done)
	}

	empty, _ := a.ParseOutput([]byte(""))
	if empty != nil {
		t.Error("expected nil for empty line")
	}
}
