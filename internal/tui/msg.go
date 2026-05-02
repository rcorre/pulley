package tui

import (
	"github.com/rcorre/pulley/internal/diff"
	"github.com/rcorre/pulley/internal/github"
)

// PRLoadedMsg is sent when PR data has been fetched successfully.
type PRLoadedMsg struct {
	PR       *github.PR
	Diffs    []diff.FileDiff
	Comments []github.ReviewComment
}

// ErrMsg is sent when an error occurs during an async operation.
type ErrMsg struct {
	Err error
}
