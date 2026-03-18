package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	e := &AppError{Code: "TEST", Status: 400, Message: "test message"}
	if e.Error() != "test message" {
		t.Errorf("Error() = %q, want %q", e.Error(), "test message")
	}
}

func TestErrConstructors(t *testing.T) {
	cases := []struct {
		name       string
		fn         func(string) *AppError
		wantCode   string
		wantStatus int
	}{
		{"ErrConflict", ErrConflict, "CONFLICT", http.StatusConflict},
		{"ErrForbidden", ErrForbidden, "FORBIDDEN", http.StatusForbidden},
		{"ErrNotFound", ErrNotFound, "NOT_FOUND", http.StatusNotFound},
		{"ErrUnprocessable", ErrUnprocessable, "UNPROCESSABLE", http.StatusUnprocessableEntity},
		{"ErrMethodNotAllowed", ErrMethodNotAllowed, "METHOD_NOT_ALLOWED", http.StatusMethodNotAllowed},
		{"ErrInternal", ErrInternal, "INTERNAL_ERROR", http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			e := tc.fn("msg")
			if e.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", e.Code, tc.wantCode)
			}
			if e.Status != tc.wantStatus {
				t.Errorf("Status = %d, want %d", e.Status, tc.wantStatus)
			}
			if e.Message != "msg" {
				t.Errorf("Message = %q, want %q", e.Message, "msg")
			}
		})
	}
}

func TestErrorHandler_AppErrorPanic(t *testing.T) {
	handler := ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(ErrNotFound("item not found"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	var resp map[string]interface{}
	_ = json.NewDecoder(rr.Body).Decode(&resp)
	if resp["code"] != "NOT_FOUND" {
		t.Errorf("code = %v, want NOT_FOUND", resp["code"])
	}
}

func TestErrorHandler_ErrorPanic(t *testing.T) {
	handler := ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrNoCookie) // a generic error
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestErrorHandler_NonErrorPanic(t *testing.T) {
	handler := ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("unexpected string panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusInternalServerError)
	}
}

func TestErrorHandler_NoPanic(t *testing.T) {
	handler := ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestWriteAppError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	WriteAppError(rr, req, ErrForbidden("access denied"))

	if rr.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rr.Code, http.StatusForbidden)
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}
