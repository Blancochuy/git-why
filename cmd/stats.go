package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats <file>",
	Short: "Show authorship statistics for a file",
	Long: `Display macro view of who has touched a file most,
what parts change most frequently, and change frequency over time.`,
	Args: cobra.ExactArgs(1),
	RunE: runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStats(cmd *cobra.Command, args []string) error {
	fmt.Printf("Stats for %s - Coming soon\n", args[0])
	return nil
}
