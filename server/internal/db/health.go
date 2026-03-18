package db

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// HealthCheck verifies the database connection is alive via PingContext.
func HealthCheck(ctx context.Context, database *sqlx.DB) error {
	return database.PingContext(ctx)
}
