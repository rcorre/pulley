// Package github provides a client for GitHub pull request operations.
package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

// PRClient provides GitHub PR operations.
type PRClient interface {
	// FindPRForBranch returns the PR number of the open PR with the given head branch.
	FindPRForBranch(owner, repo, branch string) (int, error)
	GetPR(owner, repo string, number int) (*PR, error)
	GetDiff(owner, repo string, number int) (string, error)
	GetComments(owner, repo string, number int) ([]ReviewComment, error)
	SubmitReview(owner, repo string, number int, event ReviewEvent, body string, comments []DraftComment) error
}

type ghClient struct {
	rest       *api.RESTClient
	diffClient *api.RESTClient
	graphql    *api.GraphQLClient
}

// NewClient creates a new PRClient using the system gh authentication.
func NewClient() (PRClient, error) {
	rest, err := api.DefaultRESTClient()
	if err != nil {
		return nil, fmt.Errorf("create REST client: %w", err)
	}
	diffClient, err := api.NewRESTClient(api.ClientOptions{
		Headers: map[string]string{"Accept": "application/vnd.github.v3.diff"},
	})
	if err != nil {
		return nil, fmt.Errorf("create diff client: %w", err)
	}
	gql, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("create GraphQL client: %w", err)
	}
	return &ghClient{rest: rest, diffClient: diffClient, graphql: gql}, nil
}

func (c *ghClient) FindPRForBranch(owner, repo, branch string) (int, error) {
	slog.Debug("github: FindPRForBranch", "owner", owner, "repo", repo, "branch", branch)
	start := time.Now()
	var prs []struct {
		Number int `json:"number"`
	}
	path := fmt.Sprintf("repos/%s/%s/pulls?head=%s:%s&state=open", owner, repo, owner, branch)
	if err := c.rest.Get(path, &prs); err != nil {
		return 0, fmt.Errorf("list PRs for branch %q: %w", branch, err)
	}
	if len(prs) == 0 {
		return 0, fmt.Errorf("no open PR for branch %q", branch)
	}
	slog.Debug("github: FindPRForBranch ok", "duration", time.Since(start))
	return prs[0].Number, nil
}

func (c *ghClient) GetPR(owner, repo string, number int) (*PR, error) {
	slog.Debug("github: GetPR", "owner", owner, "repo", repo, "number", number)
	start := time.Now()
	var result struct {
		Repository struct {
			PullRequest struct {
				Number      int
				Title       string
				Author      struct{ Login string }
				Body        string
				BaseRefName string
				HeadRefName string
				URL         string
				State       string
			}
		}
	}
	err := c.graphql.Do(`
		query($owner: String!, $repo: String!, $number: Int!) {
			repository(owner: $owner, name: $repo) {
				pullRequest(number: $number) {
					number
					title
					author { login }
					body
					baseRefName
					headRefName
					url
					state
				}
			}
		}
	`, map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}, &result)
	if err != nil {
		return nil, fmt.Errorf("fetch PR #%d: %w", number, err)
	}
	pr := result.Repository.PullRequest
	slog.Debug("github: GetPR ok", "duration", time.Since(start))
	return &PR{
		Number:  pr.Number,
		Title:   pr.Title,
		Author:  pr.Author.Login,
		Body:    pr.Body,
		BaseRef: pr.BaseRefName,
		HeadRef: pr.HeadRefName,
		Owner:   owner,
		Repo:    repo,
		URL:     pr.URL,
		State:   PRState(pr.State),
	}, nil
}

func (c *ghClient) GetDiff(owner, repo string, number int) (string, error) {
	slog.Debug("github: GetDiff", "owner", owner, "repo", repo, "number", number)
	start := time.Now()
	path := fmt.Sprintf("repos/%s/%s/pulls/%d", owner, repo, number)
	resp, err := c.diffClient.Request("GET", path, nil)
	if err != nil {
		return "", fmt.Errorf("fetch diff for PR #%d: %w", number, err)
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read diff for PR #%d: %w", number, err)
	}
	slog.Debug("github: GetDiff ok", "duration", time.Since(start), "bytes", len(b))
	return string(b), nil
}

func (c *ghClient) GetComments(owner, repo string, number int) ([]ReviewComment, error) {
	slog.Debug("github: GetComments", "owner", owner, "repo", repo, "number", number)
	start := time.Now()
	var result struct {
		Repository struct {
			PullRequest struct {
				ReviewThreads struct {
					Nodes []struct {
						Comments struct {
							Nodes []struct {
								DatabaseID       int64
								Path             string
								OriginalPosition int
								Body             string
								Author           struct{ Login string }
								CreatedAt        time.Time
							}
						}
					}
				}
			}
		}
	}
	err := c.graphql.Do(`
		query($owner: String!, $repo: String!, $number: Int!) {
			repository(owner: $owner, name: $repo) {
				pullRequest(number: $number) {
					reviewThreads(first: 100) {
						nodes {
							comments(first: 100) {
								nodes {
									databaseId
									path
									originalPosition
									body
									author { login }
									createdAt
								}
							}
						}
					}
				}
			}
		}
	`, map[string]interface{}{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}, &result)
	if err != nil {
		return nil, fmt.Errorf("fetch comments for PR #%d: %w", number, err)
	}

	var comments []ReviewComment
	for _, thread := range result.Repository.PullRequest.ReviewThreads.Nodes {
		for _, n := range thread.Comments.Nodes {
			comments = append(comments, ReviewComment{
				ID:        n.DatabaseID,
				Path:      n.Path,
				Position:  n.OriginalPosition,
				Body:      n.Body,
				Author:    n.Author.Login,
				CreatedAt: n.CreatedAt,
			})
			slog.Debug("github: comment", "path", n.Path, "position", n.OriginalPosition, "author", n.Author.Login)
		}
	}
	slog.Debug("github: GetComments ok", "duration", time.Since(start), "comments", len(comments))
	return comments, nil
}

func (c *ghClient) SubmitReview(owner, repo string, number int, event ReviewEvent, body string, comments []DraftComment) error {
	slog.Debug("github: SubmitReview", "owner", owner, "repo", repo, "number", number, "event", event, "comments", len(comments))
	start := time.Now()
	type reviewRequest struct {
		Body     string         `json:"body"`
		Event    ReviewEvent    `json:"event"`
		Comments []DraftComment `json:"comments"`
	}

	if comments == nil {
		comments = []DraftComment{}
	}
	reqBody, err := json.Marshal(reviewRequest{Body: body, Event: event, Comments: comments})
	if err != nil {
		return fmt.Errorf("marshal review: %w", err)
	}

	path := fmt.Sprintf("repos/%s/%s/pulls/%d/reviews", owner, repo, number)
	err = c.rest.Post(path, bytes.NewReader(reqBody), nil)
	if err == nil {
		slog.Debug("github: SubmitReview ok", "duration", time.Since(start))
	}
	return err
}
