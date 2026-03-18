package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Valid issue state transitions (mirrors the server state machine).
var issueTransitions = map[string][]string{
	"open":        {"in_progress", "closed"},
	"in_progress": {"open", "closed", "blocked"},
	"blocked":     {"in_progress", "closed"},
	"closed":      {"open"},
}

// IssueCmd returns the issue management command group.
func IssueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue",
		Short: "Manage issues",
	}

	cmd.PersistentFlags().String("company", "", "Company UUID (overrides company.default)")

	cmd.AddCommand(issueListCmd())
	cmd.AddCommand(issueGetCmd())
	cmd.AddCommand(issueCreateCmd())
	cmd.AddCommand(issueUpdateCmd())
	cmd.AddCommand(issueTransitionCmd())

	return cmd
}

func issueListCmd() *cobra.Command {
	var (
		status     string
		outputJSON bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issues for a company",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			path := "/api/companies/" + company + "/issues"
			if status != "" {
				path += "?status=" + status
			}
			var issues []map[string]any
			if err := client.Get(path, &issues); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(issues)
			}
			printTable([]string{"UUID", "Title", "Status", "Priority"}, func(row func(...string)) {
				for _, i := range issues {
					row(str(i["uuid"]), str(i["title"]), str(i["status"]), str(i["priority"]))
				}
			})
			return nil
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func issueGetCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "get <issue-uuid>",
		Short: "Get an issue by UUID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			var issue map[string]any
			if err := client.Get("/api/companies/"+company+"/issues/"+args[0], &issue); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(issue)
			}
			printKV(issue)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func issueCreateCmd() *cobra.Command {
	var (
		title       string
		description string
		priority    string
		agentUUID   string
		outputJSON  bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issue",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			if title == "" {
				return fmt.Errorf("--title is required")
			}
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			body := map[string]any{"title": title}
			if description != "" {
				body["description"] = description
			}
			if priority != "" {
				body["priority"] = priority
			}
			if agentUUID != "" {
				body["agentUuid"] = agentUUID
			}
			client := NewClientFromConfig()
			var issue map[string]any
			if err := client.Post("/api/companies/"+company+"/issues", body, &issue); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(issue)
			}
			fmt.Printf("Created issue %s (%s)\n", str(issue["title"]), str(issue["uuid"]))
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Issue title (required)")
	cmd.Flags().StringVar(&description, "description", "", "Issue description")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority: low|medium|high|critical")
	cmd.Flags().StringVar(&agentUUID, "agent", "", "Assigned agent UUID")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func issueUpdateCmd() *cobra.Command {
	var (
		title       string
		description string
		priority    string
		outputJSON  bool
	)
	cmd := &cobra.Command{
		Use:   "update <issue-uuid>",
		Short: "Update an issue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			body := map[string]any{}
			if title != "" {
				body["title"] = title
			}
			if description != "" {
				body["description"] = description
			}
			if priority != "" {
				body["priority"] = priority
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to update — provide at least one flag")
			}
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()
			var issue map[string]any
			if err := client.Patch("/api/companies/"+company+"/issues/"+args[0], body, &issue); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(issue)
			}
			fmt.Printf("Updated issue %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "New title")
	cmd.Flags().StringVar(&description, "description", "", "New description")
	cmd.Flags().StringVar(&priority, "priority", "", "New priority")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func issueTransitionCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "transition <issue-uuid> <new-status>",
		Short: "Transition an issue to a new status (validates state machine)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			issueUUID := args[0]
			newStatus := args[1]

			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}
			client := NewClientFromConfig()

			// Fetch current status for state machine validation.
			var issue map[string]any
			if err := client.Get("/api/companies/"+company+"/issues/"+issueUUID, &issue); err != nil {
				return err
			}
			currentStatus := str(issue["status"])

			allowed, ok := issueTransitions[currentStatus]
			if !ok {
				return fmt.Errorf("unknown current status %q", currentStatus)
			}
			valid := false
			for _, s := range allowed {
				if s == newStatus {
					valid = true
					break
				}
			}
			if !valid {
				return fmt.Errorf("invalid transition %q -> %q; allowed: %v", currentStatus, newStatus, allowed)
			}

			body := map[string]any{"status": newStatus}
			var updated map[string]any
			if err := client.Patch("/api/companies/"+company+"/issues/"+issueUUID, body, &updated); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(updated)
			}
			fmt.Printf("Issue %s transitioned %s -> %s\n", issueUUID, currentStatus, newStatus)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}
