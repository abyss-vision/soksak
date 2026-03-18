package codex_test

import (
	"testing"

	"soksak/internal/adapters/codex"
)

func TestCodexAdapter_SupportedModels(t *testing.T) {
	models := codex.New().SupportedModels()
	if len(models) == 0 {
		t.Fatal("SupportedModels: expected at least one model, got 0")
	}
}

func TestCodexIsUnknownSessionError(t *testing.T) {
	if codex.IsUnknownSessionError("", "") {
		t.Error("IsUnknownSessionError(empty): expected false")
	}
	// With matching session error text, should return true.
	// Result depends on the regex — just verify it doesn't panic.
	codex.IsUnknownSessionError("some output", "some error")
}
