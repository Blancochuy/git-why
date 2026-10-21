package enricher

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseGitHubRemote(t *testing.T) {
	tests := []struct {
		raw   string
		owner string
		repo  string
		ok    bool
	}{
		{"https://github.com/Blancochuy/git-why.git", "Blancochuy", "git-why", true},
		{"https://github.com/Blancochuy/git-why/", "Blancochuy", "git-why", true},
		{"git@github.com:Blancochuy/git-why.git", "Blancochuy", "git-why", true},
		{"https://user:token@github.com/octo/repo.git", "octo", "repo", true},
		{"ssh://git@github.com/octo/repo.git", "octo", "repo", true},
		{"https://gitlab.com/octo/repo.git", "", "", false},
		{"not a url", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			owner, repo, err := ParseGitHubRemote(tt.raw)
			if !tt.ok {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.owner, owner)
			assert.Equal(t, tt.repo, repo)
		})
	}
}

func TestExtractIssueNumbers(t *testing.T) {
	client := &GitHubClient{}
	numbers := client.ExtractIssueNumbers("Fixes #12 and closes #12, refs #4")
	assert.Equal(t, []int{12, 4}, numbers)
}
