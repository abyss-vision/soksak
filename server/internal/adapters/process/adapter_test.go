package process_test

import (
	"testing"

	"abyss-view/internal/adapters/process"
	"abyss-view/pkg/adapter"
)

func TestProcessAdapter_Name(t *testing.T) {
	m := process.New()
	a := process.NewProcessAdapter(m)
	if a.Name() != "process" {
		t.Errorf("got %q, want %q", a.Name(), "process")
	}
}

func TestProcessAdapter_BuildCommand_Basic(t *testing.T) {
	m := process.New()
	a := process.NewProcessAdapter(m)

	spec, err := a.BuildCommand(adapter.AdapterConfig{
		AdapterType: "echo",
		WorkDir:     "/tmp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Command != "echo" {
		t.Errorf("got command %q, want %q", spec.Command, "echo")
	}
	if spec.WorkDir != "/tmp" {
		t.Errorf("got workdir %q, want %q", spec.WorkDir, "/tmp")
	}
}

func TestProcessAdapter_BuildCommand_Override(t *testing.T) {
	m := process.New()
	a := process.NewProcessAdapter(m)

	spec, err := a.BuildCommand(adapter.AdapterConfig{
		AdapterType: "ignored",
		ExtraArgs:   map[string]string{"command": "mybin"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if spec.Command != "mybin" {
		t.Errorf("got command %q, want %q", spec.Command, "mybin")
	}
}

func TestProcessAdapter_BuildCommand_MissingCommand(t *testing.T) {
	m := process.New()
	a := process.NewProcessAdapter(m)

	_, err := a.BuildCommand(adapter.AdapterConfig{})
	if err == nil {
		t.Error("expected error when command is empty")
	}
}

func TestProcessAdapter_BuildCommand_EnvVars(t *testing.T) {
	m := process.New()
	a := process.NewProcessAdapter(m)

	spec, err := a.BuildCommand(adapter.AdapterConfig{
		AdapterType: "mybin",
		EnvVars:     map[string]string{"FOO": "bar", "BAZ": "qux"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spec.Env) != 2 {
		t.Errorf("expected 2 env entries, got %d: %v", len(spec.Env), spec.Env)
	}
}

func TestProcessAdapter_ParseOutput_PlainText(t *testing.T) {
	m := process.New()
	a := process.NewProcessAdapter(m)

	ev, err := a.ParseOutput([]byte("hello world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev == nil {
		t.Fatal("expected non-nil event")
	}
	if ev.Type != "text" {
		t.Errorf("got type %q, want %q", ev.Type, "text")
	}
	if ev.Content != "hello world" {
		t.Errorf("got content %q, want %q", ev.Content, "hello world")
	}
}

func TestProcessAdapter_ParseOutput_JSON(t *testing.T) {
	m := process.New()
	a := process.NewProcessAdapter(m)

	ev, err := a.ParseOutput([]byte(`{"type":"tool_use","content":"doing stuff"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev == nil {
		t.Fatal("expected non-nil event")
	}
	if ev.Type != "tool_use" {
		t.Errorf("got type %q, want %q", ev.Type, "tool_use")
	}
	if ev.Content != "doing stuff" {
		t.Errorf("got content %q, want %q", ev.Content, "doing stuff")
	}
}

func TestProcessAdapter_ParseOutput_EmptyLine(t *testing.T) {
	m := process.New()
	a := process.NewProcessAdapter(m)

	ev, err := a.ParseOutput([]byte("   "))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ev != nil {
		t.Errorf("expected nil event for blank line, got %+v", ev)
	}
}

func TestProcessAdapter_SupportedModels(t *testing.T) {
	m := process.New()
	a := process.NewProcessAdapter(m)

	models := a.SupportedModels()
	if models == nil {
		t.Error("SupportedModels must return non-nil slice")
	}
	if len(models) != 0 {
		t.Errorf("expected 0 models for process adapter, got %d", len(models))
	}
}
