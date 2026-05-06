package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/rcorre/pulley/internal/config"
)

// Keymap holds key bindings for all TUI actions.
type Keymap struct {
	Quit         key.Binding
	Up           key.Binding
	Down         key.Binding
	PageUp       key.Binding
	PageDown     key.Binding
	Comment      key.Binding
	Suggestion   key.Binding
	SubmitReview key.Binding
	NextFile     key.Binding
	PrevFile     key.Binding
	Confirm      key.Binding
	Cancel       key.Binding
	Retry        key.Binding
	Suspend      key.Binding
}

// NewKeymap builds a Keymap from the config's key settings.
func NewKeymap(cfg config.KeyConfig) Keymap {
	return Keymap{
		Quit:         key.NewBinding(key.WithKeys(cfg.Quit...)),
		Up:           key.NewBinding(key.WithKeys(cfg.Up...)),
		Down:         key.NewBinding(key.WithKeys(cfg.Down...)),
		PageUp:       key.NewBinding(key.WithKeys(cfg.PageUp...)),
		PageDown:     key.NewBinding(key.WithKeys(cfg.PageDown...)),
		Comment:      key.NewBinding(key.WithKeys(cfg.Comment...)),
		Suggestion:   key.NewBinding(key.WithKeys(cfg.Suggestion...)),
		SubmitReview: key.NewBinding(key.WithKeys(cfg.SubmitReview...)),
		NextFile:     key.NewBinding(key.WithKeys(cfg.NextFile...)),
		PrevFile:     key.NewBinding(key.WithKeys(cfg.PrevFile...)),
		Confirm:      key.NewBinding(key.WithKeys(cfg.Confirm...)),
		Cancel:       key.NewBinding(key.WithKeys(cfg.Cancel...)),
		Retry:        key.NewBinding(key.WithKeys(cfg.Retry...)),
		Suspend:      key.NewBinding(key.WithKeys(cfg.Suspend...)),
	}
}
