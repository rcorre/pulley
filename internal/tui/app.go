// Package tui provides the bubbletea root model for the pulley TUI.
package tui

import (
	"fmt"
	"sync"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/github"
	"github.com/rcorre/pulley/internal/tui/statusbar"
)

// Model is the root bubbletea model for pulley.
type Model struct {
	client    github.PRClient
	ref       github.PRRef
	keymap    Keymap
	styles    Styles
	statusbar statusbar.Model
	spinner   spinner.Model

	pr       *github.PR
	diffs    []diff.FileDiff
	comments []github.ReviewComment

	err    error
	width  int
	height int
}

// New creates the root Model. It does not start fetching until Init is called.
func New(client github.PRClient, ref github.PRRef, cfg config.Config) Model {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	return Model{
		client:    client,
		ref:       ref,
		keymap:    NewKeymap(cfg.Keys),
		styles:    NewStyles(cfg.Colors),
		statusbar: statusbar.New(cfg.Colors),
		spinner:   s,
	}
}

// Init implements tea.Model. It starts the spinner and kicks off async PR loading.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.loadPR())
}

// loadPR fetches PR metadata, diff, and comments concurrently.
func (m Model) loadPR() tea.Cmd {
	return func() tea.Msg {
		var (
			pr       *github.PR
			rawDiff  string
			comments []github.ReviewComment
			prErr    error
			diffErr  error
			commErr  error
		)

		var wg sync.WaitGroup
		wg.Add(3)
		go func() {
			defer wg.Done()
			pr, prErr = m.client.GetPR(m.ref.Owner, m.ref.Repo, m.ref.Number)
		}()
		go func() {
			defer wg.Done()
			rawDiff, diffErr = m.client.GetDiff(m.ref.Owner, m.ref.Repo, m.ref.Number)
		}()
		go func() {
			defer wg.Done()
			comments, commErr = m.client.GetComments(m.ref.Owner, m.ref.Repo, m.ref.Number)
		}()
		wg.Wait()

		for _, err := range []error{prErr, diffErr, commErr} {
			if err != nil {
				return ErrMsg{Err: err}
			}
		}

		diffs, err := diff.Parse(rawDiff)
		if err != nil {
			return ErrMsg{Err: fmt.Errorf("parse diff: %w", err)}
		}
		return PRLoadedMsg{PR: pr, Diffs: diffs, Comments: comments}
	}
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusbar.SetWidth(msg.Width)
		return m, nil

	case PRLoadedMsg:
		m.pr = msg.PR
		m.diffs = msg.Diffs
		m.comments = msg.Comments
		m.statusbar.SetPR(msg.PR)
		return m, nil

	case ErrMsg:
		m.err = msg.Err
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keymap.Quit) {
			return m, tea.Quit
		}
	}

	if m.pr == nil {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	if m.err != nil {
		return "Error: " + m.err.Error() + "\n"
	}

	statusBar := m.statusbar.View()

	if m.pr == nil {
		return m.spinner.View() + " Loading PR...\n" + statusBar
	}

	body := fmt.Sprintf("PR #%d: %s\n%d file(s) changed", m.pr.Number, m.pr.Title, len(m.diffs))
	return body + "\n" + statusBar
}
