package git

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseBlameOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *BlameResult
		wantErr  bool
	}{
		{
			name: "parse basic blame output",
			input: `a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e 42 42 1
author Carlos García
author-mail <carlos@example.com>
author-time 1723680000
author-tz +0000
committer GitHub
committer-mail <noreply@github.com>
committer-time 1723680000
summary feat: add JWT refresh token rotation
boundary
	fmt.Sprintf("token: %s", token)`,
			expected: &BlameResult{
				Hash:       "a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e",
				Author:     "Carlos García",
				AuthorMail: "<carlos@example.com>",
				Line:       `fmt.Sprintf("token: %s", token)`,
			},
			wantErr: false,
		},
		{
			name:    "empty input returns error",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseBlameOutput(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Hash, result.Hash)
			assert.Equal(t, tt.expected.Author, result.Author)
			assert.Equal(t, tt.expected.AuthorMail, result.AuthorMail)
			assert.Equal(t, tt.expected.Line, result.Line)
			assert.Equal(t, 42, result.OriginalLineNum)
			assert.Equal(t, 42, result.FinalLineNum)
			assert.True(t, result.Date.Equal(time.Unix(1723680000, 0)))
		})
	}
}

func TestParseCommitInfo(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *CommitInfo
		wantErr  bool
	}{
		{
			name: "parse full commit info",
			input: `a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e
Carlos García
carlos@example.com
1723680000
feat: add JWT refresh token rotation
This implements secure token rotation to prevent
session hijacking attacks.

Closes #89`,
			expected: &CommitInfo{
				Hash:        "a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e",
				Author:      "Carlos García",
				AuthorEmail: "carlos@example.com",
				Message:     "feat: add JWT refresh token rotation",
				Body: `This implements secure token rotation to prevent
session hijacking attacks.

Closes #89`,
			},
			wantErr: false,
		},
		{
			name:    "empty input returns error",
			input:   "",
			wantErr: true,
		},
		{
			name:    "insufficient lines returns error",
			input:   "hash\nauthor",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseCommitInfo(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Hash, result.Hash)
			assert.Equal(t, tt.expected.Author, result.Author)
			assert.Equal(t, tt.expected.AuthorEmail, result.AuthorEmail)
			assert.Equal(t, tt.expected.Message, result.Message)
			assert.Equal(t, tt.expected.Body, result.Body)
		})
	}
}

func TestParseLogOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		wantErr  bool
	}{
		{
			name: "parse multiple commits",
			input: `a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e|Carlos García|1723680000|feat: add JWT refresh token rotation
b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3|Ana Martínez|1723593600|fix: token validation edge case
c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4|José López|1723507200|refactor: extract token utils`,
			expected: 3,
			wantErr:  false,
		},
		{
			name:     "empty input returns empty slice",
			input:    "",
			expected: 0,
			wantErr:  false,
		},
		{
			name: "malformed lines are skipped",
			input: `a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e|Carlos García|1723680000|feat: valid commit
invalid line here
b4c5d6e7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3|Ana Martínez|1723593600|fix: another valid`,
			expected: 2,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLogOutput(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, result, tt.expected)
		})
	}
}

func TestIsUncommitted(t *testing.T) {
	assert.True(t, IsUncommitted(strings.Repeat("0", 40)))
	assert.False(t, IsUncommitted(""))
	assert.False(t, IsUncommitted("a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e"))
}

func TestFilterCommitsByTerm(t *testing.T) {
	input := strings.Join([]string{
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa\x1fAda\x1f1723680000\x1ffeat: add parser\x1fExplains blame output.\n\nSecond paragraph mentions rotation.\x1e",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb\x1fBea\x1f1723593600\x1ffix: tests\x1fNo match here.\x1e",
	}, "")

	commits, err := filterCommitsByTerm(input, "rotation", false)
	assert.NoError(t, err)
	assert.Len(t, commits, 1)
	assert.Equal(t, "Ada", commits[0].Author)
	assert.Contains(t, commits[0].Body, "Second paragraph")
	assert.True(t, commits[0].Date.Equal(time.Unix(1723680000, 0)))

	commits, err = filterCommitsByTerm(input, "fix.*tests", true)
	assert.NoError(t, err)
	assert.Len(t, commits, 1)
	assert.Equal(t, "Bea", commits[0].Author)

	_, err = filterCommitsByTerm(input, "(", true)
	assert.Error(t, err)
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected time.Time
		wantErr  bool
	}{
		{
			name:     "valid unix timestamp",
			input:    "1723680000",
			expected: time.Unix(1723680000, 0),
			wantErr:  false,
		},
		{
			name:    "invalid timestamp",
			input:   "invalid",
			wantErr: true,
		},
		{
			name:    "empty timestamp",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseTimestamp(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
