package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"abyss-view/internal/cli"
)

var version = "dev"

func main() {
	root := &cobra.Command{
		Use:     "soksak-cli",
		Short:   "Soksak CLI — manage your Soksak instance",
		Version: version,
	}

	root.AddCommand(cli.OnboardCmd())
	root.AddCommand(cli.DoctorCmd())
	root.AddCommand(cli.ConfigCmd())
	root.AddCommand(cli.CompanyCmd())
	root.AddCommand(cli.AgentCmd())
	root.AddCommand(cli.IssueCmd())
	root.AddCommand(cli.ApprovalCmd())
	root.AddCommand(cli.ActivityCmd())
	root.AddCommand(cli.DashboardCmd())
	root.AddCommand(cli.PluginCmd())
	root.AddCommand(cli.AuthBootstrapCmd())
	root.AddCommand(cli.WorktreeCmd())

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
