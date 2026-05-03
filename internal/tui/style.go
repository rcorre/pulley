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
}

// NewStyles builds a Styles from the config's color settings.
func NewStyles(cfg config.ColorConfig) Styles {
	return Styles{
		AddLine:    lipgloss.NewStyle().Foreground(colorVal(cfg.AddFg)),
		RemoveLine: lipgloss.NewStyle().Foreground(colorVal(cfg.RemoveFg)),
		HunkHeader: lipgloss.NewStyle().Foreground(colorVal(cfg.HunkFg)),
		FileHeader: lipgloss.NewStyle().Foreground(colorVal(cfg.FileHeader)).Bold(true),
		LineNum:    lipgloss.NewStyle().Foreground(colorVal(cfg.LineNum)),
		CursorLine: lipgloss.NewStyle().Background(colorVal(cfg.CursorBg)),
		CommentFg:  lipgloss.NewStyle().Foreground(colorVal(cfg.CommentFg)),
		DraftFg:    lipgloss.NewStyle().Foreground(colorVal(cfg.DraftFg)),
	}
}

func colorVal(cv config.ColorValue) color.Color {
	return lipgloss.Color(cv.String())
}

func fgStyle(cv config.ColorValue) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(colorVal(cv))
}

func bgStyle(cv config.ColorValue) lipgloss.Style {
	return lipgloss.NewStyle().Background(colorVal(cv))
}
