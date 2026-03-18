package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AgentCmd returns the agent management command group.
func AgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage agents",
	}

	cmd.PersistentFlags().String("company", "", "Company UUID (overrides company.default)")

	cmd.AddCommand(agentListCmd())
	cmd.AddCommand(agentGetCmd())
	cmd.AddCommand(agentHireCmd())
	cmd.AddCommand(agentFireCmd())
	cmd.AddCommand(agentPauseCmd())
	cmd.AddCommand(agentResumeCmd())

	return cmd
}

func resolveCompany(cmd *cobra.Command) (string, error) {
	company, _ := cmd.Flags().GetString("company")
	if company == "" {
		// Walk up to persistent flags on parent.
		if p := cmd.Parent(); p != nil {
			company, _ = p.PersistentFlags().GetString("company")
		}
	}
	if company == "" {
		InitConfig()
		company = viper.GetString("company.default")
	}
	if company == "" {
		return "", fmt.Errorf("company UUID required — pass --company or set company.default")
	}
	return company, nil
}

func agentListCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agents for a company",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			var agents []map[string]any
			if err := client.Get("/api/companies/"+company+"/agents", &agents); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(agents)
			}
			printTable([]string{"UUID", "Name", "Role", "Status"}, func(row func(...string)) {
				for _, a := range agents {
					row(str(a["uuid"]), str(a["name"]), str(a["role"]), str(a["status"]))
				}
			})
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func agentGetCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "get <agent-uuid>",
		Short: "Get an agent by UUID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			var agent map[string]any
			if err := client.Get("/api/companies/"+company+"/agents/"+args[0], &agent); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(agent)
			}
			printKV(agent)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func agentHireCmd() *cobra.Command {
	var (
		name        string
		role        string
		adapterType string
		outputJSON  bool
	)
	cmd := &cobra.Command{
		Use:   "hire",
		Short: "Hire (create) a new agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			if name == "" || role == "" || adapterType == "" {
				return fmt.Errorf("--name, --role, and --adapter-type are required")
			}
			client := NewClientFromConfig()
			body := map[string]any{
				"name":        name,
				"role":        role,
				"adapterType": adapterType,
			}
			var agent map[string]any
			if err := client.Post("/api/companies/"+company+"/agents", body, &agent); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(agent)
			}
			fmt.Printf("Hired agent %s (%s)\n", str(agent["name"]), str(agent["uuid"]))
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Agent name (required)")
	cmd.Flags().StringVar(&role, "role", "", "Agent role (required)")
	cmd.Flags().StringVar(&adapterType, "adapter-type", "", "Adapter type (required)")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func agentFireCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "fire <agent-uuid>",
		Short: "Fire (delete) an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			if !force {
				fmt.Printf("Fire agent %s? Pass --force to confirm.\n", args[0])
				return nil
			}
			client := NewClientFromConfig()
			if err := client.Delete("/api/companies/"+company+"/agents/"+args[0], nil); err != nil {
				return err
			}
			fmt.Printf("Fired agent %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion")
	return cmd
}

func agentPauseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause <agent-uuid>",
		Short: "Pause an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			var agent map[string]any
			if err := client.Post("/api/companies/"+company+"/agents/"+args[0]+"/pause", nil, &agent); err != nil {
				return err
			}
			fmt.Printf("Paused agent %s\n", args[0])
			return nil
		},
	}
	return cmd
}

func agentResumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <agent-uuid>",
		Short: "Resume a paused agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			var agent map[string]any
			if err := client.Post("/api/companies/"+company+"/agents/"+args[0]+"/resume", nil, &agent); err != nil {
				return err
			}
			fmt.Printf("Resumed agent %s\n", args[0])
			return nil
		},
	}
	return cmd
}
