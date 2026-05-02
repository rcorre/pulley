package github

import "time"

// PR holds metadata for a GitHub pull request.
type PR struct {
	Number  int
	Title   string
	Author  string
	Body    string
	BaseRef string
	HeadRef string
	Owner   string
	Repo    string
	URL     string
	State   PRState
}

// ReviewComment is an existing review comment on a pull request.
type ReviewComment struct {
	ID        int64
	Path      string
	Position  int
	Body      string
	Author    string
	CreatedAt time.Time
}

// DraftComment is a comment drafted locally, not yet submitted.
type DraftComment struct {
	Path     string
	Position int
	Body     string
}

// PRState is the state of a pull request.
type PRState string

// PR state constants.
const (
	StateOpen   PRState = "OPEN"
	StateClosed PRState = "CLOSED"
	StateMerged PRState = "MERGED"
)

// ReviewEvent is the action taken when submitting a review.
type ReviewEvent string

// Review event constants for use with SubmitReview.
const (
	EventApprove        ReviewEvent = "APPROVE"
	EventRequestChanges ReviewEvent = "REQUEST_CHANGES"
	EventComment        ReviewEvent = "COMMENT"
)
