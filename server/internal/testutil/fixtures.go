package testutil

import (
	"os"
	"path/filepath"
	"testing"
)

// LoadFixture reads a fixture file from the testdata/ directory relative to the caller's test file.
// The path argument is relative to the testdata/ directory (e.g., "users/valid.json").
func LoadFixture(t *testing.T, path string) []byte {
	t.Helper()

	fullPath := filepath.Join("testdata", path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("could not load fixture %q: %v", fullPath, err)
	}
	return data
}
