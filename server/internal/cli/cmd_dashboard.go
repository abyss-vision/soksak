package cli

import (
	"github.com/spf13/cobra"
)

// DashboardCmd returns the dashboard summary command.
func DashboardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboard",
		Short: "Show dashboard summary",
	}

	cmd.PersistentFlags().String("company", "", "Company UUID (overrides company.default)")

	cmd.AddCommand(dashboardShowCmd())

	return cmd
}

func dashboardShowCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Print the dashboard summary table",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			var summary map[string]any
			if err := client.Get("/api/companies/"+company+"/dashboard", &summary); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(summary)
			}
			// Print summary as key-value pairs.
			printKV(summary)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}
