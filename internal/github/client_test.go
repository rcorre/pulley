package github

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient implements PRClient for testing.
type mockClient struct {
	findPRForBranch func(owner, repo, branch string) (int, error)
	getPR           func(owner, repo string, number int) (*PR, error)
	getDiff         func(owner, repo string, number int) (string, error)
	getComments     func(owner, repo string, number int) ([]ReviewComment, error)
	submitReview    func(owner, repo string, number int, event ReviewEvent, body string, comments []DraftComment) error
}

func (m *mockClient) FindPRForBranch(owner, repo, branch string) (int, error) {
	return m.findPRForBranch(owner, repo, branch)
}

func (m *mockClient) GetPR(owner, repo string, number int) (*PR, error) {
	return m.getPR(owner, repo, number)
}

func (m *mockClient) GetDiff(owner, repo string, number int) (string, error) {
	return m.getDiff(owner, repo, number)
}

func (m *mockClient) GetComments(owner, repo string, number int) ([]ReviewComment, error) {
	return m.getComments(owner, repo, number)
}

func (m *mockClient) SubmitReview(owner, repo string, number int, event ReviewEvent, body string, comments []DraftComment) error {
	return m.submitReview(owner, repo, number, event, body, comments)
}

func newTestResolver(client PRClient) *resolver {
	return &resolver{
		client: client,
		currentRepo: func() (string, string, string, error) {
			return "github.com", "myorg", "myrepo", nil
		},
		currentBranch: func() (string, error) {
			return "current-branch", nil
		},
	}
}

func TestResolveNumber(t *testing.T) {
	r := newTestResolver(&mockClient{})
	ref, err := r.resolve("42")
	require.NoError(t, err)
	assert.Equal(t, &PRRef{Owner: "myorg", Repo: "myrepo", Number: 42}, ref)
}

func TestResolveURL(t *testing.T) {
	r := newTestResolver(&mockClient{})
	ref, err := r.resolve("https://github.com/otherorg/otherrepo/pull/99")
	require.NoError(t, err)
	assert.Equal(t, &PRRef{Owner: "otherorg", Repo: "otherrepo", Number: 99}, ref)
}

func TestResolveBranch(t *testing.T) {
	client := &mockClient{
		findPRForBranch: func(owner, repo, branch string) (int, error) {
			assert.Equal(t, "myorg", owner)
			assert.Equal(t, "myrepo", repo)
			assert.Equal(t, "feature-branch", branch)
			return 7, nil
		},
	}
	r := newTestResolver(client)

	ref, err := r.resolve("feature-branch")
	require.NoError(t, err)
	assert.Equal(t, &PRRef{Owner: "myorg", Repo: "myrepo", Number: 7}, ref)
}

func TestResolveEmpty(t *testing.T) {
	client := &mockClient{
		findPRForBranch: func(owner, repo, branch string) (int, error) {
			assert.Equal(t, "myorg", owner)
			assert.Equal(t, "myrepo", repo)
			assert.Equal(t, "current-branch", branch)
			return 13, nil
		},
	}
	r := newTestResolver(client)

	ref, err := r.resolve("")
	require.NoError(t, err)
	assert.Equal(t, &PRRef{Owner: "myorg", Repo: "myrepo", Number: 13}, ref)
}

func TestParseURL(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		expectedHost string
		want         *PRRef
		wantErr      bool
	}{
		{
			name:         "valid github.com URL",
			url:          "https://github.com/owner/repo/pull/123",
			expectedHost: "github.com",
			want:         &PRRef{Owner: "owner", Repo: "repo", Number: 123},
		},
		{
			name:         "valid GitHub Enterprise URL",
			url:          "https://github.enterprise.com/owner/repo/pull/456",
			expectedHost: "github.enterprise.com",
			want:         &PRRef{Owner: "owner", Repo: "repo", Number: 456},
		},
		{
			name:         "host mismatch",
			url:          "https://github.com/owner/repo/pull/123",
			expectedHost: "github.enterprise.com",
			wantErr:      true,
		},
		{
			name:         "non-GitHub host",
			url:          "https://gitlab.com/owner/repo/pull/123",
			expectedHost: "github.com",
			wantErr:      true,
		},
		{
			name:         "repo URL without PR path",
			url:          "https://github.com/owner/repo",
			expectedHost: "github.com",
			wantErr:      true,
		},
		{
			name:         "issues URL",
			url:          "https://github.com/owner/repo/issues/1",
			expectedHost: "github.com",
			wantErr:      true,
		},
		{
			name:         "non-numeric PR number",
			url:          "https://github.com/owner/repo/pull/abc",
			expectedHost: "github.com",
			wantErr:      true,
		},
	}
		for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseURL(tt.url, tt.expectedHost)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetPR(t *testing.T) {
	expected := &PR{
		Number: 42, Title: "Test PR", Author: "testuser",
		Body: "PR body", BaseRef: "main", HeadRef: "feature",
		Owner: "myorg", Repo: "myrepo", URL: "https://github.com/myorg/myrepo/pull/42",
		State: StateOpen,
	}
	client := &mockClient{
		getPR: func(owner, repo string, number int) (*PR, error) {
			assert.Equal(t, "myorg", owner)
			assert.Equal(t, "myrepo", repo)
			assert.Equal(t, 42, number)
			return expected, nil
		},
	}
	pr, err := client.GetPR("myorg", "myrepo", 42)
	require.NoError(t, err)
	assert.Equal(t, expected, pr)
}

func TestGetDiff(t *testing.T) {
	expected := "diff --git a/file b/file\n..."
	client := &mockClient{
		getDiff: func(owner, repo string, number int) (string, error) {
			assert.Equal(t, "myorg", owner)
			assert.Equal(t, "myrepo", repo)
			assert.Equal(t, 42, number)
			return expected, nil
		},
	}
	diff, err := client.GetDiff("myorg", "myrepo", 42)
	require.NoError(t, err)
	assert.Equal(t, expected, diff)
}

func TestGetComments(t *testing.T) {
	expected := []ReviewComment{
		{ID: 1, Path: "file.go", Position: 5, Body: "looks good", Author: "reviewer", CreatedAt: testTime()},
	}
	client := &mockClient{
		getComments: func(owner, repo string, number int) ([]ReviewComment, error) {
			assert.Equal(t, "myorg", owner)
			assert.Equal(t, "myrepo", repo)
			assert.Equal(t, 42, number)
			return expected, nil
		},
	}
	comments, err := client.GetComments("myorg", "myrepo", 42)
	require.NoError(t, err)
	assert.Equal(t, expected, comments)
}

func TestSubmitReview(t *testing.T) {
	drafts := []DraftComment{
		{Path: "file.go", Position: 5, Body: "needs work"},
	}
	client := &mockClient{
		submitReview: func(owner, repo string, number int, event ReviewEvent, body string, comments []DraftComment) error {
			assert.Equal(t, "myorg", owner)
			assert.Equal(t, "myrepo", repo)
			assert.Equal(t, 42, number)
			assert.Equal(t, EventRequestChanges, event)
			assert.Equal(t, "LGTM overall", body)
			assert.Equal(t, drafts, comments)
			return nil
		},
	}
	err := client.SubmitReview("myorg", "myrepo", 42, EventRequestChanges, "LGTM overall", drafts)
	require.NoError(t, err)
}

func testTime() time.Time {
	t, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00Z")
	return t
}
