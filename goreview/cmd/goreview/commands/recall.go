package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/history"
)

var recallCmd = &cobra.Command{
	Use:   "recall [query]",
	Short: "Search and recall past commit analyses",
	Long: `Search through historical commit analyses to find past issues and patterns.

The recall command allows you to search through all previously analyzed commits
to find recurring issues, understand code evolution, and learn from past reviews.

Analyses are stored in .git/goreview/commits/ and include:
- Full analysis results with issues
- Context used during review (model, personality, modes)
- Markdown summaries for human reading

Examples:
  # Search for authentication-related issues
  goreview recall "authentication"

  # View analysis for a specific commit
  goreview recall --commit a1b2c3d

  # Show history for a specific file
  goreview recall --file src/auth/login.go

  # Search by author
  goreview recall --author john

  # Filter by severity
  goreview recall "memory" --severity critical

  # List all analyzed commits
  goreview recall --list`,
	RunE: runRecall,
}

var (
	recallCommit   string
	recallFile     string
	recallAuthor   string
	recallSeverity string
	recallLimit    int
	recallList     bool
	recallSince    string
	recallUntil    string
)

func init() {
	rootCmd.AddCommand(recallCmd)

	recallCmd.Flags().StringVarP(&recallCommit, "commit", "C", "", "View analysis for a specific commit")
	recallCmd.Flags().StringVarP(&recallFile, "file", "f", "", "Filter by file path")
	recallCmd.Flags().StringVarP(&recallAuthor, "author", "a", "", "Filter by author")
	recallCmd.Flags().StringVarP(&recallSeverity, "severity", "s", "", "Filter by severity (critical, error, warning, info)")
	recallCmd.Flags().IntVarP(&recallLimit, "limit", "l", 20, "Maximum number of results")
	recallCmd.Flags().BoolVar(&recallList, "list", false, "List all analyzed commits")
	recallCmd.Flags().StringVar(&recallSince, "since", "", "Show analyses since date (YYYY-MM-DD)")
	recallCmd.Flags().StringVar(&recallUntil, "until", "", "Show analyses until date (YYYY-MM-DD)")
}

func runRecall(cmd *cobra.Command, args []string) error {
	// Get repository root
	repoRoot, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("finding repository root: %w", err)
	}

	store, err := history.NewCommitStore(repoRoot)
	if err != nil {
		return fmt.Errorf("opening commit store: %w", err)
	}

	// List all commits
	if recallList {
		return listAnalyzedCommits(store)
	}

	// View specific commit
	if recallCommit != "" {
		return viewCommitAnalysis(store, recallCommit)
	}

	// File history
	if recallFile != "" && len(args) == 0 {
		return viewFileHistory(store, recallFile)
	}

	// Search query
	query := ""
	if len(args) > 0 {
		query = strings.Join(args, " ")
	}

	return searchAnalyses(store, query)
}

func listAnalyzedCommits(store *history.CommitStore) error {
	summaries, err := store.List()
	if err != nil {
		return err
	}

	if len(summaries) == 0 {
		fmt.Println("No commit analyses found.")
		fmt.Println("Run 'goreview review --commit <hash>' to analyze commits.")
		return nil
	}

	fmt.Printf("Analyzed Commits (%d total)\n", len(summaries))
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	for _, s := range summaries {
		// Format severity badges
		var badges []string
		if count, ok := s.Severities["critical"]; ok && count > 0 {
			badges = append(badges, fmt.Sprintf("%d critical", count))
		}
		if count, ok := s.Severities["error"]; ok && count > 0 {
			badges = append(badges, fmt.Sprintf("%d errors", count))
		}
		if count, ok := s.Severities["warning"]; ok && count > 0 {
			badges = append(badges, fmt.Sprintf("%d warnings", count))
		}

		issueStr := ""
		if s.IssueCount > 0 {
			issueStr = fmt.Sprintf(" (%s)", strings.Join(badges, ", "))
		}

		date := s.AnalyzedAt.Format(dateTimeFormat)
		msg := truncate(s.Message, 45)

		fmt.Printf("%s  %s  %s%s\n", s.Hash[:7], date, msg, issueStr)
	}

	return nil
}

func viewCommitAnalysis(store *history.CommitStore, commitHash string) error {
	analysis, err := store.Load(commitHash)
	if err != nil {
		return fmt.Errorf("commit analysis not found: %w", err)
	}

	printAnalysisHeader(analysis)
	printAnalysisSummary(analysis)
	printAnalysisSeverities(analysis)
	printAnalysisRecommendation(analysis)
	printAnalysisFiles(analysis)
	printAnalysisContext(analysis)

	return nil
}

func printAnalysisHeader(analysis *history.CommitAnalysis) {
	fmt.Printf("Commit Analysis: %s\n", analysis.CommitHash[:7])
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Printf("Commit:     %s\n", analysis.CommitMsg)
	fmt.Printf("Author:     %s <%s>\n", analysis.Author, analysis.AuthorEmail)
	fmt.Printf("Branch:     %s\n", analysis.Branch)
	fmt.Printf("Analyzed:   %s\n", analysis.AnalyzedAt.Format(time.RFC3339))
	fmt.Println()
}

func printAnalysisSummary(analysis *history.CommitAnalysis) {
	fmt.Println("Summary")
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("  Files:     %d\n", analysis.Summary.TotalFiles)
	fmt.Printf("  Issues:    %d\n", analysis.Summary.TotalIssues)
	fmt.Printf("  Score:     %.1f/100\n", analysis.Summary.OverallScore)
	fmt.Println()
}

func printAnalysisSeverities(analysis *history.CommitAnalysis) {
	if len(analysis.Summary.BySeverity) == 0 {
		return
	}
	fmt.Println("By Severity")
	fmt.Println(strings.Repeat("-", 30))
	for sev, count := range analysis.Summary.BySeverity {
		emoji := getSeverityEmoji(sev)
		fmt.Printf("  %s %-10s %d\n", emoji, sev, count)
	}
	fmt.Println()
}

func printAnalysisRecommendation(analysis *history.CommitAnalysis) {
	if analysis.Summary.Recommendation == "" {
		return
	}
	fmt.Println("Recommendation")
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("  %s\n\n", analysis.Summary.Recommendation)
}

func printAnalysisFiles(analysis *history.CommitAnalysis) {
	for _, file := range analysis.Files {
		if len(file.Issues) == 0 {
			continue
		}
		printFileWithIssues(file)
	}
}

func printFileWithIssues(file history.AnalyzedFile) {
	fmt.Printf("File: %s\n", file.Path)
	fmt.Printf("  Language: %s, Changes: +%d/-%d\n", file.Language, file.LinesAdded, file.LinesRemoved)
	fmt.Println()

	for _, issue := range file.Issues {
		printIssue(issue)
	}
}

func printIssue(issue history.Issue) {
	emoji := getSeverityEmoji(issue.Severity)
	location := ""
	if issue.Line > 0 {
		location = fmt.Sprintf(" (line %d)", issue.Line)
	}
	fmt.Printf("  %s [%s]%s\n", emoji, issue.Severity, location)
	fmt.Printf("     %s\n", issue.Message)
	if issue.Suggestion != "" {
		fmt.Printf("     Suggestion: %s\n", issue.Suggestion)
	}
	if issue.RootCause != nil {
		fmt.Printf("     Root Cause: %s\n", issue.RootCause.Description)
	}
	fmt.Println()
}

func printAnalysisContext(analysis *history.CommitAnalysis) {
	fmt.Println("Analysis Context")
	fmt.Println(strings.Repeat("-", 30))
	fmt.Printf("  Provider:    %s\n", analysis.Context.Provider)
	fmt.Printf("  Model:       %s\n", analysis.Context.Model)
	if analysis.Context.Personality != "" {
		fmt.Printf("  Personality: %s\n", analysis.Context.Personality)
	}
	if len(analysis.Context.Modes) > 0 {
		fmt.Printf("  Modes:       %s\n", strings.Join(analysis.Context.Modes, ", "))
	}
	if len(analysis.Context.RAGSources) > 0 {
		fmt.Printf("  RAG Sources: %s\n", strings.Join(analysis.Context.RAGSources, ", "))
	}
}

func viewFileHistory(store *history.CommitStore, filePath string) error {
	history, err := store.GetFileHistory(filePath)
	if err != nil {
		return err
	}

	if history.AnalyzedCommits == 0 {
		fmt.Printf("No analysis history found for: %s\n", filePath)
		return nil
	}

	fmt.Printf("File History: %s\n", filePath)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	// Stats
	fmt.Printf("Analyzed Commits:  %d\n", history.AnalyzedCommits)
	fmt.Printf("Total Issues:      %d\n", history.IssueStats.TotalIssues)
	fmt.Printf("Trend:             %s\n", formatTrend(history.IssueStats.TrendDirection))
	fmt.Println()

	// Commits
	fmt.Println("Commits")
	fmt.Println(strings.Repeat("-", 50))
	for _, c := range history.Commits {
		date := c.AnalyzedAt.Format(dateFormat)
		msg := truncate(c.Message, 35)
		fmt.Printf("%s  %s  %s  (%d issues)\n", c.Hash[:7], date, msg, c.IssueCount)
	}

	return nil
}

func searchAnalyses(store *history.CommitStore, query string) error {
	opts := history.RecallOptions{
		Query:    query,
		FilePath: recallFile,
		Author:   recallAuthor,
		Severity: recallSeverity,
		Limit:    recallLimit,
	}

	if recallSince != "" {
		t, err := time.Parse(dateFormat, recallSince)
		if err == nil {
			opts.Since = t
		}
	}
	if recallUntil != "" {
		t, err := time.Parse(dateFormat, recallUntil)
		if err == nil {
			opts.Until = t
		}
	}

	results, err := store.Recall(opts)
	if err != nil {
		return err
	}

	if len(results) == 0 {
		if query != "" {
			fmt.Printf("No results found for: %s\n", query)
		} else {
			fmt.Println("No results found with the given filters.")
		}
		return nil
	}

	if query != "" {
		fmt.Printf("Search Results for: %s\n", query)
	} else {
		fmt.Println("Search Results")
	}
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	for _, r := range results {
		date := r.AnalyzedAt.Format(dateFormat)
		matchIcon := getMatchIcon(r.MatchType)

		fmt.Printf("%s %s  %s  @%s\n", matchIcon, r.CommitHash[:7], date, r.Author)

		if r.FilePath != "" {
			fmt.Printf("   File: %s\n", r.FilePath)
		}

		fmt.Printf("   %s\n", r.Snippet)
		fmt.Println()
	}

	return nil
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("not a git repository")
		}
		dir = parent
	}
}

func formatTrend(trend string) string {
	switch trend {
	case "improving":
		return "Improving (fewer issues over time)"
	case "worsening":
		return "Worsening (more issues over time)"
	default:
		return "Stable"
	}
}

func getMatchIcon(matchType string) string {
	switch matchType {
	case "commit":
		return "C"
	case "file":
		return "F"
	case "issue":
		return "I"
	default:
		return "-"
	}
}
