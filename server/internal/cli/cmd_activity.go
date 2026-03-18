package cli

import (
	"github.com/spf13/cobra"
)

// ActivityCmd returns the activity log command group.
func ActivityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activity",
		Short: "Browse activity logs",
	}

	cmd.PersistentFlags().String("company", "", "Company UUID (overrides company.default)")

	cmd.AddCommand(activityListCmd())

	return cmd
}

func activityListCmd() *cobra.Command {
	var (
		limit      int
		offset     int
		from       string
		to         string
		agentUUID  string
		entityType string
		entityID   string
		outputJSON bool
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List activity log entries with optional filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			company, err := resolveCompany(cmd)
			if err != nil {
				return err
			}

			path := "/api/companies/" + company + "/activity"
			sep := "?"
			addParam := func(key, val string) {
				if val == "" {
					return
				}
				path += sep + key + "=" + val
				sep = "&"
			}
			addIntParam := func(key string, val int) {
				if val <= 0 {
					return
				}
				path += sep + key + "=" + intToStr(val)
				sep = "&"
			}

			addIntParam("limit", limit)
			addIntParam("offset", offset)
			addParam("from", from)
			addParam("to", to)
			addParam("agentUuid", agentUUID)
			addParam("entityType", entityType)
			addParam("entityId", entityID)

			client := NewClientFromConfig()
			var entries []map[string]any
			if err := client.Get(path, &entries); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(entries)
			}
			printTable([]string{"UUID", "Type", "Agent", "Entity", "Created"}, func(row func(...string)) {
				for _, e := range entries {
					row(str(e["uuid"]), str(e["activityType"]), str(e["agentUuid"]), str(e["entityType"]), str(e["createdAt"]))
				}
			})
			return nil
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 0, "Max number of results")
	cmd.Flags().IntVar(&offset, "offset", 0, "Pagination offset")
	cmd.Flags().StringVar(&from, "from", "", "Start date (RFC3339)")
	cmd.Flags().StringVar(&to, "to", "", "End date (RFC3339)")
	cmd.Flags().StringVar(&agentUUID, "agent", "", "Filter by agent UUID")
	cmd.Flags().StringVar(&entityType, "entity-type", "", "Filter by entity type")
	cmd.Flags().StringVar(&entityID, "entity-id", "", "Filter by entity ID")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func intToStr(n int) string {
	return str(n)
}
