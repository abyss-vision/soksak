package openclaw_test

import (
	"testing"

	"abyss-view/internal/adapters/openclaw"
)

func TestOpenClawAdapter_SupportedModels(t *testing.T) {
	// openclaw returns empty (server-side models) — verify no panic.
	models := openclaw.New().SupportedModels()
	if models == nil {
		t.Error("SupportedModels returned nil, want empty slice")
	}
}
