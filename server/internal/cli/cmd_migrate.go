package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"soksak/internal/db"
)

// MigrateCmd returns the migrate command.
func MigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			databaseURL := os.Getenv("DATABASE_URL")
			if databaseURL == "" {
				return fmt.Errorf("DATABASE_URL is required")
			}

			database, err := db.NewDB(databaseURL)
			if err != nil {
				return fmt.Errorf("connect to database: %w", err)
			}
			defer database.Close()

			if err := db.RunMigrations(context.Background(), database.DB); err != nil {
				return fmt.Errorf("run migrations: %w", err)
			}

			fmt.Println("Migrations complete.")
			return nil
		},
	}
}
