package routes

import (
	"net/http/httptest"
	"testing"
)

func TestParseDateParam_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/?from=", nil)
	result, appErr := parseDateParam(req, "from")
	if appErr != nil {
		t.Errorf("parseDateParam empty: unexpected error: %v", appErr)
	}
	if result != nil {
		t.Error("parseDateParam empty: expected nil, got non-nil")
	}
}

func TestParseDateParam_Valid(t *testing.T) {
	req := httptest.NewRequest("GET", "/?from=2024-01-15T00:00:00Z", nil)
	result, appErr := parseDateParam(req, "from")
	if appErr != nil {
		t.Errorf("parseDateParam valid: unexpected error: %v", appErr)
	}
	if result == nil {
		t.Fatal("parseDateParam valid: expected non-nil, got nil")
	}
	if result.Year() != 2024 {
		t.Errorf("year = %d, want 2024", result.Year())
	}
}

func TestParseDateParam_Invalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/?from=not-a-date", nil)
	_, appErr := parseDateParam(req, "from")
	if appErr == nil {
		t.Fatal("parseDateParam invalid: expected error, got nil")
	}
}

func TestParseLimitParam_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	n, appErr := parseLimitParam(req)
	if appErr != nil {
		t.Errorf("parseLimitParam empty: unexpected error: %v", appErr)
	}
	if n != 100 {
		t.Errorf("limit = %d, want 100", n)
	}
}

func TestParseLimitParam_Valid(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=50", nil)
	n, appErr := parseLimitParam(req)
	if appErr != nil {
		t.Errorf("parseLimitParam valid: unexpected error: %v", appErr)
	}
	if n != 50 {
		t.Errorf("limit = %d, want 50", n)
	}
}

func TestParseLimitParam_TooLarge(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=999", nil)
	_, appErr := parseLimitParam(req)
	if appErr == nil {
		t.Fatal("parseLimitParam too large: expected error, got nil")
	}
}

func TestParseLimitParam_Zero(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=0", nil)
	_, appErr := parseLimitParam(req)
	if appErr == nil {
		t.Fatal("parseLimitParam zero: expected error, got nil")
	}
}

func TestParseLimitParam_Negative(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=-5", nil)
	_, appErr := parseLimitParam(req)
	if appErr == nil {
		t.Fatal("parseLimitParam negative: expected error, got nil")
	}
}

func TestParseLimitParam_Invalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/?limit=abc", nil)
	_, appErr := parseLimitParam(req)
	if appErr == nil {
		t.Fatal("parseLimitParam invalid: expected error, got nil")
	}
}

func TestParseOffsetParam_Empty(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	n := parseOffsetParam(req)
	if n != 0 {
		t.Errorf("offset empty = %d, want 0", n)
	}
}

func TestParseOffsetParam_Valid(t *testing.T) {
	req := httptest.NewRequest("GET", "/?offset=20", nil)
	n := parseOffsetParam(req)
	if n != 20 {
		t.Errorf("offset = %d, want 20", n)
	}
}

func TestParseOffsetParam_Negative(t *testing.T) {
	req := httptest.NewRequest("GET", "/?offset=-5", nil)
	n := parseOffsetParam(req)
	if n != 0 {
		t.Errorf("offset negative = %d, want 0 (fallback)", n)
	}
}

func TestParseOffsetParam_Invalid(t *testing.T) {
	req := httptest.NewRequest("GET", "/?offset=abc", nil)
	n := parseOffsetParam(req)
	if n != 0 {
		t.Errorf("offset invalid = %d, want 0 (fallback)", n)
	}
}
