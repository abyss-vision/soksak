package auth

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestHashToken(t *testing.T) {
	raw := "my-session-token"
	want := fmt.Sprintf("%x", sha256.Sum256([]byte(raw)))
	got := hashToken(raw)
	if got != want {
		t.Errorf("hashToken(%q) = %q, want %q", raw, got, want)
	}
}

func TestHashToken_Empty(t *testing.T) {
	got := hashToken("")
	want := fmt.Sprintf("%x", sha256.Sum256([]byte("")))
	if got != want {
		t.Errorf("hashToken(\"\") = %q, want %q", got, want)
	}
}
