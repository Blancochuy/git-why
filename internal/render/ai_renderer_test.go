package render

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/chuy/git-why/internal/git"
)

func TestAIRenderer_RenderContext(t *testing.T) {
	renderer := NewAIRenderer()

	blame := &git.BlameResult{
		Hash:   "abcdef1234567890",
		Author: "testauthor",
	}

	commit := &git.CommitInfo{
		Hash:    "abcdef1234567890",
		Author:  "testauthor",
		Date:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Message: "feat: test feature\n\nThis is a test body.",
		Body:    "This is a test body.",
		Diff:    "diff --git a/test.txt b/test.txt\n+added line",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	renderer.RenderContext("test.go", 10, blame, commit)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedSubstrings := []string{
		"ID: abcdef123456 | 2024-01-01 | @testauthor",
		"FILE: test.go:10",
		"MSG: feat: test feature",
		"--- RATIONALE ---",
		"This is a test body.",
		"--- DIFF ---",
		"+added line",
	}

	for _, sub := range expectedSubstrings {
		if !strings.Contains(output, sub) {
			t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", sub, output)
		}
	}

	// Verify noise is gone
	noisyHeader := "diff --git"
	if strings.Contains(output, noisyHeader) {
		t.Errorf("Output still contains noisy header %q, but it should be cleaned.", noisyHeader)
	}
}
