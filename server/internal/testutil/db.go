package testutil

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// SetupTestDB spins up a postgres:17-alpine container via dockertest and returns a connected *sqlx.DB.
// The container is automatically removed when the test completes via t.Cleanup.
func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("could not connect to docker: %v", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "17-alpine",
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_DB=testdb",
		},
	}, func(hc *docker.HostConfig) {
		hc.AutoRemove = true
		hc.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		t.Fatalf("could not start postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			t.Logf("could not purge docker resource: %v", err)
		}
	})

	var database *sqlx.DB
	dsn := fmt.Sprintf(
		"postgres://test:test@localhost:%s/testdb?sslmode=disable",
		resource.GetPort("5432/tcp"),
	)

	if err := pool.Retry(func() error {
		var err error
		database, err = sqlx.Open("pgx", dsn)
		if err != nil {
			return err
		}
		return database.Ping()
	}); err != nil {
		t.Fatalf("could not connect to test database: %v", err)
	}

	t.Cleanup(func() {
		if database != nil {
			database.Close()
		}
	})

	return database
}
