package opencode_test

import (
	"testing"

	"soksak/internal/adapters/opencode"
)

func TestOpenCodeAdapter_SupportedModels(t *testing.T) {
	models := opencode.New().SupportedModels()
	if len(models) == 0 {
		t.Fatal("SupportedModels: expected at least one model, got 0")
	}
}

func TestOpenCodeIsUnknownSessionError(t *testing.T) {
	if opencode.IsUnknownSessionError("", "") {
		t.Error("IsUnknownSessionError(empty): expected false")
	}
}
