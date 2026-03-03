package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/chuy/git-why/internal/git"
	"github.com/chuy/git-why/internal/render"
)

type OutputJSON struct {
	File      string    `json:"file"`
	Line      int       `json:"line"`
	Hash      string    `json:"hash"`
	ShortHash string    `json:"shortHash"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Date      time.Time `json:"date"`
	Message   string    `json:"message"`
	Body      string    `json:"body,omitempty"`
	Summary   string    `json:"summary"`
}

type OutputJSONRange struct {
	File    string        `json:"file"`
	Range   string        `json:"range"`
	Count   int           `json:"count"`
	Commits []CommitEntry `json:"commits"`
}

type CommitEntry struct {
	Hash      string    `json:"hash"`
	ShortHash string    `json:"shortHash"`
	Author    string    `json:"author"`
	Date      time.Time `json:"date"`
	Message   string    `json:"message"`
}

func outputJSON(file string, line int, blame *git.BlameResult, commit *git.CommitInfo) error {
	output := OutputJSON{
		File:      file,
		Line:      line,
		Hash:      blame.Hash,
		ShortHash: blame.Hash[:7],
		Author:    blame.Author,
		Email:     strings.Trim(blame.AuthorMail, "<>"),
		Date:      commit.Date,
		Message:   commit.Message,
		Body:      strings.TrimSpace(commit.Body),
		Summary:   strings.Split(commit.Message, "\n")[0],
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func outputJSONRange(file string, start, end int, commits []git.CommitInfo) error {
	entries := make([]CommitEntry, len(commits))
	for i, c := range commits {
		entries[i] = CommitEntry{
			Hash:      c.Hash,
			ShortHash: c.Hash[:7],
			Author:    c.Author,
			Date:      c.Date,
			Message:   strings.Split(c.Message, "\n")[0],
		}
	}

	output := OutputJSONRange{
		File:    file,
		Range:   fmt.Sprintf("%d-%d", start, end),
		Count:   len(commits),
		Commits: entries,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func outputMarkdown(file string, line int, blame *git.BlameResult, commit *git.CommitInfo) error {
	fmt.Printf("# %s:%d\n\n", file, line)
	fmt.Printf("| Property | Value |\n")
	fmt.Printf("|----------|-------|\n")
	fmt.Printf("| **Hash** | `%s` |\n", blame.Hash[:12])
	fmt.Printf("| **Author** | @%s |\n", blame.Author)
	fmt.Printf("| **Date** | %s |\n", commit.Date.Format("2006-01-02"))
	fmt.Printf("\n---\n\n")
	fmt.Printf("## Commit Message\n\n%s\n\n", commit.Message)

	if commit.Body != "" && strings.TrimSpace(commit.Body) != "" {
		fmt.Printf("## Context\n\n```\n%s\n```\n", strings.TrimSpace(commit.Body))
	}

	return nil
}

func outputMarkdownRange(file string, start, end int, commits []git.CommitInfo) error {
	fmt.Printf("# %s:%d-%d\n\n", file, start, end)
	fmt.Printf("**%d commits** touched this range:\n\n", len(commits))

	for _, c := range commits {
		fmt.Printf("### %s\n", c.Hash[:7])
		fmt.Printf("- **Author:** %s\n", c.Author)
		fmt.Printf("- **Date:** %s\n", c.Date.Format("2006-01-02"))
		fmt.Printf("- **Message:** %s\n\n", strings.Split(c.Message, "\n")[0])
	}

	return nil
}

func renderResult(file string, line int, blame *git.BlameResult, commit *git.CommitInfo) {
	renderer := render.NewRenderer(plainOutput)
	renderer.RenderLine(file, line, blame, commit)
}
