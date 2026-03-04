package render

import (
	"fmt"
	"strings"

	"github.com/blancochuy/git-why/internal/git"
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

	if commit.PRURL != "" {
		sb.WriteString(fmt.Sprintf("\n--- GITHUB PR (%s) ---\n", commit.PRURL))
		if strings.TrimSpace(commit.PRBody) != "" {
			sb.WriteString(fmt.Sprintf("%s\n", strings.TrimSpace(commit.PRBody)))
		}
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

	diffText := strings.Join(cleaned, "\n")

	// Truncate to avoid AI context window blowup
	maxLines := 500
	maxChars := 15000

	if len(cleaned) > maxLines {
		truncated := strings.Join(cleaned[:maxLines], "\n")
		return truncated + "\n\n[... Diff truncated for AI context limits ...]"
	}

	if len(diffText) > maxChars {
		return diffText[:maxChars] + "\n\n[... Diff truncated for AI context limits ...]"
	}

	return diffText
}
