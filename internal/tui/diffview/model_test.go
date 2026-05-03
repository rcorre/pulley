package diffview

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/syntax"
	"github.com/stretchr/testify/assert"
)

func testConfig() Config {
	return Config{
		Up:       []string{"k"},
		Down:     []string{"j"},
		PageUp:   []string{"ctrl+u"},
		PageDown: []string{"ctrl+d"},
		NextHunk: []string{"]"},
		PrevHunk: []string{"["},
		HunkFg:   lipgloss.NewStyle(),
		LineNum:  lipgloss.NewStyle(),
		AddFg:    lipgloss.NewStyle(),
		AddBg:    lipgloss.NewStyle(),
		RemoveFg: lipgloss.NewStyle(),
		RemoveBg: lipgloss.NewStyle(),
		CursorBg: lipgloss.NewStyle(),
	}
}

// testFile returns a FileDiff with two hunks for use in tests.
func testFile() diff.FileDiff {
	return diff.FileDiff{
		NewName: "test.go",
		Hunks: []diff.Hunk{
			{
				Header: "@@ -1,3 +1,4 @@",
				Lines: []diff.Line{
					{Kind: diff.LineContext, Content: "foo", OldLine: 1, NewLine: 1},
					{Kind: diff.LineAdd, Content: "bar", NewLine: 2},
					{Kind: diff.LineRemove, Content: "baz", OldLine: 2},
					{Kind: diff.LineContext, Content: "qux", OldLine: 3, NewLine: 3},
				},
			},
			{
				Header: "@@ -10,2 +11,2 @@",
				Lines: []diff.Line{
					{Kind: diff.LineContext, Content: "alpha", OldLine: 10, NewLine: 11},
					{Kind: diff.LineContext, Content: "beta", OldLine: 11, NewLine: 12},
				},
			},
		},
	}
}

func newTestModel() Model {
	m := New(testConfig(), syntax.NewHighlighter(""))
	m.SetSize(80, 10)
	m.SetFile(testFile())
	return m
}

func pressKey(m Model, key string) Model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated
}

func TestSetFileRendersLines(t *testing.T) {
	m := newTestModel()
	// 2 hunks * 1 header + 4 lines in hunk1 + 2 lines in hunk2 = 8 total
	assert.Len(t, m.lines, 8)
	assert.Len(t, m.hunkRows, 2)
}

func TestSetFileResetsPosition(t *testing.T) {
	m := newTestModel()
	m.cursor = 5
	m.offset = 3

	m.SetFile(testFile())
	assert.Equal(t, 0, m.cursor)
	assert.Equal(t, 0, m.offset)
}

func TestRenderContainsMarkers(t *testing.T) {
	m := newTestModel()
	// lines[2] is the add line, lines[3] is the remove line.
	// Check the marker+space pattern; content may contain ANSI codes from syntax highlighting.
	assert.Contains(t, m.lines[2], " + ")
	assert.Contains(t, m.lines[3], " - ")
}

func TestRenderContainsLineNumbers(t *testing.T) {
	m := newTestModel()
	// Line numbers for the context lines should appear in the gutter
	joined := strings.Join(m.lines, "\n")
	assert.Contains(t, joined, "1")
	assert.Contains(t, joined, "10")
}

func TestRenderHunkHeader(t *testing.T) {
	m := newTestModel()
	// First line is the first hunk header
	assert.Contains(t, m.lines[0], "@@ -1,3 +1,4 @@")
}

func TestCursorMoveDown(t *testing.T) {
	m := newTestModel()
	assert.Equal(t, 0, m.cursor)

	m = pressKey(m, "j")
	assert.Equal(t, 1, m.cursor)

	m = pressKey(m, "j")
	assert.Equal(t, 2, m.cursor)
}

func TestCursorMoveUp(t *testing.T) {
	m := newTestModel()
	m.cursor = 3

	m = pressKey(m, "k")
	assert.Equal(t, 2, m.cursor)
}

func TestCursorClampAtStart(t *testing.T) {
	m := newTestModel()
	assert.Equal(t, 0, m.cursor)

	m = pressKey(m, "k")
	assert.Equal(t, 0, m.cursor)
}

func TestCursorClampAtEnd(t *testing.T) {
	m := newTestModel()
	m.cursor = len(m.lines) - 1

	m = pressKey(m, "j")
	assert.Equal(t, len(m.lines)-1, m.cursor)
}

func TestPageDown(t *testing.T) {
	m := newTestModel()
	m.SetSize(80, 3)

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlD})
	m = updated
	assert.Equal(t, 3, m.cursor)
}

func TestPageUp(t *testing.T) {
	m := newTestModel()
	m.SetSize(80, 3)
	m.cursor = 5
	m.offset = 3

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlU})
	m = updated
	assert.Equal(t, 2, m.cursor)
}

func TestHunkJumpNext(t *testing.T) {
	m := newTestModel()
	// Cursor at 0 (first hunk header). Jump to next hunk.
	m = pressKey(m, "]")
	assert.Equal(t, m.hunkRows[1], m.cursor)
}

func TestHunkJumpNextWraps(t *testing.T) {
	m := newTestModel()
	m.cursor = m.hunkRows[1] // at last hunk

	m = pressKey(m, "]")
	assert.Equal(t, m.hunkRows[0], m.cursor)
}

func TestHunkJumpPrev(t *testing.T) {
	m := newTestModel()
	m.cursor = m.hunkRows[1] // at second hunk header

	m = pressKey(m, "[")
	assert.Equal(t, m.hunkRows[0], m.cursor)
}

func TestHunkJumpPrevWraps(t *testing.T) {
	m := newTestModel()
	m.cursor = m.hunkRows[0] // at first hunk header

	m = pressKey(m, "[")
	assert.Equal(t, m.hunkRows[1], m.cursor)
}

func TestViewHeightClamped(t *testing.T) {
	m := newTestModel()
	m.SetSize(80, 4)

	view := m.View()
	lines := strings.Split(view, "\n")
	assert.Len(t, lines, 4)
}

func TestViewScrollsWithCursor(t *testing.T) {
	m := newTestModel()
	m.SetSize(80, 3)
	// Move cursor past the visible area
	for range 5 {
		m = pressKey(m, "j")
	}
	// Offset should have advanced to keep cursor visible
	assert.Equal(t, 5, m.cursor)
	assert.GreaterOrEqual(t, m.offset, 3)
}
