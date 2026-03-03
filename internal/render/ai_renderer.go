package render

import (
	"fmt"
	"strings"

	"github.com/chuy/git-why/internal/git"
)

type AIRenderer struct{}

func NewAIRenderer() *AIRenderer {
	return &AIRenderer{}
}

func (r *AIRenderer) RenderContext(file string, line int, blame *git.BlameResult, commit *git.CommitInfo) {
	fmt.Print(r.GetContext(file, line, blame, commit))
}

func (r *AIRenderer) GetContext(file string, line int, blame *git.BlameResult, commit *git.CommitInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("ID: %s | %s | @%s\n", blame.Hash[:12], commit.Date.Format("2006-01-02"), blame.Author))
	sb.WriteString(fmt.Sprintf("FILE: %s:%d\n", file, line))
	sb.WriteString(fmt.Sprintf("MSG: %s\n", strings.Split(commit.Message, "\n")[0]))

	if strings.TrimSpace(commit.Body) != "" {
		sb.WriteString(fmt.Sprintf("\n--- RATIONALE ---\n%s\n", strings.TrimSpace(commit.Body)))
	}

	if commit.Diff != "" {
		sb.WriteString("\n--- DIFF ---\n")
		sb.WriteString(r.cleanDiff(commit.Diff))
	}

	return sb.String()
}

func (r *AIRenderer) cleanDiff(diff string) string {
	var cleaned []string
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		// Skip noisy git headers
		if strings.HasPrefix(line, "diff --git") ||
			strings.HasPrefix(line, "index ") ||
			strings.HasPrefix(line, "--- ") ||
			strings.HasPrefix(line, "+++ ") {
			continue
		}
		cleaned = append(cleaned, line)
	}
	return strings.Join(cleaned, "\n")
}
