package tui_test

import (
	"bytes"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/rcorre/pulley/internal/config"
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/github"
	"github.com/rcorre/pulley/internal/tui"
	"github.com/rcorre/pulley/internal/tui/comment"
	"github.com/rcorre/pulley/internal/tui/review"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubClient provides no-op implementations of the rarely-needed PRClient methods.
type stubClient struct{}

func (stubClient) FindPRForBranch(_, _, _ string) (int, error) { return 0, nil }
func (stubClient) SubmitReview(_ string, _ string, _ int, _ github.ReviewEvent, _ string, _ []github.DraftComment) error {
	return nil
}

// mockClient returns pre-canned data for all PR fetch operations.
type mockClient struct {
	stubClient
	pr       *github.PR
	rawDiff  string
	comments []github.ReviewComment
}

func (c *mockClient) GetPR(_, _ string, _ int) (*github.PR, error) { return c.pr, nil }
func (c *mockClient) GetDiff(_, _ string, _ int) (string, error)   { return c.rawDiff, nil }
func (c *mockClient) GetComments(_, _ string, _ int) ([]github.ReviewComment, error) {
	return c.comments, nil
}

// blockingClient blocks all fetch operations until done is closed.
type blockingClient struct {
	stubClient
	done <-chan struct{}
}

func (c *blockingClient) GetPR(_, _ string, _ int) (*github.PR, error) {
	<-c.done
	return nil, nil
}
func (c *blockingClient) GetDiff(_, _ string, _ int) (string, error) {
	<-c.done
	return "", nil
}
func (c *blockingClient) GetComments(_, _ string, _ int) ([]github.ReviewComment, error) {
	<-c.done
	return nil, nil
}

func newTestModel(client github.PRClient) tui.Model {
	ref := github.PRRef{Owner: "owner", Repo: "repo", Number: 42}
	return tui.New(client, ref, config.Default())
}

func TestLoadingState(t *testing.T) {
	blocking := make(chan struct{})
	m := newTestModel(&blockingClient{done: blocking})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		close(blocking) // unblock loadPR goroutines so they can exit
		if err := tm.Quit(); err != nil {
			t.Logf("quit: %v", err)
		}
	})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Loading"))
	}, teatest.WithDuration(3*time.Second))
}

func TestPRLoadedMsg(t *testing.T) {
	pr := &github.PR{Number: 42, Title: "Add feature"}
	rawDiff := `diff --git a/foo.go b/foo.go
index 0000000..1111111 100644
--- a/foo.go
+++ b/foo.go
@@ -1,1 +1,2 @@
 package main
+// added
`
	diffs, err := diff.Parse(rawDiff)
	if err != nil {
		t.Fatalf("parse diff: %v", err)
	}

	m := newTestModel(&mockClient{pr: pr, rawDiff: rawDiff})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Logf("quit: %v", err)
		}
	})

	tm.Send(tui.PRLoadedMsg{PR: pr, Diffs: diffs, Comments: nil})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return strings.Contains(string(bts), "Add feature")
	}, teatest.WithDuration(3*time.Second))
}

func TestErrMsg(t *testing.T) {
	m := newTestModel(&blockingClient{done: make(chan struct{})})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Logf("quit: %v", err)
		}
	})

	tm.Send(tui.ErrMsg{Err: errors.New("network failure")})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("network failure"))
	}, teatest.WithDuration(3*time.Second))
}

// reviewCapture wraps mockClient to capture SubmitReview calls.
type reviewCapture struct {
	mockClient
	mu        sync.Mutex
	submitErr error
	submitted bool
	event     github.ReviewEvent
	body      string
	comments  []github.DraftComment
}

func (c *reviewCapture) SubmitReview(_, _ string, _ int, event github.ReviewEvent, body string, comments []github.DraftComment) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.submitted = true
	c.event = event
	c.body = body
	c.comments = comments
	return c.submitErr
}

func TestSubmitReviewSuccess(t *testing.T) {
	pr := &github.PR{Number: 42, Title: "Add feature", Owner: "owner", Repo: "repo"}
	rawDiff := `diff --git a/foo.go b/foo.go
index 0000000..1111111 100644
--- a/foo.go
+++ b/foo.go
@@ -1,1 +1,2 @@
 package main
+// added
`
	diffs, err := diff.Parse(rawDiff)
	require.NoError(t, err)

	client := &reviewCapture{
		mockClient: mockClient{pr: pr, rawDiff: rawDiff},
	}
	m := newTestModel(client)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Logf("quit: %v", err)
		}
	})

	draft := github.DraftComment{Path: "foo.go", Position: 1, Body: "nice change"}
	tm.Send(tui.PRLoadedMsg{PR: pr, Diffs: diffs, Comments: nil})
	tm.Send(comment.DraftAddedMsg{Draft: draft})
	tm.Send(review.SubmitMsg{Event: github.EventApprove, Body: "LGTM"})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Review submitted!"))
	}, teatest.WithDuration(3*time.Second))

	client.mu.Lock()
	defer client.mu.Unlock()
	assert.True(t, client.submitted)
	assert.Equal(t, github.EventApprove, client.event)
	assert.Equal(t, "LGTM", client.body)
	require.Len(t, client.comments, 1)
	assert.Equal(t, draft, client.comments[0])
}

func TestSubmitReviewError(t *testing.T) {
	pr := &github.PR{Number: 1, Title: "Fix bug", Owner: "owner", Repo: "repo"}
	client := &reviewCapture{
		mockClient: mockClient{pr: pr, rawDiff: ""},
		submitErr:  errors.New("API failure"),
	}
	m := newTestModel(client)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Logf("quit: %v", err)
		}
	})

	tm.Send(tui.PRLoadedMsg{PR: pr, Diffs: nil, Comments: nil})
	tm.Send(review.SubmitMsg{Event: github.EventComment, Body: ""})

	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("API failure"))
	}, teatest.WithDuration(3*time.Second))
}

func TestQuit(t *testing.T) {
	m := newTestModel(&mockClient{
		pr:      &github.PR{Number: 1, Title: "Test"},
		rawDiff: "",
	})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(3*time.Second))
}

func TestRetry(t *testing.T) {
	blocking := make(chan struct{})
	m := newTestModel(&blockingClient{done: blocking})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		close(blocking)
		if err := tm.Quit(); err != nil {
			t.Logf("quit: %v", err)
		}
	})

	tm.Send(tui.ErrMsg{Err: errors.New("network failure")})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("network failure")) && bytes.Contains(bts, []byte("retry"))
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Loading"))
	}, teatest.WithDuration(3*time.Second))
}

func TestWindowResize(t *testing.T) {
	pr := &github.PR{Number: 42, Title: "Resize test"}
	rawDiff := `diff --git a/foo.go b/foo.go
index 0000000..1111111 100644
--- a/foo.go
+++ b/foo.go
@@ -1,1 +1,2 @@
 package main
+// added
`
	diffs, err := diff.Parse(rawDiff)
	require.NoError(t, err)

	m := newTestModel(&mockClient{pr: pr, rawDiff: rawDiff})
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(80, 24))
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Logf("quit: %v", err)
		}
	})

	tm.Send(tui.PRLoadedMsg{PR: pr, Diffs: diffs, Comments: nil})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Resize test"))
	}, teatest.WithDuration(3*time.Second))

	tm.Send(tea.WindowSizeMsg{Width: 120, Height: 40})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Resize test"))
	}, teatest.WithDuration(3*time.Second))
}

func TestEndToEnd(t *testing.T) {
	pr := &github.PR{Number: 42, Title: "End to end", Owner: "owner", Repo: "repo"}
	rawDiff := `diff --git a/foo.go b/foo.go
index 0000000..1111111 100644
--- a/foo.go
+++ b/foo.go
@@ -1,1 +1,2 @@
 package main
+// added
diff --git a/bar.go b/bar.go
index 0000000..2222222 100644
--- a/bar.go
+++ b/bar.go
@@ -1,1 +1,2 @@
 package main
+// bar
`
	diffs, err := diff.Parse(rawDiff)
	require.NoError(t, err)

	client := &reviewCapture{
		mockClient: mockClient{pr: pr, rawDiff: rawDiff},
	}
	m := newTestModel(client)
	tm := teatest.NewTestModel(t, m, teatest.WithInitialTermSize(120, 40))
	t.Cleanup(func() {
		if err := tm.Quit(); err != nil {
			t.Logf("quit: %v", err)
		}
	})

	// Load PR data.
	tm.Send(tui.PRLoadedMsg{PR: pr, Diffs: diffs, Comments: nil})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("End to end"))
	}, teatest.WithDuration(3*time.Second))

	// Navigate to second file in the file list.
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("bar.go"))
	}, teatest.WithDuration(3*time.Second))

	// Add a draft comment.
	draft := github.DraftComment{Path: "bar.go", Position: 1, Body: "looks good"}
	tm.Send(comment.DraftAddedMsg{Draft: draft})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("1 draft"))
	}, teatest.WithDuration(3*time.Second))

	// Submit the review directly (bypasses $EDITOR).
	tm.Send(review.SubmitMsg{Event: github.EventApprove, Body: "LGTM"})
	teatest.WaitFor(t, tm.Output(), func(bts []byte) bool {
		return bytes.Contains(bts, []byte("Review submitted!"))
	}, teatest.WithDuration(3*time.Second))

	client.mu.Lock()
	defer client.mu.Unlock()
	assert.True(t, client.submitted)
	assert.Equal(t, github.EventApprove, client.event)
	require.Len(t, client.comments, 1)
	assert.Equal(t, draft, client.comments[0])
}
