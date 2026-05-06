// Package filelist provides the left-panel file list component for the pulley TUI.
package filelist

import (
	"log/slog"
	"strings"

	"charm.land/lipgloss/v2"
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

// Model is the left-panel file list showing all files changed in the PR.
type Model struct {
	files  []diff.FileDiff
	cursor int
	styles fileStyles
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

// NextFile advances the cursor to the next file, wrapping around to the first.
func (m *Model) NextFile() tea.Cmd { return m.moveFile(1) }

// PrevFile moves the cursor to the previous file, wrapping around to the last.
func (m *Model) PrevFile() tea.Cmd { return m.moveFile(-1) }

func (m *Model) moveFile(delta int) tea.Cmd {
	if len(m.files) == 0 {
		return nil
	}
	m.cursor = (m.cursor + delta + len(m.files)) % len(m.files)
	slog.Debug("filelist: cursor", "index", m.cursor, "file", m.files[m.cursor].Name())
	return m.selectCmd()
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
