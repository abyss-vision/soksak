package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// ApprovalCmd returns the approval management command group.
func ApprovalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "approval",
		Short: "Manage approvals",
	}

	cmd.PersistentFlags().String("company", "", "Company UUID (overrides company.default)")

	cmd.AddCommand(approvalListCmd())
	cmd.AddCommand(approvalGetCmd())
	cmd.AddCommand(approvalResolveCmd())

	return cmd
}

func approvalListCmd() *cobra.Command {
	var (
		status     string
		outputJSON bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List approvals for a company",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			path := "/api/companies/" + company + "/approvals"
			if status != "" {
				path += "?status=" + status
			}
			var approvals []map[string]any
			if err := client.Get(path, &approvals); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(approvals)
			}
			printTable([]string{"UUID", "Type", "Status", "Created"}, func(row func(...string)) {
				for _, a := range approvals {
					row(str(a["uuid"]), str(a["approvalType"]), str(a["status"]), str(a["createdAt"]))
				}
			})
			return nil
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by status (pending|approved|rejected)")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func approvalGetCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "get <approval-uuid>",
		Short: "Get an approval by UUID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			var approval map[string]any
			if err := client.Get("/api/companies/"+company+"/approvals/"+args[0], &approval); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(approval)
			}
			printKV(approval)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func approvalResolveCmd() *cobra.Command {
	var (
		decision string
		comment  string
	)
	cmd := &cobra.Command{
		Use:   "resolve <approval-uuid>",
		Short: "Resolve an approval (approve or reject)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			if decision != "approve" && decision != "reject" {
				return fmt.Errorf("--decision must be 'approve' or 'reject'")
			}
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			body := map[string]any{"decision": decision}
			if comment != "" {
				body["comment"] = comment
			}
			client := NewClientFromConfig()
			var approval map[string]any
			if err := client.Post("/api/companies/"+company+"/approvals/"+args[0]+"/resolve", body, &approval); err != nil {
				return err
			}
			fmt.Printf("Approval %s resolved: %s\n", args[0], decision)
			return nil
		},
	}
	cmd.Flags().StringVar(&decision, "decision", "", "approve or reject (required)")
	cmd.Flags().StringVar(&comment, "comment", "", "Optional resolution comment")
	return cmd
}
