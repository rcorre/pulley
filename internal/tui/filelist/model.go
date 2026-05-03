// Package filelist provides the left-panel file list component for the pulley TUI.
package filelist

import (
	"log/slog"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/diff"
)

// FileSelectedMsg is sent when the cursor moves to a new file.
type FileSelectedMsg struct {
	Index int
	File  diff.FileDiff
}

type fileStyles struct {
	added      lipgloss.Style
	deleted    lipgloss.Style
	modified   lipgloss.Style
	cursorBg   string         // ANSI index or hex, passed to lipgloss.Color
	cursorLine lipgloss.Style // cursorBg background + MaxWidth, updated in SetSize
}

type keymap struct {
	up   key.Binding
	down key.Binding
}

// Model is the left-panel file list showing all files changed in the PR.
type Model struct {
	files  []diff.FileDiff
	cursor int
	styles fileStyles
	keys   keymap
}

// New creates a Model from the given config.
func New(cfg config.Config) Model {
	c := cfg.Colors
	cursorBg := c.CursorBg.String()
	return Model{
		styles: fileStyles{
			added:      lipgloss.NewStyle().Foreground(lipgloss.Color(c.AddFg.String())),
			deleted:    lipgloss.NewStyle().Foreground(lipgloss.Color(c.RemoveFg.String())),
			modified:   lipgloss.NewStyle().Foreground(lipgloss.Color(c.FileModFg.String())),
			cursorBg:   cursorBg,
			cursorLine: lipgloss.NewStyle().Background(lipgloss.Color(cursorBg)),
		},
		keys: keymap{
			up:   key.NewBinding(key.WithKeys(cfg.Keys.Up...)),
			down: key.NewBinding(key.WithKeys(cfg.Keys.Down...)),
		},
	}
}

// SetFiles populates the list and resets the cursor to 0.
func (m *Model) SetFiles(files []diff.FileDiff) {
	slog.Debug("filelist: set files", "count", len(files))
	m.files = files
	m.cursor = 0
}

// SetSize sets the panel dimensions.
func (m *Model) SetSize(width, _ int) {
	base := lipgloss.NewStyle().Background(lipgloss.Color(m.styles.cursorBg))
	if width > 0 {
		m.styles.cursorLine = base.MaxWidth(width)
	} else {
		m.styles.cursorLine = base
	}
}

// SelectedIndex returns the current cursor position.
func (m Model) SelectedIndex() int { return m.cursor }

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles key input when the file list has focus.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok || len(m.files) == 0 {
		return m, nil
	}
	switch {
	case key.Matches(keyMsg, m.keys.up):
		if m.cursor > 0 {
			m.cursor--
			slog.Debug("filelist: cursor", "index", m.cursor, "file", m.files[m.cursor].Name())
			return m, m.selectCmd()
		}
	case key.Matches(keyMsg, m.keys.down):
		if m.cursor < len(m.files)-1 {
			m.cursor++
			slog.Debug("filelist: cursor", "index", m.cursor, "file", m.files[m.cursor].Name())
			return m, m.selectCmd()
		}
	}
	return m, nil
}

func (m Model) selectCmd() tea.Cmd {
	f := m.files[m.cursor]
	return func() tea.Msg { return FileSelectedMsg{Index: m.cursor, File: f} }
}

// View renders the file list panel.
func (m Model) View() string {
	if len(m.files) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, f := range m.files {
		status, statusStyle := m.fileStatus(f)
		line := statusStyle.Render(status) + " " + f.Name()
		if i == m.cursor {
			line = m.styles.cursorLine.Render(line)
		}
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(line)
	}
	return sb.String()
}

func (m Model) fileStatus(f diff.FileDiff) (string, lipgloss.Style) {
	switch {
	case f.IsNew:
		return "A", m.styles.added
	case f.IsDelete:
		return "D", m.styles.deleted
	default:
		return "M", m.styles.modified
	}
}
