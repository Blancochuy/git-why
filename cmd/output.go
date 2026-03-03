package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chuy/git-why/internal/git"
	"github.com/chuy/git-why/internal/render"
)

type OutputJSON struct {
	File     string    `json:"file"`
	Line     int       `json:"line"`
	Hash     string    `json:"hash"`
	Author   string    `json:"author"`
	Email    string    `json:"email"`
	Date     time.Time `json:"date"`
	Message  string    `json:"message"`
	Body     string    `json:"body,omitempty"`
}

func outputJSON(file string, line int, blame *git.BlameResult, commit *git.CommitInfo) error {
	output := OutputJSON{
		File:    file,
		Line:    line,
		Hash:    blame.Hash,
		Author:  blame.Author,
		Email:   blame.AuthorMail,
		Date:    commit.Date,
		Message: commit.Message,
		Body:    commit.Body,
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
	fmt.Printf("**Hash:** `%s`\n\n", blame.Hash[:12])
	fmt.Printf("**Author:** @%s\n\n", blame.Author)
	fmt.Printf("**Date:** %s\n\n", commit.Date.Format("2006-01-02"))
	fmt.Printf("**Message:** %s\n\n", commit.Message)

	if commit.Body != "" {
		fmt.Printf("## Context\n\n%s\n", commit.Body)
	}

	return nil
}

func renderResult(file string, line int, blame *git.BlameResult, commit *git.CommitInfo) {
	renderer := render.NewRenderer(plainOutput)
	renderer.RenderLine(file, line, blame, commit)
}
