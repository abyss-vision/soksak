package adapters

import (
	"testing"

	processadapter "soksak/internal/adapters/process"
	"soksak/pkg/adapter"
)

func TestRegisterAll(t *testing.T) {
	registry := adapter.NewRegistry()
	manager := processadapter.New()

	RegisterAll(registry, manager)

	names := registry.List()
	if len(names) == 0 {
		t.Fatal("RegisterAll: expected at least one adapter, got 0")
	}

	// Verify some known adapters are registered.
	known := []string{"claude_local", "codex_local", "cursor", "gemini_local", "opencode_local", "pi_local", "openclaw_gateway", "http"}
	for _, name := range known {
		if _, ok := registry.Get(name); !ok {
			t.Errorf("RegisterAll: adapter %q not registered", name)
		}
	}
}
