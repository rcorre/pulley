package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/rcorre/pulley/internal/config"
)

// Styles holds lipgloss styles for each diff element, derived from ColorConfig.
type Styles struct {
	AddLine    lipgloss.Style
	RemoveLine lipgloss.Style
	HunkHeader lipgloss.Style
	FileHeader lipgloss.Style
	LineNum    lipgloss.Style
	CursorLine lipgloss.Style
	CommentFg  lipgloss.Style
	DraftFg    lipgloss.Style
	StatusBar  lipgloss.Style
}

// NewStyles builds a Styles from the config's color settings.
func NewStyles(cfg config.ColorConfig) Styles {
	c := colorVal
	return Styles{
		AddLine:    lipgloss.NewStyle().Foreground(c(cfg.AddFg)),
		RemoveLine: lipgloss.NewStyle().Foreground(c(cfg.RemoveFg)),
		HunkHeader: lipgloss.NewStyle().Foreground(c(cfg.HunkFg)),
		FileHeader: lipgloss.NewStyle().Foreground(c(cfg.FileHeader)).Bold(true),
		LineNum:    lipgloss.NewStyle().Foreground(c(cfg.LineNum)),
		CursorLine: lipgloss.NewStyle().Background(c(cfg.CursorBg)),
		CommentFg:  lipgloss.NewStyle().Foreground(c(cfg.CommentFg)),
		DraftFg:    lipgloss.NewStyle().Foreground(c(cfg.DraftFg)),
		StatusBar:  lipgloss.NewStyle().Foreground(c(cfg.StatusFg)).Background(c(cfg.StatusBg)),
	}
}

func colorVal(cv config.ColorValue) color.Color {
	return lipgloss.Color(cv.String())
}
