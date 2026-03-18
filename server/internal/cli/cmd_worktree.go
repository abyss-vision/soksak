package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// WorktreeCmd returns the worktree setup command group.
func WorktreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "worktree",
		Short: "Manage git worktree development environments",
	}

	cmd.AddCommand(worktreeInitCmd())
	cmd.AddCommand(worktreeListCmd())
	cmd.AddCommand(worktreeEnvCmd())

	return cmd
}

func worktreeInitCmd() *cobra.Command {
	var (
		name       string
		serverPort int
		dbPort     int
		force      bool
	)
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new worktree environment with isolated config",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				// Only attempt git detection when no explicit name is provided.
				gitRoot, err := gitRootDir()
				if err != nil {
					return fmt.Errorf("not inside a git repository: %w", err)
				}
				name = suggestWorktreeName(gitRoot)
			}

			home, _ := os.UserHomeDir()
			worktreeHome := filepath.Join(home, ".soksak", "worktrees", name)

			if _, err := os.Stat(worktreeHome); err == nil && !force {
				return fmt.Errorf("worktree environment %q already exists at %s — pass --force to reinitialize", name, worktreeHome)
			}

			if err := os.MkdirAll(worktreeHome, 0o700); err != nil {
				return fmt.Errorf("create worktree dir: %w", err)
			}

			if serverPort == 0 {
				serverPort = 3200
			}
			if dbPort == 0 {
				dbPort = 54430
			}

			// Write a minimal config for the worktree.
			cfg := fmt.Sprintf(`{
  "worktree": {
    "name": "%s",
    "home": "%s",
    "serverPort": %d,
    "dbPort": %d
  },
  "server": {
    "url": "http://localhost:%d"
  }
}`, name, worktreeHome, serverPort, dbPort, serverPort)

			cfgFile := filepath.Join(worktreeHome, "config.json")
			if err := os.WriteFile(cfgFile, []byte(cfg), 0o600); err != nil {
				return fmt.Errorf("write worktree config: %w", err)
			}

			fmt.Printf("Worktree environment %q initialized at %s\n", name, worktreeHome)
			fmt.Printf("Server port: %d\n", serverPort)
			fmt.Printf("DB port:     %d\n", dbPort)
			fmt.Printf("Config:      %s\n", cfgFile)
			fmt.Println()
			fmt.Println("To use this worktree:")
			fmt.Printf("  export SOKSAK_CONFIG=%s\n", cfgFile)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Worktree environment name (default: branch name)")
	cmd.Flags().IntVar(&serverPort, "server-port", 0, "Server port (default: 3200)")
	cmd.Flags().IntVar(&dbPort, "db-port", 0, "Embedded Postgres port (default: 54430)")
	cmd.Flags().BoolVar(&force, "force", false, "Reinitialize if already exists")
	return cmd
}

func worktreeListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List worktree environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			dir := filepath.Join(home, ".soksak", "worktrees")
			entries, err := os.ReadDir(dir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No worktree environments found.")
					return nil
				}
				return err
			}
			if len(entries) == 0 {
				fmt.Println("No worktree environments found.")
				return nil
			}
			printTable([]string{"Name", "Path"}, func(row func(...string)) {
				for _, e := range entries {
					if e.IsDir() {
						row(e.Name(), filepath.Join(dir, e.Name()))
					}
				}
			})
			return nil
		},
	}
}

func worktreeEnvCmd() *cobra.Command {
	var (
		name string
	)
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Print shell exports for a worktree environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				gitRoot, err := gitRootDir()
				if err != nil {
					return fmt.Errorf("--name required (not inside a git repo)")
				}
				name = suggestWorktreeName(gitRoot)
			}
			home, _ := os.UserHomeDir()
			cfgFile := filepath.Join(home, ".soksak", "worktrees", name, "config.json")
			fmt.Printf("export SOKSAK_CONFIG=%s\n", cfgFile)
			return nil
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Worktree name (default: current branch)")
	return cmd
}

func gitRootDir() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func currentBranch() string {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "main"
	}
	return strings.TrimSpace(string(out))
}

func suggestWorktreeName(gitRoot string) string {
	branch := currentBranch()
	// Sanitize branch name to be safe as a directory name.
	safe := strings.NewReplacer("/", "-", " ", "-", ".", "-").Replace(branch)
	_ = gitRoot
	return safe
}
