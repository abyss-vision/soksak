package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"

	apii18n "soksak/internal/i18n"
)

// AppError is a structured application error with HTTP status, a machine-readable code,
// a human-readable message, and optional details (e.g. field validation errors).
type AppError struct {
	Code    string `json:"code"`
	Status  int    `json:"-"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *AppError) Error() string { return e.Message }

// Standard constructor helpers.

func ErrConflict(msg string) *AppError {
	return &AppError{Code: "CONFLICT", Status: http.StatusConflict, Message: msg}
}

func ErrForbidden(msg string) *AppError {
	return &AppError{Code: "FORBIDDEN", Status: http.StatusForbidden, Message: msg}
}

func ErrNotFound(msg string) *AppError {
	return &AppError{Code: "NOT_FOUND", Status: http.StatusNotFound, Message: msg}
}

func ErrUnprocessable(msg string) *AppError {
	return &AppError{Code: "UNPROCESSABLE", Status: http.StatusUnprocessableEntity, Message: msg}
}

func ErrMethodNotAllowed(msg string) *AppError {
	return &AppError{Code: "METHOD_NOT_ALLOWED", Status: http.StatusMethodNotAllowed, Message: msg}
}

func ErrInternal(msg string) *AppError {
	return &AppError{Code: "INTERNAL_ERROR", Status: http.StatusInternalServerError, Message: msg}
}

type errorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details any    `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ErrorHandler is a Chi-compatible recovery+error middleware.
// It catches panics and converts AppErrors to structured JSON responses.
func ErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				var appErr *AppError
				switch v := rec.(type) {
				case *AppError:
					appErr = v
				case error:
					slog.Error("recovered panic", "err", v, "method", r.Method, "path", r.URL.Path)
					appErr = ErrInternal("Internal server error")
				default:
					slog.Error("recovered panic (non-error)", "value", v, "method", r.Method, "path", r.URL.Path)
					appErr = ErrInternal("Internal server error")
				}
				writeJSON(w, appErr.Status, errorResponse{
					Error:   localizeError(r, appErr),
					Code:    appErr.Code,
					Details: appErr.Details,
				})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// localizeError returns the error message, using the go-i18n localizer when available.
// Falls back to the AppError.Message if no translation is found.
func localizeError(r *http.Request, appErr *AppError) string {
	localizer := apii18n.LocalizerFromContext(r.Context())
	if localizer == nil {
		return appErr.Message
	}
	msg, err := localizer.Localize(&i18n.LocalizeConfig{MessageID: appErr.Code})
	if err != nil || msg == "" {
		return appErr.Message
	}
	return msg
}

// WriteAppError writes an AppError as a JSON response directly (for use inside handlers).
func WriteAppError(w http.ResponseWriter, r *http.Request, appErr *AppError) {
	writeJSON(w, appErr.Status, errorResponse{
		Error:   localizeError(r, appErr),
		Code:    appErr.Code,
		Details: appErr.Details,
	})
}
