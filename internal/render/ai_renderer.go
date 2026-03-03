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
	fmt.Printf("FILE: %s:%d\n", file, line)
	fmt.Printf("COMMIT: %s\n", blame.Hash)
	fmt.Printf("AUTHOR: %s\n", blame.Author)
	fmt.Printf("DATE: %s\n", commit.Date.Format("2006-01-02"))
	fmt.Printf("SUMMARY: %s\n", strings.Split(commit.Message, "\n")[0])

	if strings.TrimSpace(commit.Body) != "" {
		fmt.Printf("\n--- RATIONALE ---\n%s\n", strings.TrimSpace(commit.Body))
	}

	if commit.Diff != "" {
		fmt.Printf("\n--- DIFF ---\n%s\n", strings.TrimSpace(commit.Diff))
	}

	fmt.Println("\n--- END CONTEXT ---")
}
