package review

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/github"
	"github.com/stretchr/testify/assert"
)

func testConfig() Config {
	return Config{
		Up:      key.NewBinding(key.WithKeys("up")),
		Down:    key.NewBinding(key.WithKeys("down")),
		Confirm: key.NewBinding(key.WithKeys("enter")),
		Cancel:  key.NewBinding(key.WithKeys("esc")),
	}
}

func keyUp() tea.KeyMsg    { return tea.KeyMsg{Type: tea.KeyUp} }
func keyDown() tea.KeyMsg  { return tea.KeyMsg{Type: tea.KeyDown} }
func keyEnter() tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyEnter} }
func keyEsc() tea.KeyMsg   { return tea.KeyMsg{Type: tea.KeyEsc} }

func TestCursorStartsAtZero(t *testing.T) {
	m := New(testConfig())
	m.Open(&github.PR{Number: 1}, nil)
	assert.Equal(t, 0, m.cursor)
}

func TestCursorDown(t *testing.T) {
	m := New(testConfig())
	m.Open(&github.PR{Number: 1}, nil)

	m, _ = m.Update(keyDown())
	assert.Equal(t, 1, m.cursor)

	m, _ = m.Update(keyDown())
	assert.Equal(t, 2, m.cursor)

	// clamps at last choice
	m, _ = m.Update(keyDown())
	assert.Equal(t, 2, m.cursor)
}

func TestCursorUp(t *testing.T) {
	m := New(testConfig())
	m.Open(&github.PR{Number: 1}, nil)

	// up at top does nothing
	m, _ = m.Update(keyUp())
	assert.Equal(t, 0, m.cursor)

	m, _ = m.Update(keyDown())
	m, _ = m.Update(keyDown())
	m, _ = m.Update(keyUp())
	assert.Equal(t, 1, m.cursor)
}

func TestOpenResetsCursor(t *testing.T) {
	m := New(testConfig())
	m.Open(&github.PR{Number: 1}, nil)
	m, _ = m.Update(keyDown())
	m, _ = m.Update(keyDown())
	assert.Equal(t, 2, m.cursor)

	m.Open(&github.PR{Number: 2}, nil)
	assert.Equal(t, 0, m.cursor)
}

func TestConfirmReturnsCmd(t *testing.T) {
	m := New(testConfig())
	m.Open(&github.PR{Number: 1}, nil)

	_, cmd := m.Update(keyEnter())
	assert.NotNil(t, cmd, "confirm should return a cmd to open the editor")
}

func TestCancelReturnsCancelMsg(t *testing.T) {
	m := New(testConfig())
	m.Open(&github.PR{Number: 1}, nil)

	_, cmd := m.Update(keyEsc())
	msg := cmd()
	_, ok := msg.(CancelMsg)
	assert.True(t, ok)
}

func TestViewContainsChoices(t *testing.T) {
	m := New(testConfig())
	m.Open(&github.PR{Number: 1}, nil)
	v := m.View()

	assert.Contains(t, v, "Submit Review")
	assert.Contains(t, v, "Approve")
	assert.Contains(t, v, "Request Changes")
	assert.Contains(t, v, "Comment")
	assert.Contains(t, v, "enter: confirm")
	assert.Contains(t, v, "esc: cancel")
}

func TestViewCursorIndicator(t *testing.T) {
	m := New(testConfig())
	m.Open(&github.PR{Number: 1}, nil)

	// cursor on Approve (index 0)
	v := m.View()
	lines := strings.Split(v, "\n")
	hasApproveWithCursor := false
	for _, l := range lines {
		if strings.Contains(l, "> Approve") {
			hasApproveWithCursor = true
		}
	}
	assert.True(t, hasApproveWithCursor, "Approve should have cursor indicator")

	// move cursor to Request Changes
	m, _ = m.Update(keyDown())
	v = m.View()
	lines = strings.Split(v, "\n")
	hasRequestWithCursor := false
	for _, l := range lines {
		if strings.Contains(l, "> Request Changes") {
			hasRequestWithCursor = true
		}
	}
	assert.True(t, hasRequestWithCursor, "Request Changes should have cursor indicator after down")
}

func TestBuildTemplateWithPRAndDrafts(t *testing.T) {
	m := New(testConfig())
	pr := &github.PR{Number: 42, Title: "Add feature"}
	drafts := []github.DraftComment{
		{Path: "foo.go", Position: 5, Body: "looks good"},
		{Path: "bar.go", Position: 10, Body: "multiline\nbody here"},
	}
	m.Open(pr, drafts)
	tmpl := m.buildTemplate()

	assert.Contains(t, tmpl, "# Reviewing PR #42: Add feature")
	assert.Contains(t, tmpl, "# - foo.go (position 5): looks good")
	assert.Contains(t, tmpl, "# - bar.go (position 10): multiline") // only first line
	assert.NotContains(t, tmpl, "body here")                        // second line stripped
	assert.Contains(t, tmpl, "# Pending comments:")
	assert.True(t, strings.HasSuffix(tmpl, "\n\n"))
}

func TestBuildTemplateNoDrafts(t *testing.T) {
	m := New(testConfig())
	m.Open(&github.PR{Number: 1, Title: "Fix bug"}, nil)
	tmpl := m.buildTemplate()

	assert.NotContains(t, tmpl, "Pending comments")
	assert.True(t, strings.HasSuffix(tmpl, "\n\n"))
}

func TestBuildTemplateDraftsCopied(t *testing.T) {
	m := New(testConfig())
	drafts := []github.DraftComment{{Path: "a.go", Position: 1, Body: "x"}}
	m.Open(&github.PR{Number: 1}, drafts)

	// mutating original slice does not affect model's copy
	drafts[0].Body = "mutated"
	tmpl := m.buildTemplate()
	assert.Contains(t, tmpl, "x")
	assert.NotContains(t, tmpl, "mutated")
}

func TestParseBodyStripsHashLines(t *testing.T) {
	input := "# comment\nactual body\n# another comment\n"
	assert.Equal(t, "actual body", parseBody(input))
}

func TestParseBodyTrims(t *testing.T) {
	assert.Equal(t, "text", parseBody("\n\ntext\n\n"))
}

func TestParseBodyEmpty(t *testing.T) {
	assert.Empty(t, parseBody("# only comments\n"))
	assert.Empty(t, parseBody(""))
}

func TestParseBodyPreservesNonHashLines(t *testing.T) {
	assert.Equal(t, "line one\nline two", parseBody("line one\nline two\n"))
}

func TestFirstLine(t *testing.T) {
	assert.Equal(t, "first", firstLine("first\nsecond\nthird"))
	assert.Equal(t, "only", firstLine("only"))
	assert.Empty(t, firstLine(""))
	assert.Empty(t, firstLine("\nsecond"))
}
