package claude_test

import (
	"testing"

	"soksak/internal/adapters/claude"
)

func TestClaudeAdapter_SupportedModels(t *testing.T) {
	a := claude.New()
	models := a.SupportedModels()
	if len(models) == 0 {
		t.Fatal("SupportedModels: expected at least one model, got 0")
	}
	for _, m := range models {
		if m.ID == "" {
			t.Error("model ID is empty")
		}
		if m.Provider == "" {
			t.Error("model provider is empty")
		}
	}
}

func TestIsUnknownSessionError(t *testing.T) {
	// Should return false for nil.
	if claude.IsUnknownSessionError(nil) {
		t.Error("IsUnknownSessionError(nil): expected false")
	}

	// Should return false for unrelated errors.
	if claude.IsUnknownSessionError(map[string]interface{}{"errors": []interface{}{"some other error"}}) {
		t.Error("IsUnknownSessionError unrelated: expected false")
	}

	// Should return true when result contains session error.
	m := map[string]interface{}{
		"result": "no conversation found with session id abc123",
	}
	if !claude.IsUnknownSessionError(m) {
		t.Error("IsUnknownSessionError session result: expected true")
	}

	// Should return true when errors array contains session error.
	m2 := map[string]interface{}{
		"errors": []interface{}{"unknown session: abc123"},
	}
	if !claude.IsUnknownSessionError(m2) {
		t.Error("IsUnknownSessionError errors array: expected true")
	}

	// Should handle map error entries.
	m3 := map[string]interface{}{
		"errors": []interface{}{
			map[string]interface{}{"message": "no conversation found with session id xyz"},
		},
	}
	if !claude.IsUnknownSessionError(m3) {
		t.Error("IsUnknownSessionError map error entry: expected true")
	}

	// Empty result + no matching errors → false.
	m4 := map[string]interface{}{"errors": []interface{}{}}
	if claude.IsUnknownSessionError(m4) {
		t.Error("IsUnknownSessionError empty errors: expected false")
	}
}
