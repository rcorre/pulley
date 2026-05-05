// Package review provides the review submission dialog for the pulley TUI.
package review

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/github"
)

// SubmitMsg is sent when the user confirms a review action and body from the editor.
type SubmitMsg struct {
	Event github.ReviewEvent
	Body  string
}

// CancelMsg is sent when the user cancels the review dialog.
type CancelMsg struct{}

// Config holds key bindings and styles for the review dialog.
type Config struct {
	Up      key.Binding
	Down    key.Binding
	Confirm key.Binding
	Cancel  key.Binding
	Cursor  lipgloss.Style
}

type choice struct {
	label string
	event github.ReviewEvent
}

var reviewChoices = []choice{
	{"Approve", github.EventApprove},
	{"Request Changes", github.EventRequestChanges},
	{"Comment", github.EventComment},
}

// Model is the review submission dialog (action selector + $EDITOR spawner).
type Model struct {
	cursor int
	config Config
	border lipgloss.Style
	drafts []github.DraftComment
	pr     *github.PR
}

// New creates a Model from the given config.
func New(cfg Config) Model {
	return Model{
		config: cfg,
		border: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 3),
	}
}

// Open sets the PR and draft context and resets the cursor to the first choice.
func (m *Model) Open(pr *github.PR, drafts []github.DraftComment) {
	m.pr = pr
	m.drafts = make([]github.DraftComment, len(drafts))
	copy(m.drafts, drafts)
	m.cursor = 0
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update handles key input for the review dialog.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(keyMsg, m.config.Up):
		if m.cursor > 0 {
			m.cursor--
		}
	case key.Matches(keyMsg, m.config.Down):
		if m.cursor < len(reviewChoices)-1 {
			m.cursor++
		}
	case key.Matches(keyMsg, m.config.Confirm):
		return m, m.openEditor()
	case key.Matches(keyMsg, m.config.Cancel):
		return m, func() tea.Msg { return CancelMsg{} }
	}
	return m, nil
}

func (m Model) openEditor() tea.Cmd {
	tmp, err := os.CreateTemp("", "pulley-review-*.md")
	if err != nil {
		return func() tea.Msg { return CancelMsg{} }
	}

	event := reviewChoices[m.cursor].event
	_, err = tmp.WriteString(m.buildTemplate())
	if closeErr := tmp.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(tmp.Name())
		return func() tea.Msg { return CancelMsg{} }
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmpName := tmp.Name()
	return tea.ExecProcess(exec.Command(editor, tmpName), func(err error) tea.Msg { //nolint:gosec
		defer func() { _ = os.Remove(tmpName) }()
		if err != nil {
			return CancelMsg{}
		}
		content, err := os.ReadFile(tmpName) //nolint:gosec
		if err != nil {
			return CancelMsg{}
		}
		body := parseBody(string(content))
		return SubmitMsg{Event: event, Body: body}
	})
}

func (m Model) buildTemplate() string {
	var b strings.Builder
	if m.pr != nil {
		fmt.Fprintf(&b, "# Reviewing PR #%d: %s\n", m.pr.Number, m.pr.Title)
	}
	b.WriteString("# Lines starting with '#' will be ignored.\n")
	if len(m.drafts) > 0 {
		b.WriteString("#\n# Pending comments:\n")
		for _, d := range m.drafts {
			fmt.Fprintf(&b, "# - %s (position %d): %s\n", d.Path, d.Position, firstLine(d.Body))
		}
	}
	b.WriteString("#\n\n")
	return b.String()
}

func parseBody(content string) string {
	var lines []string
	for line := range strings.SplitSeq(content, "\n") {
		if !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func firstLine(s string) string {
	line, _, _ := strings.Cut(s, "\n")
	return line
}

// View renders the review action selector as a bordered dialog box.
func (m Model) View() string {
	var sb strings.Builder
	sb.WriteString("Submit Review\n\n")
	for i, c := range reviewChoices {
		if i == m.cursor {
			sb.WriteString(m.config.Cursor.Render("> " + c.label))
		} else {
			sb.WriteString("  " + c.label)
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("\nenter: confirm  esc: cancel")
	return m.border.Render(sb.String())
}
