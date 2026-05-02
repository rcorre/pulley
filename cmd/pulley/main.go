// pulley is an interactive terminal UI for GitHub pull request review.
package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/github"
	"github.com/rcorre/pulley/internal/tui"
	"github.com/spf13/cobra"
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

	// Load config
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}
	configPath := filepath.Join(home, ".config", "pulley", "config.toml")
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// Create GitHub client
	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("create GitHub client: %w", err)
	}

	// Resolve PR reference
	prRef, err := github.Resolve(client, arg)
	if err != nil {
		return fmt.Errorf("resolve PR: %w", err)
	}

	// Create and run TUI
	model := tui.NewModel(cfg, client, prRef)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	_, err = p.Run()
	return err
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
