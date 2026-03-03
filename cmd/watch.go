package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch <file>",
	Short: "Monitor a file and show context interactively",
	Long: `Watch mode for integration with editors.
Reads line numbers from stdin and displays context automatically.`,
	Args: cobra.ExactArgs(1),
	RunE: runWatch,
}

func init() {
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	fmt.Printf("Watch mode for %s - Coming soon\n", args[0])
	return nil
}
