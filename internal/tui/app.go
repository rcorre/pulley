// Package tui provides the bubbletea root model for the pulley TUI.
package tui

import (
	"fmt"
	"log/slog"
	"sync"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/github"
	"github.com/rcorre/pulley/internal/syntax"
	"github.com/rcorre/pulley/internal/tui/diffview"
	"github.com/rcorre/pulley/internal/tui/filelist"
	"github.com/rcorre/pulley/internal/tui/statusbar"
)

// Focus identifies which panel has keyboard focus.
type Focus int

// Focus constants for the two panels that can receive keyboard input.
const (
	FocusFileList Focus = iota
	FocusDiff
	focusCount // sentinel: number of focusable panels
)

// Model is the root bubbletea model for pulley.
type Model struct {
	client    github.PRClient
	ref       github.PRRef
	keymap    Keymap
	styles    Styles
	statusbar statusbar.Model
	spinner   spinner.Model
	filelist  filelist.Model
	diffview  diffview.Model
	focus     Focus

	leftPanelStyle  lipgloss.Style
	rightPanelStyle lipgloss.Style

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
		filelist:  filelist.New(cfg),
		diffview:  diffview.New(newDiffViewConfig(cfg), syntax.NewHighlighter("")),
		focus:     FocusFileList,
	}
}

func newDiffViewConfig(cfg config.Config) diffview.Config {
	c := cfg.Colors
	k := cfg.Keys
	return diffview.Config{
		AddFg:     fgStyle(c.AddFg),
		AddBg:     bgStyle(c.AddBg),
		RemoveFg:  fgStyle(c.RemoveFg),
		RemoveBg:  bgStyle(c.RemoveBg),
		HunkFg:    fgStyle(c.HunkFg),
		LineNum:   fgStyle(c.LineNum),
		CursorBg:  bgStyle(c.CursorBg),
		CommentFg: fgStyle(c.CommentFg),
		CommentBg: bgStyle(c.CommentBg),
		Up:        k.Up,
		Down:      k.Down,
		PageUp:    k.PageUp,
		PageDown:  k.PageDown,
		NextHunk:  k.NextHunk,
		PrevHunk:  k.PrevHunk,
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
		fileListWidth := msg.Width / 4
		diffWidth := msg.Width - fileListWidth
		contentHeight := msg.Height - 1
		m.filelist.SetSize(fileListWidth, contentHeight)
		m.diffview.SetSize(diffWidth, contentHeight)
		m.leftPanelStyle = lipgloss.NewStyle().Width(fileListWidth).Height(contentHeight)
		m.rightPanelStyle = lipgloss.NewStyle().Width(diffWidth).Height(contentHeight)
		return m, nil

	case PRLoadedMsg:
		slog.Info("PR loaded", "title", msg.PR.Title, "files", len(msg.Diffs), "comments", len(msg.Comments))
		m.pr = msg.PR
		m.diffs = msg.Diffs
		m.comments = msg.Comments
		m.statusbar.SetPR(msg.PR)
		m.filelist.SetFiles(msg.Diffs)
		if len(msg.Diffs) > 0 {
			f := msg.Diffs[0]
			m.diffview.SetFile(f, fileComments(msg.Comments, f.Name()))
		}
		return m, nil

	case filelist.FileSelectedMsg:
		slog.Debug("file selected", "file", msg.File.Name())
		m.diffview.SetFile(msg.File, fileComments(m.comments, msg.File.Name()))
		return m, nil

	case ErrMsg:
		slog.Error("PR load failed", "err", msg.Err)
		m.err = msg.Err
		return m, nil

	case tea.KeyMsg:
		if key.Matches(msg, m.keymap.Quit) {
			return m, tea.Quit
		}
		if m.pr != nil {
			slog.Debug("key", "key", msg.String(), "focus", m.focus)
			if key.Matches(msg, m.keymap.Tab) {
				m.focus = (m.focus + 1) % focusCount
				return m, nil
			}
			if m.focus == FocusFileList {
				var cmd tea.Cmd
				m.filelist, cmd = m.filelist.Update(msg)
				return m, cmd
			}
			if m.focus == FocusDiff {
				var cmd tea.Cmd
				m.diffview, cmd = m.diffview.Update(msg)
				return m, cmd
			}
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

	if m.width == 0 || m.height == 0 {
		return statusBar
	}

	leftPanel := m.leftPanelStyle.Render(m.filelist.View())
	rightPanel := m.rightPanelStyle.Render(m.diffview.View())

	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel) + "\n" + statusBar
}

func fileComments(comments []github.ReviewComment, path string) []diffview.Comment {
	slog.Debug("fileComments", "target", path, "total", len(comments))
	var result []diffview.Comment
	for _, c := range comments {
		if c.Path == path {
			slog.Debug("comment match", "path", c.Path, "position", c.Position, "author", c.Author)
			result = append(result, diffview.Comment{
				Author:   c.Author,
				Body:     c.Body,
				Position: c.Position,
			})
		} else {
			slog.Debug("comment skip", "comment_path", c.Path, "target_path", path)
		}
	}
	slog.Debug("fileComments done", "target", path, "matched", len(result))
	return result
}
