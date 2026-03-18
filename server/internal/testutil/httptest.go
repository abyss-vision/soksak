package testutil

import (
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"abyss-view/internal/server"
)

// NewTestRouter creates an App with a test config and returns its router for use in HTTP tests.
func NewTestRouter(t *testing.T, db *sqlx.DB) chi.Router {
	t.Helper()

	cfg := &server.Config{
		Port:           "0",
		Host:           "localhost",
		DeploymentMode: "test",
		ServeUI:        false,
	}

	app := server.NewApp(cfg, db, nil)
	return app.Router
}
