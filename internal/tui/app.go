// Package tui provides the terminal user interface for pulley.
package tui

import (
	"fmt"

	"charm.land/lipgloss/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/github"
)

// Focus identifies which panel receives key input.
type Focus int

const (
	// FocusFileList indicates the file list panel is active.
	FocusFileList Focus = iota
	// FocusDiff indicates the diff view panel is active.
	FocusDiff
	// FocusReview indicates the review dialog is active.
	FocusReview
)

// Model is the root TUI model.
type Model struct {
	// Config and clients
	cfg    config.Config
	client github.PRClient
	prRef  *github.PRRef

	// Keymap and styles
	keymap Keymap
	styles Styles

	// State
	pr    *github.PR
	diff  string
	focus Focus

	// Child models
	statusBar StatusBar

	// Loading/error
	loading bool
	err     error

	// Dimensions
	width  int
	height int
}

// NewModel creates the root model. The Init method triggers data loading.
func NewModel(cfg config.Config, client github.PRClient, prRef *github.PRRef) Model {
	keymap := NewKeymap(cfg)
	styles := NewStyles(cfg)
	return Model{
		cfg:       cfg,
		client:    client,
		prRef:     prRef,
		keymap:    keymap,
		styles:    styles,
		statusBar: NewStatusBar(styles),
		loading:   true,
		focus:     FocusFileList,
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return LoadPR(m.client, m.prRef)
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.SetWidth(msg.Width)

	case tea.KeyMsg:
		if matches(m.keymap.Quit, msg) {
			return m, tea.Quit
		}

	case PRLoadedMsg:
		m.loading = false
		m.pr = msg.PR
		m.diff = msg.Diff
		m.statusBar.SetPR(&PRInfo{Number: msg.PR.Number, Title: msg.PR.Title})

	case PRFailedMsg:
		m.loading = false
		m.err = msg.Err
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error loading PR: %v\n", m.err)
	}
	if m.loading {
		return "Loading PR..."
	}
	if m.pr == nil {
		return ""
	}

	// Simple layout: status bar at bottom, content above
	content := fmt.Sprintf("PR #%d: %s\n\n%s", m.pr.Number, m.pr.Title, m.diff)

	statusBar := m.statusBar.View()

	// Calculate available height for content
	contentHeight := m.height - lipgloss.Height(statusBar)
	if contentHeight < 0 {
		contentHeight = 0
	}

	return lipgloss.JoinVertical(lipgloss.Top,
		lipgloss.NewStyle().Height(contentHeight).Render(content),
		statusBar,
	)
}
