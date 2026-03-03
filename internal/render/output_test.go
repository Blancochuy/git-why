package render

import (
	"testing"
	"time"

	"github.com/chuy/git-why/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestNewRenderer(t *testing.T) {
	plainRenderer := NewRenderer(true)
	assert.NotNil(t, plainRenderer)
	assert.True(t, plainRenderer.plain)
	assert.Equal(t, plainColors, plainRenderer.colors)

	colorRenderer := NewRenderer(false)
	assert.NotNil(t, colorRenderer)
	assert.False(t, colorRenderer.plain)
	assert.Equal(t, defaultColors, colorRenderer.colors)
}

func TestFormatShort(t *testing.T) {
	blame := &git.BlameResult{
		Hash:   "a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e",
		Author: "Carlos García",
	}

	commit := &git.CommitInfo{
		Hash:    "a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e",
		Author:  "Carlos García",
		Message: "feat: add JWT refresh token rotation\n\nDetailed body here",
		Date:    time.Date(2024, 8, 14, 12, 0, 0, 0, time.UTC),
	}

	result := FormatShort(blame, commit)
	expected := "a3f92c1 · 2024-08-14 · @Carlos García · feat: add JWT refresh token rotation"
	assert.Equal(t, expected, result)
}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "today",
			input:    time.Now().Add(-1 * time.Hour),
			expected: "today",
		},
		{
			name:     "yesterday",
			input:    time.Now().Add(-25 * time.Hour),
			expected: "yesterday",
		},
		{
			name:     "days ago",
			input:    time.Now().Add(-72 * time.Hour),
			expected: "3 days ago",
		},
		{
			name:     "weeks ago",
			input:    time.Now().Add(-168 * time.Hour),
			expected: "1 weeks ago",
		},
		{
			name:     "specific date for old commits",
			input:    time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
			expected: "2024-01-15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDate(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColors(t *testing.T) {
	assert.Equal(t, "\033[36m", defaultColors.Hash)
	assert.Equal(t, "\033[33m", defaultColors.Author)
	assert.Equal(t, "\033[32m", defaultColors.Date)
	assert.Equal(t, "\033[0m", defaultColors.Message)
	assert.Equal(t, "\033[35m", defaultColors.File)
	assert.Equal(t, "\033[1m", defaultColors.Header)
	assert.Equal(t, "\033[0m", defaultColors.Reset)

	assert.Equal(t, "", plainColors.Hash)
	assert.Equal(t, "", plainColors.Author)
	assert.Equal(t, "", plainColors.Date)
}
