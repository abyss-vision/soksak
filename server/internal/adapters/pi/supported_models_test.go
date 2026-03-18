package pi_test

import (
	"testing"

	"abyss-view/internal/adapters/pi"
)

func TestPiAdapter_SupportedModels(t *testing.T) {
	models := pi.New().SupportedModels()
	if len(models) == 0 {
		t.Fatal("SupportedModels: expected at least one model, got 0")
	}
}

func TestPiIsUnknownSessionError(t *testing.T) {
	if pi.IsUnknownSessionError("", "") {
		t.Error("IsUnknownSessionError(empty): expected false")
	}
}
