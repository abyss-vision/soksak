package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// ConfigCmd returns the config get/set command.
func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Read or write CLI configuration values",
	}

	cmd.AddCommand(configGetCmd())
	cmd.AddCommand(configSetCmd())
	cmd.AddCommand(configListCmd())

	return cmd
}

func configGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Print a config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			key := args[0]
			val := viper.GetString(key)
			if val == "" {
				fmt.Printf("%s = (not set)\n", key)
			} else {
				fmt.Printf("%s = %s\n", key, val)
			}
			return nil
		},
	}
}

func configSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value and persist it to ~/.soksak/config.json",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			key, value := args[0], args[1]
			if err := WriteConfigValue(key, value); err != nil {
				return err
			}
			fmt.Printf("Set %s = %s\n", key, value)
			return nil
		},
	}
}

func configListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all config values",
		RunE: func(cmd *cobra.Command, args []string) error {
			InitConfig()
			for _, key := range []string{"server.url", "server.token", "company.default"} {
				val := viper.GetString(key)
				if val == "" {
					fmt.Printf("%-22s (not set)\n", key)
				} else {
					fmt.Printf("%-22s %s\n", key, val)
				}
			}
			return nil
		},
	}
}
