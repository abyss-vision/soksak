package gemini_test

import (
	"testing"

	"soksak/internal/adapters/gemini"
)

func TestGeminiAdapter_SupportedModels(t *testing.T) {
	models := gemini.New().SupportedModels()
	if len(models) == 0 {
		t.Fatal("SupportedModels: expected at least one model, got 0")
	}
}

func TestGeminiIsUnknownSessionError(t *testing.T) {
	if gemini.IsUnknownSessionError("", "") {
		t.Error("IsUnknownSessionError(empty): expected false")
	}
	if !gemini.IsUnknownSessionError("cannot resume session", "") {
		t.Error("IsUnknownSessionError matching: expected true")
	}
}
