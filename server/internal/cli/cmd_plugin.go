package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// PluginCmd returns the plugin management command group.
func PluginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage plugins",
	}

	cmd.AddCommand(pluginListCmd())
	cmd.AddCommand(pluginInstallCmd())
	cmd.AddCommand(pluginUninstallCmd())
	cmd.AddCommand(pluginStatusCmd())

	return cmd
}

func pluginListCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List installed plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			client := NewClientFromConfig()
			var plugins []map[string]any
			if err := client.Get("/api/plugins", &plugins); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(plugins)
			}
			printTable([]string{"UUID", "Name", "Version", "Status"}, func(row func(...string)) {
				for _, p := range plugins {
					row(str(p["uuid"]), str(p["name"]), str(p["version"]), str(p["status"]))
				}
			})
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func pluginInstallCmd() *cobra.Command {
	var (
		name       string
		version    string
		outputJSON bool
	)
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install a plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			body := map[string]any{"name": name}
			if version != "" {
				body["version"] = version
			}
			client := NewClientFromConfig()
			var plugin map[string]any
			if err := client.Post("/api/plugins", body, &plugin); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(plugin)
			}
			fmt.Printf("Installed plugin %s (%s)\n", str(plugin["name"]), str(plugin["uuid"]))
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Plugin name (required)")
	cmd.Flags().StringVar(&version, "version", "", "Plugin version (default: latest)")
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}

func pluginUninstallCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "uninstall <plugin-uuid>",
		Short: "Uninstall a plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			if !force {
				fmt.Printf("Uninstall plugin %s? Pass --force to confirm.\n", args[0])
				return nil
			}
			client := NewClientFromConfig()
			if err := client.Delete("/api/plugins/"+args[0], nil); err != nil {
				return err
			}
			fmt.Printf("Uninstalled plugin %s\n", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Confirm uninstall")
	return cmd
}

func pluginStatusCmd() *cobra.Command {
	var outputJSON bool
	cmd := &cobra.Command{
		Use:   "status <plugin-uuid>",
		Short: "Get plugin status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			client := NewClientFromConfig()
			var plugin map[string]any
			if err := client.Get("/api/plugins/"+args[0], &plugin); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(plugin)
			}
			printKV(plugin)
			return nil
		},
	}
	cmd.Flags().BoolVar(&outputJSON, "json", false, "Output as JSON")
	return cmd
}
