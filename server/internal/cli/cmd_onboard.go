package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// OnboardCmd returns the interactive setup wizard command.
func OnboardCmd() *cobra.Command {
	var (
		serverURL string
		token     string
		company   string
		yes       bool
	)

	cmd := &cobra.Command{
		Use:   "onboard",
		Short: "Interactive setup wizard — configure server URL, token, and default company",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()

			fmt.Println("=== Soksak CLI Setup ===")

			if !yes {
				serverURL = prompt("Server URL", "http://localhost:3100")
				token = prompt("API token (leave blank if unauthenticated)", "")
				company = prompt("Default company UUID (optional)", "")
			}

			if serverURL == "" {
				serverURL = "http://localhost:3100"
			}

			if err := WriteConfigValue("server.url", serverURL); err != nil {
				return fmt.Errorf("save server.url: %w", err)
			}
			if token != "" {
				if err := WriteConfigValue("server.token", token); err != nil {
					return fmt.Errorf("save server.token: %w", err)
				}
			}
			if company != "" {
				if err := WriteConfigValue("company.default", company); err != nil {
					return fmt.Errorf("save company.default: %w", err)
				}
			}

			// Verify connectivity.
			client := NewClient(serverURL, token)
			var health map[string]any
			if err := client.Get("/api/health", &health); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not reach server at %s: %v\n", serverURL, err)
				fmt.Println("Configuration saved. Run 'soksak-cli doctor' to diagnose.")
			} else {
				fmt.Printf("Server reachable at %s\n", serverURL)
				fmt.Println("Configuration saved successfully.")
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&serverURL, "url", "", "Server base URL")
	cmd.Flags().StringVar(&token, "token", "", "API bearer token")
	cmd.Flags().StringVar(&company, "company", "", "Default company UUID")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "Accept defaults without prompting")

	return cmd
}

func prompt(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}
