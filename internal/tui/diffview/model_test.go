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
	m.SetFile(testFile(), nil)
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

	m.SetFile(testFile(), nil)
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

func TestRenderInlineComments(t *testing.T) {
	f := diff.FileDiff{
		NewName: "test.go",
		Hunks: []diff.Hunk{
			{
				Header: "@@ -1,2 +1,2 @@",
				Lines: []diff.Line{
					{Kind: diff.LineContext, Content: "foo", OldLine: 1, NewLine: 1, DiffPosition: 1},
					{Kind: diff.LineAdd, Content: "bar", NewLine: 2, DiffPosition: 2},
				},
			},
		},
	}
	comments := []Comment{
		{Author: "alice", Body: "looks good", Position: 2},
	}
	m := New(testConfig(), syntax.NewHighlighter(""))
	m.SetSize(80, 20)
	m.SetFile(f, comments)

	joined := strings.Join(m.lines, "\n")
	assert.Contains(t, joined, "@ alice")
	assert.Contains(t, joined, "looks good")
	// comment appears after the add line (position 2), not after the context line (position 1)
	addIdx := -1
	commentIdx := -1
	for i, l := range m.lines {
		if strings.Contains(l, " + ") {
			addIdx = i
		}
		if strings.Contains(l, "@ alice") {
			commentIdx = i
		}
	}
	assert.Greater(t, commentIdx, addIdx, "comment should appear after its target line")
}

func TestCursorDiffLineOnHunkHeader(t *testing.T) {
	// testFile has no DiffPosition set, so cursor on hunk header (row 0) returns false
	m := newTestModel()
	assert.Equal(t, 0, m.cursor) // starts on hunk header
	_, ok := m.CursorDiffLine()
	assert.False(t, ok)
}

func TestCursorDiffLineOnDiffLine(t *testing.T) {
	// layout: [0] hunk header, [1] foo (DiffPosition 1), [2] bar (DiffPosition 2)
	f := diff.FileDiff{
		NewName: "test.go",
		Hunks: []diff.Hunk{{
			Header: "@@ -1,2 +1,2 @@",
			Lines: []diff.Line{
				{Kind: diff.LineContext, Content: "foo", OldLine: 1, NewLine: 1, DiffPosition: 1},
				{Kind: diff.LineAdd, Content: "bar", NewLine: 2, DiffPosition: 2},
			},
		}},
	}
	m := New(testConfig(), syntax.NewHighlighter(""))
	m.SetSize(80, 20)
	m.SetFile(f, nil)

	m.cursor = 0 // hunk header row
	_, ok := m.CursorDiffLine()
	assert.False(t, ok)

	m.cursor = 1 // first diff line (foo)
	line, ok := m.CursorDiffLine()
	assert.True(t, ok)
	assert.Equal(t, 1, line.DiffPosition)

	m.cursor = 2 // second diff line (bar)
	line, ok = m.CursorDiffLine()
	assert.True(t, ok)
	assert.Equal(t, 2, line.DiffPosition)
}

func TestCursorDiffLineOnCommentFallsBack(t *testing.T) {
	f := diff.FileDiff{
		NewName: "test.go",
		Hunks: []diff.Hunk{{
			Header: "@@ -1,1 +1,1 @@",
			Lines: []diff.Line{
				{Kind: diff.LineContext, Content: "foo", OldLine: 1, NewLine: 1, DiffPosition: 1},
			},
		}},
	}
	m := New(testConfig(), syntax.NewHighlighter(""))
	m.SetSize(80, 20)
	m.SetFile(f, []Comment{{Author: "alice", Body: "note", Position: 1}})

	// lines: [0] hunk header, [1] diff line, [2] comment header "@ alice", [3] comment body "  note"
	m.cursor = 3 // on comment body row
	line, ok := m.CursorDiffLine()
	assert.True(t, ok)
	assert.Equal(t, 1, line.DiffPosition)
}

func TestRenderDraftComment(t *testing.T) {
	f := diff.FileDiff{
		NewName: "test.go",
		Hunks: []diff.Hunk{{
			Header: "@@ -1,1 +1,1 @@",
			Lines: []diff.Line{
				{Kind: diff.LineContext, Content: "foo", OldLine: 1, NewLine: 1, DiffPosition: 1},
			},
		}},
	}
	m := New(testConfig(), syntax.NewHighlighter(""))
	m.SetSize(80, 20)
	m.SetFile(f, []Comment{
		{Author: "alice", Body: "existing", Position: 1, Draft: false},
		{Body: "my draft", Position: 1, Draft: true},
	})

	joined := strings.Join(m.lines, "\n")
	assert.Contains(t, joined, "@ alice")
	assert.Contains(t, joined, "[draft]")
}

func TestRenderMultilineComment(t *testing.T) {
	f := diff.FileDiff{
		NewName: "test.go",
		Hunks: []diff.Hunk{
			{
				Header: "@@ -1,1 +1,1 @@",
				Lines: []diff.Line{
					{Kind: diff.LineContext, Content: "foo", OldLine: 1, NewLine: 1, DiffPosition: 1},
				},
			},
		},
	}
	comments := []Comment{
		{Author: "bob", Body: "line one\nline two", Position: 1},
	}
	m := New(testConfig(), syntax.NewHighlighter(""))
	m.SetSize(80, 20)
	m.SetFile(f, comments)

	joined := strings.Join(m.lines, "\n")
	assert.Contains(t, joined, "line one")
	assert.Contains(t, joined, "line two")
}
