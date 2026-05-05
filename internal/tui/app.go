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
	"github.com/rcorre/pulley/internal/tui/comment"
	"github.com/rcorre/pulley/internal/tui/diffview"
	"github.com/rcorre/pulley/internal/tui/filelist"
	"github.com/rcorre/pulley/internal/tui/review"
	"github.com/rcorre/pulley/internal/tui/statusbar"
)

// Focus identifies which panel has keyboard focus.
type Focus int

// Focus constants for panels that can receive keyboard input.
// Tab cycles only FocusFileList and FocusDiff; FocusReview is entered/exited explicitly.
const (
	FocusFileList Focus = iota
	FocusDiff
	FocusReview
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
	review    review.Model
	focus     Focus

	leftPanelStyle  lipgloss.Style
	rightPanelStyle lipgloss.Style

	pr          *github.PR
	diffs       []diff.FileDiff
	comments    []github.ReviewComment
	currentFile *diff.FileDiff
	drafts      []github.DraftComment

	err    error
	width  int
	height int
}

// New creates the root Model. It does not start fetching until Init is called.
func New(client github.PRClient, ref github.PRRef, cfg config.Config) Model {
	s := spinner.New(spinner.WithSpinner(spinner.Dot))
	km := NewKeymap(cfg.Keys)
	return Model{
		client:    client,
		ref:       ref,
		keymap:    km,
		styles:    NewStyles(cfg.Colors),
		statusbar: statusbar.New(cfg.Colors),
		spinner:   s,
		filelist:  filelist.New(cfg),
		diffview:  diffview.New(newDiffViewConfig(cfg), syntax.NewHighlighter("")),
		review:    review.New(newReviewConfig(km, cfg.Colors.CursorBg)),
		focus:     FocusFileList,
	}
}

func newReviewConfig(km Keymap, cursorBg config.ColorValue) review.Config {
	return review.Config{
		Up:      km.Up,
		Down:    km.Down,
		Confirm: km.Confirm,
		Cancel:  km.Cancel,
		Cursor:  bgStyle(cursorBg),
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
		DraftFg:   fgStyle(c.DraftFg),
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
			m.currentFile = &m.diffs[0]
			m.diffview.SetFile(f, m.diffComments(f.Name()))
		}
		return m, nil

	case filelist.FileSelectedMsg:
		slog.Debug("file selected", "file", msg.File.Name())
		m.currentFile = &m.diffs[msg.Index]
		m.diffview.SetFile(msg.File, m.diffComments(msg.File.Name()))
		return m, nil

	case comment.DraftAddedMsg:
		slog.Info("draft added", "path", msg.Draft.Path, "position", msg.Draft.Position)
		m.drafts = append(m.drafts, msg.Draft)
		m.statusbar.SetDraftCount(len(m.drafts))
		m.rerenderDiff()
		return m, nil

	case ErrMsg:
		slog.Error("PR load failed", "err", msg.Err)
		m.err = msg.Err
		return m, nil

	case review.CancelMsg:
		m.focus = FocusDiff
		return m, nil

	case review.SubmitMsg:
		slog.Info("submitting review", "event", msg.Event)
		return m, m.submitReview(msg.Event, msg.Body)

	case ReviewSubmittedMsg:
		m.focus = FocusDiff
		if msg.Err != nil {
			slog.Error("review submission failed", "err", msg.Err)
			m.statusbar.SetMessage("Error: " + msg.Err.Error())
		} else {
			slog.Info("review submitted")
			m.drafts = nil
			m.statusbar.SetDraftCount(0)
			m.statusbar.SetMessage("Review submitted!")
			m.rerenderDiff()
		}
		return m, nil

	case tea.KeyMsg:
		m.statusbar.SetMessage("")
		if key.Matches(msg, m.keymap.Quit) {
			return m, tea.Quit
		}
		if m.pr != nil {
			slog.Debug("key", "key", msg.String(), "focus", m.focus)
			if key.Matches(msg, m.keymap.Tab) && m.focus != FocusReview {
				m.focus = (m.focus + 1) % FocusReview
				return m, nil
			}
			if m.focus == FocusFileList {
				var cmd tea.Cmd
				m.filelist, cmd = m.filelist.Update(msg)
				return m, cmd
			}
			if m.focus == FocusDiff {
				if key.Matches(msg, m.keymap.Comment) {
					return m, m.openEditor(false)
				}
				if key.Matches(msg, m.keymap.Suggestion) {
					return m, m.openEditor(true)
				}
				if key.Matches(msg, m.keymap.SubmitReview) {
					m.review.Open(m.pr, m.drafts)
					m.focus = FocusReview
					return m, nil
				}
				var cmd tea.Cmd
				m.diffview, cmd = m.diffview.Update(msg)
				return m, cmd
			}
			if m.focus == FocusReview {
				var cmd tea.Cmd
				m.review, cmd = m.review.Update(msg)
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

	if m.focus == FocusReview {
		content := lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, m.review.View())
		return content + "\n" + statusBar
	}

	leftPanel := m.leftPanelStyle.Render(m.filelist.View())
	rightPanel := m.rightPanelStyle.Render(m.diffview.View())
	return lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel) + "\n" + statusBar
}

func (m Model) submitReview(event github.ReviewEvent, body string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.SubmitReview(m.ref.Owner, m.ref.Repo, m.ref.Number, event, body, m.drafts)
		return ReviewSubmittedMsg{Err: err}
	}
}

func (m Model) openEditor(suggestion bool) tea.Cmd {
	if m.currentFile == nil {
		return nil
	}
	line, ok := m.diffview.CursorDiffLine()
	if !ok {
		return nil
	}
	return comment.Open(m.currentFile.Name(), line, suggestion)
}

func (m *Model) rerenderDiff() {
	if m.currentFile == nil {
		return
	}
	m.diffview.SetFile(*m.currentFile, m.diffComments(m.currentFile.Name()))
}

func (m Model) diffComments(path string) []diffview.Comment {
	slog.Debug("diffComments", "target", path, "total", len(m.comments))
	var result []diffview.Comment
	for _, c := range m.comments {
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
	for _, d := range m.drafts {
		if d.Path == path {
			result = append(result, diffview.Comment{
				Body:     d.Body,
				Position: d.Position,
				Draft:    true,
			})
		}
	}
	slog.Debug("diffComments done", "target", path, "matched", len(result))
	return result
}
