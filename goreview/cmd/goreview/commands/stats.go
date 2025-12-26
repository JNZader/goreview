package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/history"
)

// Table border constants for dashboard output.
const (
	tableTop    = "┌─────────────────────────────────────────────────────┐"
	tableMid    = "├─────────────────────────────────────────────────────┤"
	tableBottom = "└─────────────────────────────────────────────────────┘"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show review statistics dashboard",
	Long: `Display aggregate statistics from all code reviews.

Shows a dashboard with:
- Total issues count and resolution rate
- Breakdown by severity level
- Breakdown by issue type
- Top files with most issues
- Top authors (if available)

Examples:
  # Show stats dashboard
  goreview stats

  # Output as JSON
  goreview stats --format=json`,
	RunE: runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)

	statsCmd.Flags().StringP("format", "f", "dashboard", "Output format (dashboard, json)")
}

func runStats(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	dbPath := getHistoryDBPath(cfg)

	store, err := history.NewStore(history.StoreConfig{Path: dbPath})
	if err != nil {
		return fmt.Errorf("opening history database: %w", err)
	}
	defer store.Close()

	ctx := context.Background()
	stats, err := store.GetStats(ctx)
	if err != nil {
		return fmt.Errorf("getting stats: %w", err)
	}

	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		return outputStatsJSON(stats)
	}

	outputStatsDashboard(stats)
	return nil
}

func outputStatsJSON(stats *history.Stats) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling stats: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputStatsDashboard(stats *history.Stats) {
	if stats.TotalIssues == 0 {
		fmt.Println("No review history found.")
		fmt.Println("\nRun some reviews first to collect statistics.")
		return
	}

	printDashboardHeader()
	resolutionRate := printDashboardSummary(stats)
	printResolutionProgress(resolutionRate)
	printSeveritySection(stats)
	printTypeSection(stats)
	printTopFilesSection(stats)
	printDashboardFooter()
}

func printDashboardHeader() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║              📊 GOREVIEW STATS DASHBOARD             ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()
}

func printDashboardSummary(stats *history.Stats) float64 {
	resolutionRate := float64(0)
	if stats.TotalIssues > 0 {
		resolutionRate = float64(stats.ResolvedIssues) / float64(stats.TotalIssues) * 100
	}
	pending := stats.TotalIssues - stats.ResolvedIssues

	fmt.Println(tableTop)
	fmt.Println("│                     📈 SUMMARY                      │")
	fmt.Println(tableMid)
	fmt.Printf("│  Total Issues:     %-6d                            │\n", stats.TotalIssues)
	fmt.Printf("│  Resolved:         %-6d (%.1f%%)                     │\n", stats.ResolvedIssues, resolutionRate)
	fmt.Printf("│  Pending:          %-6d                            │\n", pending)
	fmt.Println(tableBottom)
	fmt.Println()

	return resolutionRate
}

func printResolutionProgress(resolutionRate float64) {
	barWidth := 40
	filled := int(resolutionRate / 100 * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf("  Resolution Progress: [%s] %.1f%%\n\n", bar, resolutionRate)
}

func printSeveritySection(stats *history.Stats) {
	if len(stats.BySeverity) == 0 {
		return
	}

	fmt.Println(tableTop)
	fmt.Println("│                  🎯 BY SEVERITY                     │")
	fmt.Println(tableMid)

	severityOrder := []string{"critical", "error", "warning", "info"}
	for _, sev := range severityOrder {
		if count, ok := stats.BySeverity[sev]; ok && count > 0 {
			emoji := getSeverityEmoji(sev)
			percent := float64(count) / float64(stats.TotalIssues) * 100
			innerBar := makeProgressBar(int(count), int(stats.TotalIssues), 20)
			fmt.Printf("│  %s %-10s %s %-4d (%.0f%%)    │\n", emoji, sev, innerBar, count, percent)
		}
	}
	fmt.Println(tableBottom)
	fmt.Println()
}

// typeCount holds type name and count for sorting.
type typeCount struct {
	typ   string
	count int64
}

func printTypeSection(stats *history.Stats) {
	if len(stats.ByType) == 0 {
		return
	}

	fmt.Println(tableTop)
	fmt.Println("│                    🏷️  BY TYPE                      │")
	fmt.Println(tableMid)

	types := getSortedTypes(stats.ByType)
	for _, tc := range types {
		percent := float64(tc.count) / float64(stats.TotalIssues) * 100
		innerBar := makeProgressBar(int(tc.count), int(stats.TotalIssues), 20)
		fmt.Printf("│  %-12s %s %-4d (%.0f%%)        │\n", tc.typ, innerBar, tc.count, percent)
	}
	fmt.Println(tableBottom)
	fmt.Println()
}

func getSortedTypes(byType map[string]int64) []typeCount {
	types := make([]typeCount, 0, len(byType))
	for t, c := range byType {
		types = append(types, typeCount{t, c})
	}
	sort.Slice(types, func(i, j int) bool {
		return types[i].count > types[j].count
	})
	return types
}

// fileCount holds file path and issue count for sorting.
type fileCount struct {
	file  string
	count int64
}

func printTopFilesSection(stats *history.Stats) {
	if len(stats.ByFile) == 0 {
		return
	}

	fmt.Println(tableTop)
	fmt.Println("│                   📁 TOP FILES                      │")
	fmt.Println(tableMid)

	files := getSortedFiles(stats.ByFile)
	maxShow := min(10, len(files))

	for i := 0; i < maxShow; i++ {
		fc := files[i]
		displayPath := truncateFilePath(fc.file, 35)
		fmt.Printf("│  %-35s %4d issues    │\n", displayPath, fc.count)
	}
	fmt.Println(tableBottom)
	fmt.Println()
}

func getSortedFiles(byFile map[string]int64) []fileCount {
	files := make([]fileCount, 0, len(byFile))
	for f, c := range byFile {
		files = append(files, fileCount{f, c})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].count > files[j].count
	})
	return files
}

func truncateFilePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-(maxLen-3):]
}

func printDashboardFooter() {
	fmt.Println("  Use 'goreview search' to explore specific issues")
	fmt.Println("  Use 'goreview history <file>' for file-specific history")
	fmt.Println()
}

func makeProgressBar(current, total, width int) string {
	if total == 0 {
		return strings.Repeat("░", width)
	}

	filled := int(float64(current) / float64(total) * float64(width))
	if filled > width {
		filled = width
	}

	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}
