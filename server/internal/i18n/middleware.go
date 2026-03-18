package i18n

import (
	"context"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type contextKey string

const localizerKey contextKey = "localizer"

// LocaleMiddleware parses the Accept-Language header and stores a Localizer in the request context.
func LocaleMiddleware(bundle *i18n.Bundle) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lang := r.Header.Get("Accept-Language")
			localizer := GetLocalizer(bundle, lang)
			ctx := context.WithValue(r.Context(), localizerKey, localizer)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LocalizerFromContext retrieves the Localizer stored in the context by LocaleMiddleware.
func LocalizerFromContext(ctx context.Context) *i18n.Localizer {
	localizer, _ := ctx.Value(localizerKey).(*i18n.Localizer)
	return localizer
}
