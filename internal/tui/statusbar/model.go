// Package statusbar provides the status bar component for the pulley TUI.
package statusbar

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/github"
)

const keyHints = "q:quit "

// Model is the status bar displayed at the bottom of the screen.
// It shows the PR title, number, draft count, and key hints.
type Model struct {
	pr         *github.PR
	draftCount int
	width      int
	style      lipgloss.Style
}

// New creates a new status bar Model with styles derived from the given ColorConfig.
func New(colors config.ColorConfig) Model {
	return Model{
		style: lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.StatusFg.String())).
			Background(lipgloss.Color(colors.StatusBg.String())),
	}
}

// SetPR sets the pull request shown in the status bar.
func (m *Model) SetPR(pr *github.PR) { m.pr = pr }

// SetDraftCount sets the number of unsaved draft comments.
func (m *Model) SetDraftCount(n int) { m.draftCount = n }

// SetWidth sets the render width of the status bar.
func (m *Model) SetWidth(w int) { m.width = w }

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model.
func (m Model) Update(_ tea.Msg) (Model, tea.Cmd) { return m, nil }

// View renders the status bar.
func (m Model) View() string {
	if m.width == 0 {
		return ""
	}
	var left, right string
	if m.pr == nil {
		left = " Loading..."
		right = keyHints
	} else {
		left = fmt.Sprintf(" #%d %s", m.pr.Number, m.pr.Title)
		if m.draftCount > 0 {
			right = fmt.Sprintf("%d draft(s)  %s", m.draftCount, keyHints)
		} else {
			right = keyHints
		}
	}

	gap := max(1, m.width-len(left)-len(right))
	content := left + strings.Repeat(" ", gap) + right
	return m.style.Width(m.width).Render(content)
}
