package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/config"
)

// Keymap holds key bindings for the application.
type Keymap struct {
	Quit         []string
	Up           []string
	Down         []string
	PageUp       []string
	PageDown     []string
	Tab          []string
	Comment      []string
	Suggestion   []string
	SubmitReview []string
	NextFile     []string
	PrevFile     []string
	Confirm      []string
	Cancel       []string
}

// NewKeymap creates a Keymap from config.
func NewKeymap(cfg config.Config) Keymap {
	return Keymap{
		Quit:         cfg.Keys.Quit,
		Up:           cfg.Keys.Up,
		Down:         cfg.Keys.Down,
		PageUp:       cfg.Keys.PageUp,
		PageDown:     cfg.Keys.PageDown,
		Tab:          cfg.Keys.Tab,
		Comment:      cfg.Keys.Comment,
		Suggestion:   cfg.Keys.Suggestion,
		SubmitReview: cfg.Keys.SubmitReview,
		NextFile:     cfg.Keys.NextFile,
		PrevFile:     cfg.Keys.PrevFile,
		Confirm:      cfg.Keys.Confirm,
		Cancel:       cfg.Keys.Cancel,
	}
}

// matches checks if the given key matches any in the list.
func matches(keys []string, msg tea.KeyMsg) bool {
	for _, k := range keys {
		if msg.String() == k {
			return true
		}
	}
	return false
}
