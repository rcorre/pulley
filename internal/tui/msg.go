package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/github"
)

// PRLoadedMsg is sent when PR data finishes loading.
type PRLoadedMsg struct {
	PR       *github.PR
	Diff     string
	Comments []github.ReviewComment
}

// PRFailedMsg is sent when PR data fails to load.
type PRFailedMsg struct {
	Err error
}

// LoadPR returns a tea.Cmd that fetches PR data asynchronously.
func LoadPR(client github.PRClient, ref *github.PRRef) func() tea.Msg {
	return func() tea.Msg {
		pr, err := client.GetPR(ref.Owner, ref.Repo, ref.Number)
		if err != nil {
			return PRFailedMsg{Err: err}
		}
		diff, err := client.GetDiff(ref.Owner, ref.Repo, ref.Number)
		if err != nil {
			return PRFailedMsg{Err: err}
		}
		comments, err := client.GetComments(ref.Owner, ref.Repo, ref.Number)
		if err != nil {
			return PRFailedMsg{Err: err}
		}
		return PRLoadedMsg{PR: pr, Diff: diff, Comments: comments}
	}
}
