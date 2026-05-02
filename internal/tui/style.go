package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/rcorre/pulley/internal/config"
)

// Styles holds lipgloss styles for the application.
type Styles struct {
	AddFg      lipgloss.Style
	RemoveFg   lipgloss.Style
	HunkFg     lipgloss.Style
	FileHeader lipgloss.Style
	LineNum    lipgloss.Style
	CursorBg   lipgloss.Style
	CommentFg  lipgloss.Style
	DraftFg    lipgloss.Style
	StatusFg   lipgloss.Style
	StatusBg   lipgloss.Style
}

// NewStyles creates Styles from config colors.
func NewStyles(cfg config.Config) Styles {
	s := Styles{}
	cc := cfg.Colors

	fromColor := func(cv config.ColorValue) color.Color {
		if v := cv.String(); v != "" {
			return lipgloss.Color(v)
		}
		return nil
	}

	s.AddFg = lipgloss.NewStyle().Foreground(fromColor(cc.AddFg))
	s.RemoveFg = lipgloss.NewStyle().Foreground(fromColor(cc.RemoveFg))
	s.HunkFg = lipgloss.NewStyle().Foreground(fromColor(cc.HunkFg))
	s.FileHeader = lipgloss.NewStyle().Foreground(fromColor(cc.FileHeader)).Bold(true)
	s.LineNum = lipgloss.NewStyle().Foreground(fromColor(cc.LineNum))
	s.CursorBg = lipgloss.NewStyle().Background(fromColor(cc.CursorBg))
	s.CommentFg = lipgloss.NewStyle().Foreground(fromColor(cc.CommentFg))
	s.DraftFg = lipgloss.NewStyle().Foreground(fromColor(cc.DraftFg))
	s.StatusFg = lipgloss.NewStyle().Foreground(fromColor(cc.StatusFg))
	s.StatusBg = lipgloss.NewStyle().Background(fromColor(cc.StatusBg))

	return s
}
