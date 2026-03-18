package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// DoctorCmd returns the diagnostics command.
func DoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Run diagnostics: server reachable, DB connected, config valid",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			cfg := GetConfig()

			passed := 0
			warned := 0
			failed := 0

			check := func(name string, fn func() (string, bool, bool)) {
				msg, ok, warn := fn()
				switch {
				case ok:
					fmt.Printf("[PASS] %s: %s\n", name, msg)
					passed++
				case warn:
					fmt.Fprintf(os.Stderr, "[WARN] %s: %s\n", name, msg)
					warned++
				default:
					fmt.Fprintf(os.Stderr, "[FAIL] %s: %s\n", name, msg)
					failed++
				}
			}

			// 1. Config valid
			check("Config", func() (string, bool, bool) {
				if cfg.ServerURL == "" {
					return "server.url not set — run 'soksak-cli onboard'", false, false
				}
				return fmt.Sprintf("server.url=%s", cfg.ServerURL), true, false
			})

			// 2. Server reachable
			check("Server reachable", func() (string, bool, bool) {
				client := NewClientFromConfig()
				var health map[string]any
				if err := client.Get("/api/health", &health); err != nil {
					return fmt.Sprintf("cannot reach %s: %v", cfg.ServerURL, err), false, false
				}
				return cfg.ServerURL, true, false
			})

			// 3. Auth token configured
			check("Auth token", func() (string, bool, bool) {
				if cfg.ServerToken == "" {
					return "no token configured (unauthenticated access)", true, true
				}
				return "token present", true, false
			})

			// 4. Default company configured
			check("Default company", func() (string, bool, bool) {
				if cfg.CompanyDefault == "" {
					return "company.default not set — some commands require --company", true, true
				}
				return cfg.CompanyDefault, true, false
			})

			fmt.Printf("\nSummary: %d passed, %d warnings, %d failed\n", passed, warned, failed)
			if failed > 0 {
				return fmt.Errorf("%d check(s) failed", failed)
			}
			return nil
		},
	}
}
