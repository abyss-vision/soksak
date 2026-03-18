package i18n

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/text/language"
)

func TestNewBundle(t *testing.T) {
	bundle := NewBundle(language.English)
	if bundle == nil {
		t.Fatal("NewBundle returned nil")
	}
}

func TestGetLocalizer(t *testing.T) {
	bundle := NewBundle(language.English)
	loc := GetLocalizer(bundle, "en")
	if loc == nil {
		t.Fatal("GetLocalizer returned nil")
	}
}

func TestGetLocalizer_Unknown(t *testing.T) {
	bundle := NewBundle(language.English)
	// Unknown language falls back without error.
	loc := GetLocalizer(bundle, "xx-ZZ")
	if loc == nil {
		t.Fatal("GetLocalizer unknown language returned nil")
	}
}

func TestLocaleMiddleware(t *testing.T) {
	bundle := NewBundle(language.English)
	mw := LocaleMiddleware(bundle)

	var capturedLocalizer interface{}
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedLocalizer = LocalizerFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "en-US")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
	if capturedLocalizer == nil {
		t.Error("LocalizerFromContext returned nil inside middleware")
	}
}

func TestLocalizerFromContext_Missing(t *testing.T) {
	loc := LocalizerFromContext(context.Background())
	if loc != nil {
		t.Error("LocalizerFromContext on empty context: expected nil, got non-nil")
	}
}
