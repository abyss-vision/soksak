package claude_test

import (
	"testing"

	"soksak/internal/adapters/claude"
	"soksak/pkg/adapter"
)

func TestClaudeAdapter_Name(t *testing.T) {
	a := claude.New()
	if a.Name() != "claude_local" {
		t.Errorf("got %q, want %q", a.Name(), "claude_local")
	}
}

func TestClaudeAdapter_BuildCommand_Defaults(t *testing.T) {
	a := claude.New()
	spec, err := a.BuildCommand(adapter.AdapterConfig{WorkDir: "/tmp"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Command != "claude" {
		t.Errorf("got command %q, want %q", spec.Command, "claude")
	}
	// Must include stream-json args
	found := false
	for i, arg := range spec.Args {
		if arg == "--output-format" && i+1 < len(spec.Args) && spec.Args[i+1] == "stream-json" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected --output-format stream-json in args: %v", spec.Args)
	}
}

func TestClaudeAdapter_BuildCommand_WithModel(t *testing.T) {
	a := claude.New()
	spec, err := a.BuildCommand(adapter.AdapterConfig{Model: "claude-opus-4-5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for i, arg := range spec.Args {
		if arg == "--model" && i+1 < len(spec.Args) && spec.Args[i+1] == "claude-opus-4-5" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected --model claude-opus-4-5 in args: %v", spec.Args)
	}
}

func TestClaudeAdapter_ParseOutput_StreamJSON(t *testing.T) {
	a := claude.New()
	line := []byte(`{"type":"assistant","session_id":"s1","message":{"content":[{"type":"text","text":"hello"}]}}`)
	ev, err := a.ParseOutput(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev == nil {
		t.Fatal("expected non-nil event")
	}
	if ev.Type != "assistant" {
		t.Errorf("got type %q, want %q", ev.Type, "assistant")
	}
	if ev.Content != "hello" {
		t.Errorf("got content %q, want %q", ev.Content, "hello")
	}
}

func TestClaudeAdapter_ParseOutput_Empty(t *testing.T) {
	a := claude.New()
	ev, err := a.ParseOutput([]byte("   "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev != nil {
		t.Errorf("expected nil event for blank line, got %+v", ev)
	}
}

func TestClaudeAdapter_BuildCommand_KoreanLanguageDirective(t *testing.T) {
	a := claude.New()
	cfg := adapter.AdapterConfig{
		WorkDir:               "/tmp",
		Prompt:                "Do something",
		CommunicationLanguage: "ko",
	}
	spec, err := a.BuildCommand(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "[IMPORTANT] You MUST respond in Korean (한국어). All your output must be in Korean (한국어).\n\nDo something"
	if spec.Stdin != want {
		t.Errorf("got stdin %q, want %q", spec.Stdin, want)
	}
}

func TestClaudeAdapter_BuildCommand_NoLanguageDirectiveWhenEmpty(t *testing.T) {
	a := claude.New()
	cfg := adapter.AdapterConfig{
		WorkDir: "/tmp",
		Prompt:  "Do something",
	}
	spec, err := a.BuildCommand(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Stdin != "Do something" {
		t.Errorf("got stdin %q, want %q", spec.Stdin, "Do something")
	}
}

func TestClaudeAdapter_BuildCommand_EnglishLanguageDirective(t *testing.T) {
	a := claude.New()
	cfg := adapter.AdapterConfig{
		WorkDir:               "/tmp",
		Prompt:                "Hello",
		CommunicationLanguage: "en",
	}
	spec, err := a.BuildCommand(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "[IMPORTANT] You MUST respond in English. All your output must be in English.\n\nHello"
	if spec.Stdin != want {
		t.Errorf("got stdin %q, want %q", spec.Stdin, want)
	}
}

func TestClaudeAdapter_BuildCommand_JapaneseLanguageDirective(t *testing.T) {
	a := claude.New()
	cfg := adapter.AdapterConfig{
		WorkDir:               "/tmp",
		Prompt:                "Konnichiwa",
		CommunicationLanguage: "ja",
	}
	spec, err := a.BuildCommand(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "[IMPORTANT] You MUST respond in Japanese (日本語). All your output must be in Japanese (日本語).\n\nKonnichiwa"
	if spec.Stdin != want {
		t.Errorf("got stdin %q, want %q", spec.Stdin, want)
	}
}
