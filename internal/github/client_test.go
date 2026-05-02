package github

import (
	"testing"

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
		currentRepo: func() (string, string, error) {
			return "myorg", "myrepo", nil
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
		name    string
		url     string
		want    *PRRef
		wantErr bool
	}{
		{
			name: "valid URL",
			url:  "https://github.com/owner/repo/pull/123",
			want: &PRRef{Owner: "owner", Repo: "repo", Number: 123},
		},
		{
			name:    "repo URL without PR path",
			url:     "https://github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "issues URL",
			url:     "https://github.com/owner/repo/issues/1",
			wantErr: true,
		},
		{
			name:    "non-numeric PR number",
			url:     "https://github.com/owner/repo/pull/abc",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseURL(tt.url)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
