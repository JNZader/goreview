package commands

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/mcp"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp-serve",
	Short: "Start GoReview as an MCP (Model Context Protocol) server",
	Long: `Start GoReview as an MCP server for integration with Claude Code and other MCP clients.

The server communicates over stdin/stdout using JSON-RPC 2.0.

Available tools:
  - goreview_review    Analyze code changes and identify issues
  - goreview_commit    Generate AI-powered commit messages
  - goreview_fix       Automatically fix issues found in review
  - goreview_search    Search through past code reviews
  - goreview_stats     Get code review statistics
  - goreview_changelog Generate changelog from commits
  - goreview_doc       Generate documentation for changes

Usage with Claude Code:

  # Add as local MCP server
  claude mcp add --transport stdio goreview -- goreview mcp-serve

  # Or configure in .mcp.json
  {
    "mcpServers": {
      "goreview": {
        "type": "stdio",
        "command": "goreview",
        "args": ["mcp-serve"]
      }
    }
  }

Examples:
  # Start server (used by MCP clients, not directly)
  goreview mcp-serve

  # Test with manual JSON-RPC input
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | goreview mcp-serve`,
	RunE: runMCPServe,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}

func runMCPServe(cmd *cobra.Command, args []string) error {
	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Create and configure MCP server
	server := mcp.NewServer()
	mcp.RegisterGoReviewTools(server)

	// Log to stderr (stdout is for MCP protocol)
	fmt.Fprintln(os.Stderr, "GoReview MCP server starting...")

	// Start serving
	if err := server.Serve(ctx); err != nil {
		if ctx.Err() != nil {
			// Context cancelled, normal shutdown
			fmt.Fprintln(os.Stderr, "GoReview MCP server stopped")
			return nil
		}
		return fmt.Errorf("MCP server error: %w", err)
	}

	fmt.Fprintln(os.Stderr, "GoReview MCP server stopped")
	return nil
}
