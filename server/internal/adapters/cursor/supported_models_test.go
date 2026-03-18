package cursor_test

import (
	"testing"

	"soksak/internal/adapters/cursor"
)

func TestCursorAdapter_SupportedModels(t *testing.T) {
	models := cursor.New().SupportedModels()
	if len(models) == 0 {
		t.Fatal("SupportedModels: expected at least one model, got 0")
	}
	for _, m := range models {
		if m.ID == "" {
			t.Error("model ID is empty")
		}
	}
}

func TestCursorIsUnknownSessionError(t *testing.T) {
	if cursor.IsUnknownSessionError("", "") {
		t.Error("IsUnknownSessionError(empty): expected false")
	}
	if !cursor.IsUnknownSessionError("could not resume session abc", "") {
		t.Error("IsUnknownSessionError matching text: expected true")
	}
}
