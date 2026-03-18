package main

import (
	"log/slog"
	"os"

	"github.com/jmoiron/sqlx"
	"golang.org/x/text/language"

	apii18n "abyss-view/internal/i18n"
	"abyss-view/internal/db"
	"abyss-view/internal/server"
)

func main() {
	cfg, err := server.LoadConfig()
	if err != nil {
		slog.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	bundle := apii18n.NewBundle(language.English)
	localeFiles := []string{
		"internal/i18n/locales/en.json",
		"internal/i18n/locales/ko.json",
		"internal/i18n/locales/ja.json",
	}
	for _, f := range localeFiles {
		if _, err := bundle.LoadMessageFile(f); err != nil {
			slog.Warn("could not load locale file", "file", f, "err", err)
		}
	}

	var database *sqlx.DB
	if cfg.DatabaseURL != "" {
		database, err = db.NewDB(cfg.DatabaseURL)
		if err != nil {
			slog.Error("failed to connect to database", "err", err)
			os.Exit(1)
		}
	}

	app := server.NewApp(cfg, database, bundle)
	if err := server.Run(app); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
