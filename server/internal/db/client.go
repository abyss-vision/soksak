package db

import (
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// NewDB creates a new sqlx.DB using the pgx/v5 driver with recommended pool settings.
// It pings the database to verify connectivity before returning.
func NewDB(databaseURL string) (*sqlx.DB, error) {
	database, err := sqlx.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	database.SetMaxOpenConns(25)
	database.SetMaxIdleConns(5)
	database.SetConnMaxLifetime(5 * time.Minute)

	if err := database.Ping(); err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}
