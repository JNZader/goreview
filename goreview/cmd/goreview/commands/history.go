package commands

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/history"
)

var historyCmd = &cobra.Command{
	Use:   "history [path]",
	Short: "Show review history for a file or directory",
	Long: `Display the review history for a specific file or directory.

Shows aggregate statistics including:
- Total issues found over time
- Resolution rate
- Issue breakdown by severity and type
- First and last review dates

Examples:
  # History for a specific file
  goreview history src/auth/login.go

  # History for a directory
  goreview history src/api/

  # History for entire project
  goreview history .`,
	RunE: runHistory,
}

func init() {
	rootCmd.AddCommand(historyCmd)

	historyCmd.Flags().Bool("detailed", false, "Show detailed issue list")
	historyCmd.Flags().Int("limit", 20, "Number of issues to show in detailed mode")
}

func runHistory(cmd *cobra.Command, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	store, err := history.NewStore(history.StoreConfig{Path: getHistoryDBPath(cfg)})
	if err != nil {
		return fmt.Errorf("opening history database: %w", err)
	}
	defer store.Close()

	ctx := context.Background()
	hist, err := store.GetFileHistory(ctx, path)
	if err != nil {
		return fmt.Errorf("getting history: %w", err)
	}

	if hist.TotalIssues == 0 {
		fmt.Printf("No review history found for: %s\n", path)
		return nil
	}

	printHistoryHeader(path)
	printHistorySummary(hist)
	printHistoryTimeline(hist)
	printSeverityBreakdown(hist)
	printTypeBreakdown(hist)

	detailed, _ := cmd.Flags().GetBool("detailed")
	if detailed {
		limit, _ := cmd.Flags().GetInt("limit")
		return printDetailedIssues(ctx, store, path, limit)
	}

	return nil
}

func printHistoryHeader(path string) {
	fmt.Printf("📁 Review History: %s\n", path)
	fmt.Println(repeatChar('=', 50))
	fmt.Println()
}

func printHistorySummary(hist *history.FileHistory) {
	resolutionRate := float64(0)
	if hist.TotalIssues > 0 {
		resolutionRate = float64(hist.Resolved) / float64(hist.TotalIssues) * 100
	}

	fmt.Printf("📊 Summary\n")
	fmt.Printf("   Total Issues:    %d\n", hist.TotalIssues)
	fmt.Printf("   Resolved:        %d (%.1f%%)\n", hist.Resolved, resolutionRate)
	fmt.Printf("   Pending:         %d\n", hist.Pending)
	fmt.Printf("   Review Rounds:   %d\n", hist.ReviewRounds)
	fmt.Println()
}

func printHistoryTimeline(hist *history.FileHistory) {
	if hist.FirstReview.IsZero() {
		return
	}
	fmt.Printf("📅 Timeline\n")
	fmt.Printf("   First Review:    %s\n", hist.FirstReview.Format("2006-01-02 15:04"))
	fmt.Printf("   Last Review:     %s\n", hist.LastReview.Format("2006-01-02 15:04"))
	fmt.Println()
}

func printSeverityBreakdown(hist *history.FileHistory) {
	if len(hist.BySeverity) == 0 {
		return
	}
	fmt.Printf("🎯 By Severity\n")
	severityOrder := []string{"critical", "error", "warning", "info"}
	for _, sev := range severityOrder {
		if count, ok := hist.BySeverity[sev]; ok && count > 0 {
			emoji := getSeverityEmoji(sev)
			bar := progressBar(count, int(hist.TotalIssues), 20)
			fmt.Printf("   %s %-10s %s %d\n", emoji, sev, bar, count)
		}
	}
	fmt.Println()
}

func printTypeBreakdown(hist *history.FileHistory) {
	if len(hist.ByType) == 0 {
		return
	}
	fmt.Printf("🏷️  By Type\n")
	for typ, count := range hist.ByType {
		bar := progressBar(count, int(hist.TotalIssues), 20)
		fmt.Printf("   %-15s %s %d\n", typ, bar, count)
	}
	fmt.Println()
}

func printDetailedIssues(ctx context.Context, store *history.Store, path string, limit int) error {
	result, err := store.Search(ctx, history.SearchQuery{
		File:  path,
		Limit: limit,
	})
	if err != nil {
		return fmt.Errorf("searching issues: %w", err)
	}

	if len(result.Records) == 0 {
		return nil
	}

	fmt.Printf("📋 Recent Issues (showing %d of %d)\n", len(result.Records), result.TotalCount)
	fmt.Println()

	for _, r := range result.Records {
		printIssueRecord(r)
	}
	return nil
}

func printIssueRecord(r history.ReviewRecord) {
	emoji := getSeverityEmoji(r.Severity)
	status := ""
	if r.Resolved {
		status = " ✓"
	}

	location := r.FilePath
	if r.Line > 0 {
		location = fmt.Sprintf("%s:%d", r.FilePath, r.Line)
	}

	fmt.Printf("%s [%s]%s %s\n", emoji, r.Severity, status, location)
	fmt.Printf("   %s\n", truncate(r.Message, 70))
	fmt.Println()
}

func progressBar(current, total, width int) string {
	if total == 0 {
		return repeatChar('░', width)
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	return repeatChar('█', filled) + repeatChar('░', width-filled)
}

func repeatChar(c rune, n int) string {
	if n <= 0 {
		return ""
	}
	result := make([]rune, n)
	for i := range result {
		result[i] = c
	}
	return string(result)
}
