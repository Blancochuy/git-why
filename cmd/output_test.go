package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/blancochuy/git-why/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestOutputJSON(t *testing.T) {
	blame := &git.BlameResult{
		Hash:       "a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e",
		Author:     "testuser",
		AuthorMail: "<test@example.com>",
	}

	commit := &git.CommitInfo{
		Hash:    "a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e",
		Author:  "testuser",
		Message: "feat: test message",
		Body:    "Test body",
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputJSON("test.go", 42, blame, commit)

	w.Close()
	os.Stdout = old

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)

	var result OutputJSON
	err = json.Unmarshal(buf.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "test.go", result.File)
	assert.Equal(t, 42, result.Line)
	assert.Equal(t, "a3f92c1", result.ShortHash)
	assert.Equal(t, "test@example.com", result.Email)
}

func TestOutputMarkdown(t *testing.T) {
	blame := &git.BlameResult{
		Hash:       "a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e",
		Author:     "testuser",
		AuthorMail: "<test@example.com>",
	}

	commit := &git.CommitInfo{
		Hash:    "a3f92c1a2b4c3d4e5f6a7b8c9d0e1f2a3b4c5d6e",
		Author:  "testuser",
		Message: "feat: test message",
		Body:    "Test body content",
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputMarkdown("test.go", 42, blame, commit)

	w.Close()
	os.Stdout = old

	assert.NoError(t, err)

	var buf bytes.Buffer
	buf.ReadFrom(r)

	output := buf.String()
	assert.Contains(t, output, "# test.go:42")
	assert.Contains(t, output, "testuser")
	assert.Contains(t, output, "feat: test message")
	assert.Contains(t, output, "Test body content")
}
