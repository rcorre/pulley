// pulley is an interactive terminal UI for GitHub pull request review.
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/github"
	plog "github.com/rcorre/pulley/internal/log"
	"github.com/rcorre/pulley/internal/tui"
)

var rootCmd = &cobra.Command{
	Use:   "pulley [<number> | <url> | <branch>]",
	Short: "Interactive terminal UI for GitHub pull request review",
	Long: `pulley opens a pull request for review in an interactive TUI.

Without arguments, opens the PR associated with the current branch.
Accepts a PR number, URL, or branch name, matching the behavior of 'gh pr view'.`,
	Args:              cobra.MaximumNArgs(1),
	RunE:              run,
	SilenceUsage:      true,
	ValidArgsFunction: noFileCompletion,
}

// noFileCompletion tells the shell not to fall back to filename completion for
// the PR argument (numbers/URLs/branches can't be completed from the filesystem).
func noFileCompletion(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func run(_ *cobra.Command, args []string) error {
	var arg string
	if len(args) > 0 {
		arg = args[0]
	}

	logCloser, err := plog.Init(os.Getenv("PULLEY_LOG"))
	if err != nil {
		return fmt.Errorf("init logging: %w", err)
	}
	defer func() { _ = logCloser.Close() }()
	slog.Info("starting pulley", "arg", arg)

	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("locate config dir: %w", err)
	}
	cfgPath := filepath.Join(cfgDir, "pulley", "config.toml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	slog.Info("config loaded", "path", cfgPath)

	client, err := github.NewClient()
	if err != nil {
		return fmt.Errorf("create GitHub client: %w", err)
	}

	ref, err := github.Resolve(client, arg)
	if err != nil {
		return fmt.Errorf("resolve PR: %w", err)
	}
	slog.Info("PR resolved", "owner", ref.Owner, "repo", ref.Repo, "number", ref.Number)

	m := tui.New(client, *ref, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("run TUI: %w", err)
	}
	slog.Info("TUI exited")
	return nil
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
