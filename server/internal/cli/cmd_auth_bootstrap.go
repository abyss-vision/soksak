package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// AuthBootstrapCmd returns the auth-bootstrap-ceo command.
func AuthBootstrapCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Auth management commands",
	}

	cmd.AddCommand(authBootstrapCEOCmd())

	return cmd
}

func authBootstrapCEOCmd() *cobra.Command {
	var (
		baseURL      string
		expiresHours int
		force        bool
		outputJSON   bool
	)
	cmd := &cobra.Command{
		Use:   "bootstrap-ceo",
		Short: "Create the first admin (CEO) bootstrap invite",
		Long: `Creates a one-time bootstrap invite link for the first CEO / instance admin.
Only works when the server is in 'authenticated' deployment mode.
The token is invalidated after use.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			client := NewClientFromConfig()

			body := map[string]any{
				"force": force,
			}
			if expiresHours > 0 {
				body["expiresHours"] = expiresHours
			}
			if baseURL != "" {
				body["baseUrl"] = baseURL
			}

			var result map[string]any
			if err := client.Post("/api/auth/bootstrap-ceo", body, &result); err != nil {
				return fmt.Errorf("bootstrap-ceo: %w", err)
			}

			if outputJSON {
				return printJSON(result)
			}

			inviteURL := str(result["inviteUrl"])
			expires := str(result["expiresAt"])
			if inviteURL == "" {
				fmt.Println("Bootstrap invite created (no URL returned — check server logs).")
			} else {
				fmt.Printf("Bootstrap invite URL: %s\n", inviteURL)
			}
			if expires != "" {
				fmt.Printf("Expires: %s\n", expires)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&baseURL, "base-url", "", "Override base URL for the invite link")
	cmd.Flags().IntVar(&expiresHours, "expires-hours", 72, "Invite expiry in hours (1-720)")
	cmd.Flags().BoolVar(&force, "force", false, "Create new invite even if an admin already exists")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}
