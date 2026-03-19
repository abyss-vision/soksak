package db

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewDB creates a new sqlx.DB using the pgx/v5 driver with recommended pool settings.
// It retries the connection up to 10 times with 2-second intervals.
func NewDB(databaseURL string) (*sqlx.DB, error) {
	database, err := sqlx.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	const maxRetries = 10
	const retryInterval = 2 * time.Second

	for i := range maxRetries {
		if err = database.Ping(); err == nil {
			return database, nil
		}
		slog.Warn("db connection failed, retrying",
			"attempt", fmt.Sprintf("%d/%d", i+1, maxRetries),
			"err", err,
		)
		time.Sleep(retryInterval)
	}

	database.Close()
	return nil, fmt.Errorf("failed to connect after %d attempts: %w", maxRetries, err)
}
