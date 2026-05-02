// pulley is an interactive terminal UI for GitHub pull request review.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/github"
	"github.com/rcorre/pulley/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "pulley [<number> | <url> | <branch>]",
	Short: "Interactive terminal UI for GitHub pull request review",
	Long: `pulley opens a pull request for review in an interactive TUI.

Without arguments, opens the PR associated with the current branch.
Accepts a PR number, URL, or branch name, matching the behavior of 'gh pr view'.`,
	Args:         cobra.MaximumNArgs(1),
	RunE:         run,
	SilenceUsage: true,
}

func run(_ *cobra.Command, args []string) error {
	var arg string
	if len(args) > 0 {
		arg = args[0]
	}

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("locate config dir: %w", err)
	}
	cfg, err := config.Load(filepath.Join(cfgDir, "pulley", "config.toml"))
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("create GitHub client: %w", err)
	}

	ref, err := github.Resolve(client, arg)
	if err != nil {
		return fmt.Errorf("resolve PR: %w", err)
	}

	m := tui.New(client, *ref, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
