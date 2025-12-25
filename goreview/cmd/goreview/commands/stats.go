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

	return outputStatsDashboard(stats)
}

func outputStatsJSON(stats *history.Stats) error {
	data, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling stats: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputStatsDashboard(stats *history.Stats) error {
	if stats.TotalIssues == 0 {
		fmt.Println("No review history found.")
		fmt.Println("\nRun some reviews first to collect statistics.")
		return nil
	}

	// Header
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║              📊 GOREVIEW STATS DASHBOARD             ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	// Summary Section
	resolutionRate := float64(0)
	if stats.TotalIssues > 0 {
		resolutionRate = float64(stats.ResolvedIssues) / float64(stats.TotalIssues) * 100
	}
	pending := stats.TotalIssues - stats.ResolvedIssues

	fmt.Println("┌─────────────────────────────────────────────────────┐")
	fmt.Println("│                     📈 SUMMARY                      │")
	fmt.Println("├─────────────────────────────────────────────────────┤")
	fmt.Printf("│  Total Issues:     %-6d                            │\n", stats.TotalIssues)
	fmt.Printf("│  Resolved:         %-6d (%.1f%%)                     │\n", stats.ResolvedIssues, resolutionRate)
	fmt.Printf("│  Pending:          %-6d                            │\n", pending)
	fmt.Println("└─────────────────────────────────────────────────────┘")
	fmt.Println()

	// Resolution Progress Bar
	barWidth := 40
	filled := int(resolutionRate / 100 * float64(barWidth))
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
	fmt.Printf("  Resolution Progress: [%s] %.1f%%\n\n", bar, resolutionRate)

	// By Severity
	if len(stats.BySeverity) > 0 {
		fmt.Println("┌─────────────────────────────────────────────────────┐")
		fmt.Println("│                  🎯 BY SEVERITY                     │")
		fmt.Println("├─────────────────────────────────────────────────────┤")

		severityOrder := []string{"critical", "error", "warning", "info"}
		for _, sev := range severityOrder {
			if count, ok := stats.BySeverity[sev]; ok && count > 0 {
				emoji := getSeverityEmoji(sev)
				percent := float64(count) / float64(stats.TotalIssues) * 100
				innerBar := makeProgressBar(int(count), int(stats.TotalIssues), 20)
				fmt.Printf("│  %s %-10s %s %-4d (%.0f%%)    │\n", emoji, sev, innerBar, count, percent)
			}
		}
		fmt.Println("└─────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	// By Type
	if len(stats.ByType) > 0 {
		fmt.Println("┌─────────────────────────────────────────────────────┐")
		fmt.Println("│                    🏷️  BY TYPE                      │")
		fmt.Println("├─────────────────────────────────────────────────────┤")

		// Sort by count descending
		type typeCount struct {
			typ   string
			count int64
		}
		var types []typeCount
		for t, c := range stats.ByType {
			types = append(types, typeCount{t, c})
		}
		sort.Slice(types, func(i, j int) bool {
			return types[i].count > types[j].count
		})

		for _, tc := range types {
			percent := float64(tc.count) / float64(stats.TotalIssues) * 100
			innerBar := makeProgressBar(int(tc.count), int(stats.TotalIssues), 20)
			fmt.Printf("│  %-12s %s %-4d (%.0f%%)        │\n", tc.typ, innerBar, tc.count, percent)
		}
		fmt.Println("└─────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	// Top Files
	if len(stats.ByFile) > 0 {
		fmt.Println("┌─────────────────────────────────────────────────────┐")
		fmt.Println("│                   📁 TOP FILES                      │")
		fmt.Println("├─────────────────────────────────────────────────────┤")

		// Sort by count descending
		type fileCount struct {
			file  string
			count int64
		}
		var files []fileCount
		for f, c := range stats.ByFile {
			files = append(files, fileCount{f, c})
		}
		sort.Slice(files, func(i, j int) bool {
			return files[i].count > files[j].count
		})

		maxShow := 10
		if len(files) < maxShow {
			maxShow = len(files)
		}

		for i := 0; i < maxShow; i++ {
			fc := files[i]
			// Truncate long file paths
			displayPath := fc.file
			if len(displayPath) > 35 {
				displayPath = "..." + displayPath[len(displayPath)-32:]
			}
			fmt.Printf("│  %-35s %4d issues    │\n", displayPath, fc.count)
		}
		fmt.Println("└─────────────────────────────────────────────────────┘")
		fmt.Println()
	}

	// Footer
	fmt.Println("  Use 'goreview search' to explore specific issues")
	fmt.Println("  Use 'goreview history <file>' for file-specific history")
	fmt.Println()

	return nil
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
