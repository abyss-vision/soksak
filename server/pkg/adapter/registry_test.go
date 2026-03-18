package adapter_test

import (
	"testing"

	"abyss-view/pkg/adapter"
)

// stubAdapter is a minimal ServerAdapter for test use.
type stubAdapter struct {
	name string
}

func (s *stubAdapter) Name() string                                      { return s.name }
func (s *stubAdapter) BuildCommand(_ adapter.AdapterConfig) (*adapter.CommandSpec, error) {
	return &adapter.CommandSpec{Command: "echo"}, nil
}
func (s *stubAdapter) ParseOutput(_ []byte) (*adapter.OutputEvent, error) { return nil, nil }
func (s *stubAdapter) SupportedModels() []adapter.ModelInfo              { return nil }

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := adapter.NewRegistry()
	a := &stubAdapter{name: "test_adapter"}
	r.Register(a)

	got, ok := r.Get("test_adapter")
	if !ok {
		t.Fatal("expected adapter to be found after registration")
	}
	if got.Name() != "test_adapter" {
		t.Errorf("got name %q, want %q", got.Name(), "test_adapter")
	}
}

func TestRegistry_GetMissing(t *testing.T) {
	r := adapter.NewRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected Get to return false for unknown adapter")
	}
}

func TestRegistry_List(t *testing.T) {
	r := adapter.NewRegistry()
	r.Register(&stubAdapter{name: "a"})
	r.Register(&stubAdapter{name: "b"})

	names := r.List()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
	seen := map[string]bool{}
	for _, n := range names {
		seen[n] = true
	}
	if !seen["a"] || !seen["b"] {
		t.Errorf("list missing expected names: %v", names)
	}
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	r := adapter.NewRegistry()
	r.Register(&stubAdapter{name: "dup"})

	defer func() {
		if rec := recover(); rec == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	r.Register(&stubAdapter{name: "dup"})
}
