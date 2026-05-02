package tui

import (
	"testing"

	"github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/github"
	"github.com/stretchr/testify/assert"
)

// mockClient is a simple mock for testing.
type mockClient struct {
	pr       *github.PR
	diff     string
	comments []github.ReviewComment
	err      error
}

func (m *mockClient) FindPRForBranch(_, _ string, _ string) (int, error) {
	return 1, nil
}

func (m *mockClient) GetPR(_, _ string, _ int) (*github.PR, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pr, nil
}

func (m *mockClient) GetDiff(_, _ string, _ int) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.diff, nil
}

func (m *mockClient) GetComments(_, _ string, _ int) ([]github.ReviewComment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.comments, nil
}

func (m *mockClient) SubmitReview(_, _ string, _ int, _ github.ReviewEvent, _ string, _ []github.DraftComment) error {
	return m.err
}

func TestModelInit(t *testing.T) {
	pr := &github.PR{Number: 1, Title: "Test PR"}
	mc := &mockClient{pr: pr, diff: "mock diff", comments: nil}
	ref := &github.PRRef{Owner: "o", Repo: "r", Number: 1}
	cfg := config.Default()
	model := NewModel(cfg, mc, ref)

	cmd := model.Init()
	assert.NotNil(t, cmd)
}

func TestModelLoadingState(t *testing.T) {
	pr := &github.PR{Number: 1, Title: "Test PR"}
	mc := &mockClient{pr: pr, diff: "mock diff", comments: nil}
	ref := &github.PRRef{Owner: "o", Repo: "r", Number: 1}
	cfg := config.Default()
	model := NewModel(cfg, mc, ref)

	assert.True(t, model.loading)
	assert.Nil(t, model.pr)

	// Simulate WindowSizeMsg
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	model = newModel.(Model)
	assert.Equal(t, 80, model.width)
	assert.Equal(t, 24, model.height)
}

func TestModelPRLoadedMsg(t *testing.T) {
	mc := &mockClient{}
	ref := &github.PRRef{Owner: "o", Repo: "r", Number: 1}
	cfg := config.Default()
	model := NewModel(cfg, mc, ref)

	pr := &github.PR{Number: 42, Title: "Awesome Feature"}
	msg := PRLoadedMsg{PR: pr, Diff: "diff content", Comments: nil}

	newModel, _ := model.Update(msg)
	model = newModel.(Model)

	assert.False(t, model.loading)
	assert.Equal(t, 42, model.pr.Number)
	assert.Equal(t, "Awesome Feature", model.pr.Title)
	assert.Equal(t, "diff content", model.diff)
	assert.NotNil(t, model.statusBar.pr)
}

func TestModelQuit(t *testing.T) {
	mc := &mockClient{}
	ref := &github.PRRef{Owner: "o", Repo: "r", Number: 1}
	cfg := config.Default()
	cfg.Keys.Quit = []string{"q"}
	model := NewModel(cfg, mc, ref)

	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	assert.NotNil(t, cmd)

	// Check that the command is tea.Quit
	msg := cmd()
	assert.Equal(t, tea.Quit(), msg)
}

func TestStatusBarView(t *testing.T) {
	cfg := config.Default()
	styles := NewStyles(cfg)
	sb := NewStatusBar(styles)

	sb.SetPR(&PRInfo{Number: 1, Title: "Test"})
	view := sb.View()
	assert.Contains(t, view, "#1:")
	assert.Contains(t, view, "Test")

	sb.SetDraftCount(3)
	view = sb.View()
	assert.Contains(t, view, "[3 draft(s)]")
}
