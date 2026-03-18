package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testInput struct {
	Name  string `json:"name" validate:"required"`
	Count int    `json:"count" validate:"min=1"`
}

func TestBindAndValidate_Success(t *testing.T) {
	body := `{"name":"Alice","count":3}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	result, err := BindAndValidate[testInput](req)
	if err != nil {
		t.Fatalf("BindAndValidate: unexpected error: %v", err)
	}
	if result.Name != "Alice" {
		t.Errorf("Name = %q, want %q", result.Name, "Alice")
	}
	if result.Count != 3 {
		t.Errorf("Count = %d, want 3", result.Count)
	}
}

func TestBindAndValidate_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")

	_, err := BindAndValidate[testInput](req)
	if err == nil {
		t.Fatal("BindAndValidate invalid JSON: expected error, got nil")
	}
	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Code != "VALIDATION_ERROR" {
		t.Errorf("Code = %q, want VALIDATION_ERROR", appErr.Code)
	}
	if appErr.Status != http.StatusUnprocessableEntity {
		t.Errorf("Status = %d, want 422", appErr.Status)
	}
}

func TestBindAndValidate_ValidationFails(t *testing.T) {
	body := `{"name":"","count":0}` // both fields fail validation
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	_, err := BindAndValidate[testInput](req)
	if err == nil {
		t.Fatal("BindAndValidate invalid struct: expected error, got nil")
	}
	appErr, ok := err.(*AppError)
	if !ok {
		t.Fatalf("expected *AppError, got %T", err)
	}
	if appErr.Status != http.StatusUnprocessableEntity {
		t.Errorf("Status = %d, want 422", appErr.Status)
	}
}

func TestBindAndValidate_UnknownField(t *testing.T) {
	body := `{"name":"Alice","count":1,"unknown_field":"value"}`
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// DisallowUnknownFields should cause an error.
	_, err := BindAndValidate[testInput](req)
	if err == nil {
		t.Fatal("BindAndValidate unknown field: expected error, got nil")
	}
}
