package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	contextLines int
	shortOutput  bool
	plainOutput  bool
	jsonOutput   bool
	mdOutput     bool
	llmOutput    bool
	includeDiff  bool
	version      = "dev"
	commit       = "none"
	buildDate    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "git-why",
	Short: "Understand why code exists, not just who wrote it",
	Long: `git-why answers the question every developer asks when reading legacy code:
"Why does this line exist?"

It combines git log, git blame, and semantic search to present historical
context for any line of code in a readable and actionable format.`,
	Version: version,
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return runWhy(cmd, args)
	},
}

func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	buildDate = d
	rootCmd.Version = fmt.Sprintf("%s (commit %s, built %s)", version, commit, buildDate)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().IntVarP(&contextLines, "context", "c", 3, "number of context lines before and after")
	rootCmd.PersistentFlags().BoolVarP(&shortOutput, "short", "s", false, "compact one-line output")
	rootCmd.PersistentFlags().BoolVarP(&plainOutput, "plain", "p", false, "output without colors")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&mdOutput, "md", false, "output in Markdown format")
	rootCmd.PersistentFlags().BoolVar(&llmOutput, "llm", false, "output optimized for LLM agents")
	rootCmd.PersistentFlags().BoolVar(&includeDiff, "include-diff", false, "include commit diff in output")
}
