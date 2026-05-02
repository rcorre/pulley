package github

import (
	"fmt"
	"net/url"
	"os/exec"
	"strconv"
	"strings"

	"github.com/cli/go-gh/v2/pkg/repository"
)

// PRRef identifies a specific pull request.
type PRRef struct {
	Owner  string
	Repo   string
	Number int
}

// Resolve returns a PRRef from user input.
// arg can be: empty (uses current branch), a PR number, a GitHub URL, or a branch name.
func Resolve(client PRClient, arg string) (*PRRef, error) {
	r := &resolver{
		client:        client,
		currentRepo:   defaultCurrentRepo,
		currentBranch: defaultCurrentBranch,
	}
	return r.resolve(arg)
}

// resolver holds injectable dependencies so Resolve can be unit-tested.
type resolver struct {
	client        PRClient
	currentRepo   func() (owner, name string, err error)
	currentBranch func() (string, error)
}

func (r *resolver) resolve(arg string) (*PRRef, error) {
	if n, err := strconv.Atoi(arg); err == nil {
		owner, repo, err := r.currentRepo()
		if err != nil {
			return nil, fmt.Errorf("get current repo: %w", err)
		}
		return &PRRef{Owner: owner, Repo: repo, Number: n}, nil
	}

	if ref, err := parseURL(arg); err == nil {
		return ref, nil
	}

	branch := arg
	if branch == "" {
		var err error
		branch, err = r.currentBranch()
		if err != nil {
			return nil, fmt.Errorf("get current branch: %w", err)
		}
	}

	owner, repo, err := r.currentRepo()
	if err != nil {
		return nil, fmt.Errorf("get current repo: %w", err)
	}

	number, err := r.client.FindPRForBranch(owner, repo, branch)
	if err != nil {
		return nil, fmt.Errorf("find PR for branch %q: %w", branch, err)
	}

	return &PRRef{Owner: owner, Repo: repo, Number: number}, nil
}

// parseURL extracts a PRRef from a GitHub pull request URL of the form
// https://github.com/owner/repo/pull/123.
func parseURL(rawURL string) (*PRRef, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) != 4 || parts[2] != "pull" {
		return nil, fmt.Errorf("not a PR URL: %s", rawURL)
	}
	n, err := strconv.Atoi(parts[3])
	if err != nil {
		return nil, fmt.Errorf("invalid PR number in URL: %s", rawURL)
	}
	return &PRRef{Owner: parts[0], Repo: parts[1], Number: n}, nil
}

func defaultCurrentRepo() (string, string, error) {
	repo, err := repository.Current()
	if err != nil {
		return "", "", fmt.Errorf("get current repository: %w", err)
	}
	return repo.Owner, repo.Name, nil
}

func defaultCurrentBranch() (string, error) {
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
