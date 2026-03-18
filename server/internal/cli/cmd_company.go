package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// CompanyCmd returns the company management command group.
func CompanyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "company",
		Short: "Manage companies",
	}

	cmd.AddCommand(companyListCmd())
	cmd.AddCommand(companyGetCmd())
	cmd.AddCommand(companyCreateCmd())
	cmd.AddCommand(companyUpdateCmd())
	cmd.AddCommand(companyDeleteCmd())

	return cmd
}

func companyListCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all companies",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			client := NewClientFromConfig()
			var companies []map[string]any
			if err := client.Get("/api/companies", &companies); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(companies)
			}
			printTable([]string{"UUID", "Name"}, func(row func(...string)) {
				for _, c := range companies {
					row(str(c["uuid"]), str(c["name"]))
				}
			})
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func companyGetCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "get <company-uuid>",
		Short: "Get a company by UUID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			client := NewClientFromConfig()
			var company map[string]any
			if err := client.Get("/api/companies/"+args[0], &company); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(company)
			}
			printKV(company)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func companyCreateCmd() *cobra.Command {
	var (
		name        string
		description string
		outputJSON  bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new company",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			client := NewClientFromConfig()
			body := map[string]any{"name": name}
			if description != "" {
				body["description"] = description
			}
			var company map[string]any
			if err := client.Post("/api/companies", body, &company); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(company)
			}
			fmt.Printf("Created company %s (%s)\n", str(company["name"]), str(company["uuid"]))
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Company name (required)")
	cmd.Flags().StringVar(&description, "description", "", "Company description")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func companyUpdateCmd() *cobra.Command {
	var (
		name        string
		description string
		outputJSON  bool
	)
	cmd := &cobra.Command{
		Use:   "update <company-uuid>",
		Short: "Update a company",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			body := map[string]any{}
			if name != "" {
				body["name"] = name
			}
			if description != "" {
				body["description"] = description
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to update — provide at least one flag")
			}
			client := NewClientFromConfig()
			var company map[string]any
			if err := client.Patch("/api/companies/"+args[0], body, &company); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(company)
			}
			fmt.Printf("Updated company %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "New company name")
	cmd.Flags().StringVar(&description, "description", "", "New company description")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func companyDeleteCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete <company-uuid>",
		Short: "Delete a company",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			if !force {
				fmt.Printf("Delete company %s? Pass --force to confirm.\n", args[0])
				return nil
			}
			client := NewClientFromConfig()
			if err := client.Delete("/api/companies/"+args[0], nil); err != nil {
				return err
			}
			fmt.Printf("Deleted company %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Confirm deletion")
	return cmd
}

// printJSON is a shared JSON pretty-printer used by multiple commands.
func printJSON(v any) error {
	enc, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(enc))
	return nil
}

// printKV prints a map as key: value lines.
func printKV(m map[string]any) {
	for k, v := range m {
		fmt.Printf("%-20s %v\n", k, v)
	}
}

// str safely converts an any to a string.
func str(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// printTable prints a simple ASCII table.
func printTable(headers []string, fill func(row func(...string))) {
	// Collect rows first to compute widths.
	var rows [][]string
	fill(func(cols ...string) {
		rows = append(rows, cols)
	})

	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	fmtRow := func(cols []string) {
		for i, col := range cols {
			pad := ""
			if i < len(widths) {
				pad = fmt.Sprintf("%-*s", widths[i], col)
			}
			if i > 0 {
				fmt.Print("  ")
			}
			fmt.Print(pad)
		}
		fmt.Println()
	}

	fmtRow(headers)
	sep := make([]string, len(headers))
	for i, w := range widths {
		sep[i] = repeatStr("-", w)
	}
	fmtRow(sep)
	for _, row := range rows {
		fmtRow(row)
	}
}

func repeatStr(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
