package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chuy/git-why/internal/git"
	"github.com/chuy/git-why/internal/render"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch <file>",
	Short: "Monitor a file and show context interactively",
	Long: `Watch mode for integration with editors.
Reads line numbers from stdin and displays context automatically.

This is useful for editor integrations (VSCode, Neovim, etc.)
where you can pipe the current line number to get instant context.

Examples:
  # Interactive mode - type line numbers
  git-why watch src/auth.ts

  # Pipe from editor
  echo 142 | git-why watch src/auth.ts

  # Integration with Neovim
  :.!git-why watch src/auth.ts`,
	Args: cobra.ExactArgs(1),
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	file := args[0]

	repo := git.NewRepo()
	if !repo.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	stat, _ := os.Stdin.Stat()
	isPiped := (stat.Mode() & os.ModeCharDevice) == 0

	if isPiped {
		return runWatchPiped(repo, file)
	}
	return runWatchInteractive(repo, file)
}

func runWatchPiped(repo *git.Repo, file string) error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lineStr := strings.TrimSpace(scanner.Text())
		lineStr = strings.Trim(lineStr, "\ufeff")
		lineStr = strings.TrimSpace(lineStr)
		if lineStr == "" {
			continue
		}

		lineNum, err := strconv.Atoi(lineStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid line number: %s\n", lineStr)
			continue
		}

		if err := showLineWatch(repo, file, lineNum); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}
	return nil
}

func runWatchInteractive(repo *git.Repo, file string) error {
	fmt.Printf("Watching %s\n", file)
	fmt.Println("Enter line number (or 'q' to quit):")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "q" || input == "quit" || input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		lineNum, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Invalid line number. Enter a number or 'q' to quit.")
			continue
		}

		if err := showLineWatch(repo, file, lineNum); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
		fmt.Println()
	}

	return nil
}

func showLineWatch(repo *git.Repo, file string, lineNum int) error {
	blame, err := repo.Blame(file, lineNum)
	if err != nil {
		return err
	}

	commit, err := repo.ShowCommit(blame.Hash)
	if err != nil {
		return err
	}

	if plainOutput {
		fmt.Printf("Line %d: %s · %s · @%s · %s\n",
			lineNum,
			blame.Hash[:7],
			commit.Date.Format("2006-01-02"),
			blame.Author,
			strings.Split(commit.Message, "\n")[0],
		)
		return nil
	}

	renderer := render.NewRenderer(false)
	renderer.RenderLine(file, lineNum, blame, commit)
	return nil
}

type WatchOutput struct {
	File    string    `json:"file"`
	Line    int       `json:"line"`
	Hash    string    `json:"hash"`
	Author  string    `json:"author"`
	Date    time.Time `json:"date"`
	Message string    `json:"message"`
}

func formatWatchOutput(file string, lineNum int, blame *git.BlameResult, commit *git.CommitInfo) string {
	return fmt.Sprintf("Line %d: %s · %s · @%s · %s",
		lineNum,
		blame.Hash[:7],
		commit.Date.Format("2006-01-02"),
		blame.Author,
		strings.Split(commit.Message, "\n")[0],
	)
}

type gitBlameResult = git.BlameResult
type gitCommitInfo = git.CommitInfo
