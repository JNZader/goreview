package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/history"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search review history",
	Long: `Search through past code reviews using full-text search.

Examples:
  # Full-text search for "memory leak"
  goreview search "memory leak"

  # Search issues in a specific file
  goreview search --file=auth.go

  # Search by author
  goreview search --author=john

  # Search critical security issues
  goreview search --severity=critical --type=security

  # Combine filters
  goreview search "null pointer" --file="src/api/*" --severity=error`,
	RunE: runSearch,
}

func init() {
	rootCmd.AddCommand(searchCmd)

	searchCmd.Flags().String("file", "", "Filter by file path (supports glob patterns)")
	searchCmd.Flags().String("author", "", "Filter by commit author")
	searchCmd.Flags().String("severity", "", "Filter by severity (info, warning, error, critical)")
	searchCmd.Flags().String("type", "", "Filter by issue type (bug, security, performance, style)")
	searchCmd.Flags().String("branch", "", "Filter by git branch")
	searchCmd.Flags().String("since", "", "Filter issues after date (YYYY-MM-DD)")
	searchCmd.Flags().String("until", "", "Filter issues before date (YYYY-MM-DD)")
	searchCmd.Flags().Bool("resolved", false, "Show only resolved issues")
	searchCmd.Flags().Bool("unresolved", false, "Show only unresolved issues")
	searchCmd.Flags().Int("limit", 50, "Maximum number of results")
	searchCmd.Flags().StringP("format", "f", "table", "Output format (table, json)")
}

//nolint:gocyclo,gocognit // CLI command with multiple flag handling paths
func runSearch(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Get history database path
	dbPath := getHistoryDBPath(cfg)

	store, err := history.NewStore(history.StoreConfig{Path: dbPath})
	if err != nil {
		return fmt.Errorf("opening history database: %w", err)
	}
	defer store.Close()

	// Build query
	query := history.SearchQuery{}

	if len(args) > 0 {
		query.Text = strings.Join(args, " ")
	}

	if file, _ := cmd.Flags().GetString("file"); file != "" {
		query.File = file
	}
	if author, _ := cmd.Flags().GetString("author"); author != "" {
		query.Author = author
	}
	if severity, _ := cmd.Flags().GetString("severity"); severity != "" {
		query.Severity = severity
	}
	if typ, _ := cmd.Flags().GetString("type"); typ != "" {
		query.Type = typ
	}
	if branch, _ := cmd.Flags().GetString("branch"); branch != "" {
		query.Branch = branch
	}

	if since, _ := cmd.Flags().GetString("since"); since != "" {
		sinceTime, parseErr := time.Parse(dateFormat, since)
		if parseErr != nil {
			return fmt.Errorf("invalid since date: %w", parseErr)
		}
		query.Since = sinceTime
	}
	if until, _ := cmd.Flags().GetString("until"); until != "" {
		untilTime, parseErr := time.Parse(dateFormat, until)
		if parseErr != nil {
			return fmt.Errorf("invalid until date: %w", parseErr)
		}
		query.Until = untilTime
	}

	resolved, _ := cmd.Flags().GetBool("resolved")
	unresolved, _ := cmd.Flags().GetBool("unresolved")
	if resolved {
		r := true
		query.Resolved = &r
	} else if unresolved {
		r := false
		query.Resolved = &r
	}

	query.Limit, _ = cmd.Flags().GetInt("limit")

	// Execute search
	ctx := context.Background()
	result, err := store.Search(ctx, query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	// Output results
	format, _ := cmd.Flags().GetString("format")
	return outputSearchResults(result, format)
}

//nolint:unparam // error return kept for consistency with other output functions
func outputSearchResults(result *history.SearchResult, format string) error {
	if len(result.Records) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("Found %d results (showing %d)\n\n", result.TotalCount, len(result.Records))

	if format == "json" {
		// JSON output
		for _, r := range result.Records {
			fmt.Printf(`{"file":"%s","line":%d,"severity":"%s","type":"%s","message":"%s"}%s`,
				r.FilePath, r.Line, r.Severity, r.IssueType,
				strings.ReplaceAll(r.Message, `"`, `\"`), "\n")
		}
		return nil
	}

	// Table output
	for _, r := range result.Records {
		emoji := getSeverityEmoji(r.Severity)
		location := r.FilePath
		if r.Line > 0 {
			location = fmt.Sprintf("%s:%d", r.FilePath, r.Line)
		}

		status := ""
		if r.Resolved {
			status = " [RESOLVED]"
		}

		fmt.Printf("%s [%s] %s%s\n", emoji, strings.ToUpper(r.Severity), location, status)
		fmt.Printf("   %s\n", truncate(r.Message, 80))
		if r.Suggestion != "" {
			fmt.Printf("   💡 %s\n", truncate(r.Suggestion, 80))
		}
		fmt.Printf("   📅 %s | 👤 %s | 🔀 %s\n\n",
			r.CreatedAt.Format("2006-01-02"), r.Author, r.Branch)
	}

	return nil
}

func getSeverityEmoji(severity string) string {
	switch severity {
	case "critical":
		return "🚨"
	case "error":
		return "❌"
	case "warning":
		return "⚠️"
	default:
		return "ℹ️"
	}
}

// truncate is defined in commit_interactive.go

func getHistoryDBPath(_ *config.Config) string {
	// Default to .goreview/history.db in home directory
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "."
	}

	dir := filepath.Join(home, ".goreview")
	_ = os.MkdirAll(dir, 0750) //nolint:errcheck // Best effort directory creation

	return filepath.Join(dir, "history.db")
}
