package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

// StatusBar displays PR info, draft count, and key hints.
type StatusBar struct {
	pr         *PRInfo
	draftCount int
	width      int
	styles     Styles
}

// PRInfo holds minimal PR info for the status bar.
type PRInfo struct {
	Number int
	Title  string
}

// NewStatusBar creates a new status bar.
func NewStatusBar(styles Styles) StatusBar {
	return StatusBar{styles: styles}
}

// SetPR sets the PR info.
func (s *StatusBar) SetPR(pr *PRInfo) {
	s.pr = pr
}

// SetDraftCount updates the draft comment count.
func (s *StatusBar) SetDraftCount(n int) {
	s.draftCount = n
}

// SetWidth sets the status bar width.
func (s *StatusBar) SetWidth(w int) {
	s.width = w
}

// View renders the status bar.
func (s StatusBar) View() string {
	var parts []string

	if s.pr != nil {
		pr := fmt.Sprintf("#%d: %s", s.pr.Number, s.pr.Title)
		parts = append(parts, s.styles.StatusFg.Render(pr))
	}

	if s.draftCount > 0 {
		drafts := fmt.Sprintf("[%d draft(s)]", s.draftCount)
		parts = append(parts, s.styles.DraftFg.Render(drafts))
	}

	// Key hints
	hints := s.styles.StatusFg.Render("q: quit | tab: focus")
	parts = append(parts, hints)

	line := strings.Join(parts, "  ")

	// Truncate to width if needed
	line = lipgloss.NewStyle().MaxWidth(s.width).Render(line)

	return s.styles.StatusBg.Render(line)
}
