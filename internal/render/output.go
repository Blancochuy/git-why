package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/chuy/git-why/internal/git"
)

type Colors struct {
	Hash    string
	Author  string
	Date    string
	Message string
	File    string
	Header  string
	Reset   string
}

var defaultColors = Colors{
	Hash:    "\033[36m",
	Author:  "\033[33m",
	Date:    "\033[32m",
	Message: "\033[0m",
	File:    "\033[35m",
	Header:  "\033[1m",
	Reset:   "\033[0m",
}

var plainColors = Colors{
	Hash:    "",
	Author:  "",
	Date:    "",
	Message: "",
	File:    "",
	Header:  "",
	Reset:   "",
}

type Renderer struct {
	colors Colors
	plain  bool
}

func NewRenderer(plain bool) *Renderer {
	if plain {
		return &Renderer{colors: plainColors, plain: true}
	}
	return &Renderer{colors: defaultColors, plain: false}
}

func (r *Renderer) RenderLine(file string, lineNum int, blame *git.BlameResult, commit *git.CommitInfo) {
	c := r.colors

	fmt.Printf("\n%s%sLine %d%s — %s%s%s — Added: %s%s by %s@%s%s\n",
		c.Header, c.File, lineNum, c.Reset,
		c.Hash, blame.Hash[:12], c.Reset,
		c.Date, commit.Date.Format("2006-01-02"),
		c.Author, blame.Author, c.Reset,
	)

	fmt.Printf("\n%sMessage:%s\n  %s\n",
		c.Header, c.Reset,
		strings.Split(commit.Message, "\n")[0],
	)

	if commit.Body != "" && strings.TrimSpace(commit.Body) != "" {
		fmt.Printf("\n%sContext:%s\n  %s\n",
			c.Header, c.Reset,
			strings.TrimSpace(commit.Body),
		)
	}

	fmt.Println()
}

func (r *Renderer) RenderRange(file string, start, end int, commits []git.CommitInfo) {
	c := r.colors

	fmt.Printf("\n%s%s%s:%d-%d%s\n\n",
		c.Header, c.File, file, start, end, c.Reset,
	)

	for _, commit := range commits {
		fmt.Printf("  %s%s%s %s %s@%s%s\n    %s%s\n\n",
			c.Hash, commit.Hash[:7], c.Reset,
			commit.Date.Format("2006-01-02"),
			c.Author, commit.Author, c.Reset,
			c.Message, strings.Split(commit.Message, "\n")[0],
		)
	}
}

func FormatShort(blame *git.BlameResult, commit *git.CommitInfo) string {
	return fmt.Sprintf("%s · %s · @%s · %s",
		blame.Hash[:7],
		commit.Date.Format("2006-01-02"),
		blame.Author,
		strings.Split(commit.Message, "\n")[0],
	)
}

func (r *Renderer) RenderStats(file string, stats map[string]int, totalLines int) {
	c := r.colors

	fmt.Printf("\n%sStatistics for %s%s%s\n", c.Header, c.File, file, c.Reset)
	fmt.Printf("Total lines: %d\n\n", totalLines)

	fmt.Println("Top contributors by lines modified:")
	for author, count := range stats {
		pct := float64(count) / float64(totalLines) * 100
		fmt.Printf("  %s@%s%s: %d lines (%.1f%%)\n",
			c.Author, author, c.Reset, count, pct,
		)
	}
}

func FormatDate(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff.Hours() < 24:
		return "today"
	case diff.Hours() < 48:
		return "yesterday"
	case diff.Hours() < 24*7:
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	case diff.Hours() < 24*30:
		return fmt.Sprintf("%d weeks ago", int(diff.Hours()/(24*7)))
	default:
		return t.Format("2006-01-02")
	}
}
