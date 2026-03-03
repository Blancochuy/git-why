package cmd

import (
	"os"

	"github.com/chuy/git-why/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the git-why MCP server on stdio",
	Long: `Start a Model Context Protocol (MCP) server that allows AI agents
to use git-why as a native tool. The server communicates over standard I/O.`,
	Run: func(cmd *cobra.Command, args []string) {
		server := mcp.NewServer(os.Stdin, os.Stdout)
		if err := server.Run(); err != nil {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
