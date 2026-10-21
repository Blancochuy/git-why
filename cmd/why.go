package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/blancochuy/git-why/internal/git"
	"github.com/blancochuy/git-why/internal/render"
	"github.com/spf13/cobra"
)

var (
	funcName string
)

var whyCmd = &cobra.Command{
	Use:   "why <file>:<line> or <file>:<start>-<end> or <file> --fn <function>",
	Short: "Show historical context for a line or range of lines",
	Long: `Show who wrote a line, when, and most importantly - why.

Examples:
  git-why src/auth.ts:142                    # Single line
  git-why src/auth.ts:140-155                # Range of lines
  git-why src/auth.ts --fn authenticateUser  # Function history
  git-why src/auth.ts:142 --short            # Compact output`,
	Args: cobra.ExactArgs(1),
	RunE: runWhy,
}

func init() {
	// Registered on the root command as well so `git-why <file> --fn <name>`
	// works. The why subcommand keeps its own flag for `git-why why ... --fn`.
	rootCmd.Flags().StringVarP(&funcName, "fn", "f", "", "function name to analyze")
	whyCmd.Flags().StringVarP(&funcName, "fn", "f", "", "function name to analyze")
	rootCmd.AddCommand(whyCmd)
}

func runWhy(cmd *cobra.Command, args []string) error {
	target := args[0]

	repo := git.NewRepo()
	if !repo.IsGitRepo() {
		return fmt.Errorf("not a git repository (or any parent up to mount point)")
	}

	if funcName != "" {
		return showFunction(repo, target, funcName)
	}

	file, line, endLine, err := parseTarget(target)
	if err != nil {
		return err
	}

	if endLine > 0 {
		return showRange(repo, file, line, endLine)
	}
	return showLine(repo, file, line)
}

func showFunction(repo *git.Repo, file, funcName string) error {
	bounds, err := repo.FindFunction(file, funcName)
	if err != nil {
		return err
	}

	if !jsonOutput && !mdOutput && !llmOutput {
		fmt.Printf("Function %s() in %s (lines %d-%d):\n\n", funcName, file, bounds.StartLine, bounds.EndLine)
	}
	return showRange(repo, file, bounds.StartLine, bounds.EndLine)
}

func parseTarget(target string) (file string, line, endLine int, err error) {
	rangePattern := regexp.MustCompile(`^(.+):(\d+)-(\d+)$`)
	singlePattern := regexp.MustCompile(`^(.+):(\d+)$`)

	if matches := rangePattern.FindStringSubmatch(target); matches != nil {
		file = matches[1]
		line, _ = strconv.Atoi(matches[2])
		endLine, _ = strconv.Atoi(matches[3])
		if line < 1 || endLine < line {
			err = fmt.Errorf("invalid range %d-%d", line, endLine)
		}
		return
	}

	if matches := singlePattern.FindStringSubmatch(target); matches != nil {
		file = matches[1]
		line, _ = strconv.Atoi(matches[2])
		return
	}

	err = fmt.Errorf("invalid target format. Use <file>:<line> or <file>:<start>-<end>")
	return
}

func showLine(repo *git.Repo, file string, line int) error {
	blame, err := repo.Blame(file, line)
	if err != nil {
		return err
	}
	if git.IsUncommitted(blame.Hash) {
		return fmt.Errorf("%s:%d has uncommitted changes", file, line)
	}

	commit, err := repo.ShowCommit(blame.Hash)
	if err != nil {
		return err
	}

	if shortOutput {
		fmt.Println(render.FormatShort(blame, commit))
		return nil
	}

	if includeDiff || llmOutput {
		diff, err := repo.GetCommitDiff(blame.Hash)
		if err == nil {
			commit.Diff = diff
		}
	}

	if llmOutput {
		return outputLLM(file, line, blame, commit)
	}

	if jsonOutput {
		return outputJSON(file, line, blame, commit)
	}

	if mdOutput {
		return outputMarkdown(file, line, blame, commit)
	}

	renderResult(file, line, blame, commit)
	return nil
}

func showRange(repo *git.Repo, file string, start, end int) error {
	commits, err := repo.LogRange(file, start, end)
	if err != nil {
		return err
	}

	if jsonOutput {
		return outputJSONRange(file, start, end, commits)
	}
	if mdOutput {
		return outputMarkdownRange(file, start, end, commits)
	}
	if llmOutput {
		return outputLLMRange(file, start, end, commits)
	}

	fmt.Printf("Commits that modified %s:%d-%d:\n\n", file, start, end)
	for _, c := range commits {
		if shortOutput {
			fmt.Printf("%s - %s - @%s - %s\n",
				c.Hash[:7],
				c.Date.Format("2006-01-02"),
				c.Author,
				strings.Split(c.Message, "\n")[0],
			)
		} else {
			fmt.Printf("  %s %s - %s\n    %s\n\n",
				c.Hash[:7],
				c.Date.Format("2006-01-02"),
				c.Author,
				strings.Split(c.Message, "\n")[0],
			)
		}
	}
	return nil
}
