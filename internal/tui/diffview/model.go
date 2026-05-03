// Package diffview provides a scrollable diff viewport with syntax highlighting.
package diffview

import (
	"log/slog"
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/syntax"
)

// Config holds styles and key bindings for the diffview.
type Config struct {
	AddFg     lipgloss.Style
	AddBg     lipgloss.Style
	RemoveFg  lipgloss.Style
	RemoveBg  lipgloss.Style
	HunkFg    lipgloss.Style
	LineNum   lipgloss.Style
	CursorBg  lipgloss.Style
	CommentFg lipgloss.Style
	CommentBg lipgloss.Style
	DraftFg   lipgloss.Style
	Up        []string
	Down      []string
	PageUp    []string
	PageDown  []string
	NextHunk  []string
	PrevHunk  []string
}

// Model is the scrollable diff viewport with cursor line tracking.
type Model struct {
	cfg      Config
	hlr      *syntax.Highlighter
	lines    []string
	hunkRows []int
	lineMap  []diff.Line // parallel to lines; DiffPosition==0 means non-diff row
	cursor   int
	offset   int
	width    int
	height   int
	plain    lipgloss.Style
	cursorSt lipgloss.Style
}

// New creates a new Model.
func New(cfg Config, hlr *syntax.Highlighter) Model {
	return Model{cfg: cfg, hlr: hlr}
}

// SetSize sets the viewport dimensions and adjusts scroll to keep cursor visible.
func (m *Model) SetSize(w, h int) {
	slog.Debug("diffview: size", "w", w, "h", h)
	m.width = w
	m.height = h
	m.plain = lipgloss.NewStyle().MaxWidth(w)
	m.cursorSt = m.cfg.CursorBg.MaxWidth(w)
	m.clampScroll()
}

// SetFile renders the given FileDiff with inline comments and resets cursor to the top.
// comments should be pre-filtered to only include those for this file.
func (m *Model) SetFile(f diff.FileDiff, comments []Comment) {
	m.lines, m.hunkRows, m.lineMap = Render(f, m.cfg, m.hlr, comments)
	slog.Debug("diffview: set file", "file", f.Name(), "lines", len(m.lines), "comments", len(comments))
	m.cursor = 0
	m.offset = 0
}

// CursorDiffLine returns the diff.Line at or above the cursor position.
// It searches backwards from the cursor to skip over hunk header and comment rows.
// Returns false if no diff line is found (e.g. empty file).
func (m Model) CursorDiffLine() (diff.Line, bool) {
	for i := m.cursor; i >= 0; i-- {
		if i < len(m.lineMap) && m.lineMap[i].DiffPosition > 0 {
			return m.lineMap[i], true
		}
	}
	return diff.Line{}, false
}

// Update processes key messages for cursor movement and hunk navigation.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case matches(m.cfg.Down, keyMsg):
		m.moveCursor(1)
	case matches(m.cfg.Up, keyMsg):
		m.moveCursor(-1)
	case matches(m.cfg.PageDown, keyMsg):
		m.moveCursor(m.height)
	case matches(m.cfg.PageUp, keyMsg):
		m.moveCursor(-m.height)
	case matches(m.cfg.NextHunk, keyMsg):
		m.jumpHunk(1)
	case matches(m.cfg.PrevHunk, keyMsg):
		m.jumpHunk(-1)
	}
	return m, nil
}

// View renders the visible portion of the diff with cursor highlighting.
func (m Model) View() string {
	if m.width == 0 || len(m.lines) == 0 {
		return ""
	}

	end := min(m.offset+m.height, len(m.lines))

	rendered := make([]string, 0, end-m.offset)
	for i := m.offset; i < end; i++ {
		if i == m.cursor {
			rendered = append(rendered, m.cursorSt.Render(m.lines[i]))
		} else {
			rendered = append(rendered, m.plain.Render(m.lines[i]))
		}
	}

	return strings.Join(rendered, "\n")
}

// CursorLine returns the 0-based line index of the cursor within the rendered lines.
func (m Model) CursorLine() int {
	return m.cursor
}

func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if len(m.lines) > 0 && m.cursor >= len(m.lines) {
		m.cursor = len(m.lines) - 1
	}
	m.clampScroll()
}

func (m *Model) jumpHunk(dir int) {
	if len(m.hunkRows) == 0 {
		return
	}
	if dir > 0 {
		m.cursor = nextHunk(m.hunkRows, m.cursor)
	} else {
		m.cursor = prevHunk(m.hunkRows, m.cursor)
	}
	m.clampScroll()
}

func nextHunk(rows []int, cursor int) int {
	for _, row := range rows {
		if row > cursor {
			return row
		}
	}
	return rows[0]
}

func prevHunk(rows []int, cursor int) int {
	for i := len(rows) - 1; i >= 0; i-- {
		if rows[i] < cursor {
			return rows[i]
		}
	}
	return rows[len(rows)-1]
}

func (m *Model) clampScroll() {
	if m.height <= 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.height {
		m.offset = m.cursor - m.height + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
	maxOffset := max(0, len(m.lines)-m.height)
	if m.offset > maxOffset {
		m.offset = maxOffset
	}
}

func matches(keys []string, msg tea.KeyMsg) bool {
	return slices.Contains(keys, msg.String())
}
