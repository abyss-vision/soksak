package http_test

import (
	"testing"

	httpadapter "abyss-view/internal/adapters/http"
	"abyss-view/pkg/adapter"
)

func TestHTTPAdapter_Name(t *testing.T) {
	if httpadapter.New().Name() != "http" {
		t.Error("wrong name")
	}
}

func TestHTTPAdapter_BuildCommandErrors(t *testing.T) {
	_, err := httpadapter.New().BuildCommand(adapter.AdapterConfig{})
	if err == nil {
		t.Error("expected error: http is not a process adapter")
	}
}

func TestHTTPAdapter_ParseOutput_JSON(t *testing.T) {
	a := httpadapter.New()
	ev, err := a.ParseOutput([]byte(`{"type":"text","content":"webhook response"}`))
	if err != nil || ev == nil {
		t.Fatalf("err=%v ev=%v", err, ev)
	}
	if ev.Type != "text" {
		t.Errorf("got type %q want text", ev.Type)
	}
	if ev.Content != "webhook response" {
		t.Errorf("got content %q want 'webhook response'", ev.Content)
	}
}

func TestHTTPAdapter_ParseOutput_PlainText(t *testing.T) {
	ev, err := httpadapter.New().ParseOutput([]byte("plain text response"))
	if err != nil || ev == nil {
		t.Fatalf("err=%v ev=%v", err, ev)
	}
	if ev.Type != "text" {
		t.Errorf("got type %q want text", ev.Type)
	}
}

func TestHTTPAdapter_ParseOutput_DONE(t *testing.T) {
	ev, _ := httpadapter.New().ParseOutput([]byte("data: [DONE]"))
	if ev == nil || ev.Type != "done" {
		t.Errorf("expected done event, got %+v", ev)
	}
}

func TestHTTPAdapter_ParseOutput_Empty(t *testing.T) {
	ev, _ := httpadapter.New().ParseOutput([]byte(""))
	if ev != nil {
		t.Error("expected nil for empty line")
	}
}

func TestHTTPAdapter_BuildRequestPayload_IncludesLanguage(t *testing.T) {
	a := httpadapter.New()
	cfg := adapter.AdapterConfig{
		Prompt:                "test prompt",
		Model:                 "gpt-4",
		CommunicationLanguage: "ko",
	}
	body := a.BuildRequestPayload(cfg)
	lang, ok := body["language"]
	if !ok {
		t.Fatal("expected 'language' field in request payload")
	}
	if lang != "ko" {
		t.Errorf("got language %q, want %q", lang, "ko")
	}
}

func TestHTTPAdapter_BuildRequestPayload_NoLanguageWhenEmpty(t *testing.T) {
	a := httpadapter.New()
	cfg := adapter.AdapterConfig{
		Prompt: "test prompt",
		Model:  "gpt-4",
	}
	body := a.BuildRequestPayload(cfg)
	if _, ok := body["language"]; ok {
		t.Error("language field should not be present when CommunicationLanguage is empty")
	}
}
