package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDFromContext_Empty(t *testing.T) {
	id := RequestIDFromContext(context.Background())
	if id != "" {
		t.Errorf("RequestIDFromContext empty: expected empty, got %q", id)
	}
}

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()
	if id1 == "" {
		t.Error("generateRequestID returned empty string")
	}
	if len(id1) != 16 { // 8 bytes as hex = 16 chars
		t.Errorf("generateRequestID length = %d, want 16", len(id1))
	}
	if id1 == id2 {
		t.Error("generateRequestID returned same value twice (unlikely unless broken)")
	}
}

func TestRequestLogger(t *testing.T) {
	var capturedID string
	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if capturedID == "" {
		t.Error("RequestIDFromContext inside handler: expected non-empty, got empty")
	}
	if rr.Header().Get("X-Request-ID") != capturedID {
		t.Errorf("X-Request-ID header = %q, want %q", rr.Header().Get("X-Request-ID"), capturedID)
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, status: http.StatusOK}

	rw.WriteHeader(http.StatusCreated)
	if rw.status != http.StatusCreated {
		t.Errorf("status = %d, want %d", rw.status, http.StatusCreated)
	}
	// Second call should not change status (wrote=true).
	rw.WriteHeader(http.StatusNotFound)
	if rw.status != http.StatusCreated {
		t.Errorf("second WriteHeader changed status to %d, should remain %d", rw.status, http.StatusCreated)
	}
}

func TestResponseWriter_Write_SetsStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: rec, status: http.StatusOK}

	_, _ = rw.Write([]byte("hello"))
	if rw.status != http.StatusOK {
		t.Errorf("Write did not set status correctly: got %d", rw.status)
	}
	if !rw.wrote {
		t.Error("wrote should be true after Write")
	}
}

func TestRequestLogger_StatusCaptured(t *testing.T) {
	handler := RequestLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodDelete, "/resource", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", rr.Code)
	}
}
