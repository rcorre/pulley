// pulley is an interactive terminal UI for GitHub pull request review.
package main

import (
	"fmt"
	"os"

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

func run(cmd *cobra.Command, args []string) error {
	var arg string
	if len(args) > 0 {
		arg = args[0]
	}
	// TODO: replace with full startup flow in Unit 6
	_, err := fmt.Fprintf(cmd.OutOrStdout(), "pulley: %q\n", arg)
	return err
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
