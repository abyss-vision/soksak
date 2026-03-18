package i18n

import (
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

// NewBundle creates a new i18n bundle loading JSON locale files from the embedded locales directory.
func NewBundle(defaultLang language.Tag) *i18n.Bundle {
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	return bundle
}

// GetLocalizer returns a localizer for the given language string.
func GetLocalizer(bundle *i18n.Bundle, lang string) *i18n.Localizer {
	return i18n.NewLocalizer(bundle, lang)
}
