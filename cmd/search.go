package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/blancochuy/git-why/internal/git"
	"github.com/spf13/cobra"
)

var (
	searchTerm string
	useRegex   bool
)

var searchCmd = &cobra.Command{
	Use:   "search <file> --term <text>",
	Short: "Search commit history for a file by text or regex",
	Long: `Filter and show commits where the message or diff contains the search term.

Examples:
  git-why search src/auth.ts --term "security"
  git-why search src/auth.ts --term "fix.*token" --regex`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

func init() {
	searchCmd.Flags().StringVarP(&searchTerm, "term", "t", "", "search term (required)")
	searchCmd.Flags().BoolVarP(&useRegex, "regex", "r", false, "treat term as regex")
	searchCmd.MarkFlagRequired("term")
	rootCmd.AddCommand(searchCmd)
}

func runSearch(cmd *cobra.Command, args []string) error {
	file := args[0]

	repo := git.NewRepo()
	if !repo.IsGitRepo() {
		return fmt.Errorf("not a git repository")
	}

	commits, err := repo.SearchHistory(file, searchTerm, useRegex)
	if err != nil {
		return err
	}

	if len(commits) == 0 {
		fmt.Printf("No commits found matching '%s' in %s\n", searchTerm, file)
		return nil
	}

	fmt.Printf("Commits in %s matching '%s':\n\n", file, searchTerm)

	for _, c := range commits {
		msg := strings.Split(c.Message, "\n")[0]
		if shortOutput {
			fmt.Printf("%s - %s - @%s - %s\n",
				c.Hash[:7],
				c.Date.Format("2006-01-02"),
				c.Author,
				highlightTerm(msg, searchTerm),
			)
		} else {
			fmt.Printf("  %s %s - %s\n    %s\n",
				c.Hash[:7],
				c.Date.Format("2006-01-02"),
				c.Author,
				highlightTerm(msg, searchTerm),
			)
			if c.Body != "" {
				fmt.Printf("    %s\n", highlightTerm(strings.TrimSpace(c.Body), searchTerm))
			}
			fmt.Println()
		}
	}

	return nil
}

func highlightTerm(text, term string) string {
	if plainOutput {
		return text
	}

	var pattern *regexp.Regexp
	if useRegex {
		pattern = regexp.MustCompile(term)
	} else {
		pattern = regexp.MustCompile(regexp.QuoteMeta(term))
	}

	highlight := "\033[33;1m"
	reset := "\033[0m"

	return pattern.ReplaceAllStringFunc(text, func(match string) string {
		return highlight + match + reset
	})
}

