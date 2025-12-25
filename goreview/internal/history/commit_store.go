package history

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// CommitStore handles file-based storage of commit analyses.
// Analyses are stored in .git/goreview/commits/<hash>/
type CommitStore struct {
	repoRoot string
	baseDir  string
}

// NewCommitStore creates a new commit store for the given repository.
func NewCommitStore(repoRoot string) (*CommitStore, error) {
	// Find .git directory
	gitDir := filepath.Join(repoRoot, ".git")
	if info, err := os.Stat(gitDir); err != nil || !info.IsDir() {
		return nil, fmt.Errorf("not a git repository: %s", repoRoot)
	}

	baseDir := filepath.Join(gitDir, "goreview", "commits")
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("creating commits directory: %w", err)
	}

	return &CommitStore{
		repoRoot: repoRoot,
		baseDir:  baseDir,
	}, nil
}

// Store saves a commit analysis.
func (cs *CommitStore) Store(analysis *CommitAnalysis) error {
	// Use short hash for directory name (first 7 chars)
	shortHash := analysis.CommitHash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	commitDir := filepath.Join(cs.baseDir, shortHash)
	if err := os.MkdirAll(commitDir, 0755); err != nil {
		return fmt.Errorf("creating commit directory: %w", err)
	}

	// Store full analysis as JSON
	analysisPath := filepath.Join(commitDir, "analysis.json")
	data, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling analysis: %w", err)
	}
	if err := os.WriteFile(analysisPath, data, 0644); err != nil {
		return fmt.Errorf("writing analysis: %w", err)
	}

	// Store issues separately for quick access
	issuesPath := filepath.Join(commitDir, "issues.json")
	var allIssues []Issue
	for _, f := range analysis.Files {
		allIssues = append(allIssues, f.Issues...)
	}
	issuesData, _ := json.MarshalIndent(allIssues, "", "  ")
	_ = os.WriteFile(issuesPath, issuesData, 0644)

	// Store context for reference
	contextPath := filepath.Join(commitDir, "context.json")
	contextData, _ := json.MarshalIndent(analysis.Context, "", "  ")
	_ = os.WriteFile(contextPath, contextData, 0644)

	// Generate markdown summary
	mdPath := filepath.Join(commitDir, "analysis.md")
	_ = cs.generateMarkdownSummary(analysis, mdPath)

	return nil
}

// Load retrieves a commit analysis by hash.
func (cs *CommitStore) Load(commitHash string) (*CommitAnalysis, error) {
	// Handle both short and full hashes
	shortHash := commitHash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}

	analysisPath := filepath.Join(cs.baseDir, shortHash, "analysis.json")
	data, err := os.ReadFile(analysisPath)
	if err != nil {
		return nil, fmt.Errorf("reading analysis: %w", err)
	}

	var analysis CommitAnalysis
	if err := json.Unmarshal(data, &analysis); err != nil {
		return nil, fmt.Errorf("parsing analysis: %w", err)
	}

	return &analysis, nil
}

// Exists checks if an analysis exists for a commit.
func (cs *CommitStore) Exists(commitHash string) bool {
	shortHash := commitHash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}
	_, err := os.Stat(filepath.Join(cs.baseDir, shortHash, "analysis.json"))
	return err == nil
}

// List returns all stored commit analyses.
func (cs *CommitStore) List() ([]CommitSummary, error) {
	entries, err := os.ReadDir(cs.baseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading commits directory: %w", err)
	}

	var summaries []CommitSummary
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		analysis, err := cs.Load(entry.Name())
		if err != nil {
			continue
		}

		severities := make(map[string]int)
		for _, f := range analysis.Files {
			for _, issue := range f.Issues {
				severities[issue.Severity]++
			}
		}

		summaries = append(summaries, CommitSummary{
			Hash:       analysis.CommitHash,
			Message:    analysis.CommitMsg,
			Author:     analysis.Author,
			AnalyzedAt: analysis.AnalyzedAt,
			IssueCount: analysis.Summary.TotalIssues,
			Severities: severities,
		})
	}

	// Sort by analysis date, most recent first
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].AnalyzedAt.After(summaries[j].AnalyzedAt)
	})

	return summaries, nil
}

// Recall searches commit analyses for a query.
func (cs *CommitStore) Recall(opts RecallOptions) ([]RecallResult, error) {
	var results []RecallResult
	query := strings.ToLower(opts.Query)

	// If specific commit requested
	if opts.CommitHash != "" {
		analysis, err := cs.Load(opts.CommitHash)
		if err != nil {
			return nil, err
		}

		results = append(results, RecallResult{
			CommitHash: analysis.CommitHash,
			CommitMsg:  analysis.CommitMsg,
			Author:     analysis.Author,
			AnalyzedAt: analysis.AnalyzedAt,
			MatchType:  "commit",
			Snippet:    formatAnalysisSummary(analysis),
			Score:      1.0,
		})
		return results, nil
	}

	// Search all analyses
	summaries, err := cs.List()
	if err != nil {
		return nil, err
	}

	for _, summary := range summaries {
		analysis, err := cs.Load(summary.Hash)
		if err != nil {
			continue
		}

		// Apply filters
		if !opts.Since.IsZero() && analysis.AnalyzedAt.Before(opts.Since) {
			continue
		}
		if !opts.Until.IsZero() && analysis.AnalyzedAt.After(opts.Until) {
			continue
		}
		if opts.Author != "" && analysis.Author != opts.Author {
			continue
		}

		// Search in various fields
		matches := cs.searchAnalysis(analysis, query, opts)
		results = append(results, matches...)
	}

	// Sort by score
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if opts.Limit > 0 && len(results) > opts.Limit {
		results = results[:opts.Limit]
	}

	return results, nil
}

// searchAnalysis searches a single analysis for matches.
func (cs *CommitStore) searchAnalysis(analysis *CommitAnalysis, query string, opts RecallOptions) []RecallResult {
	var results []RecallResult

	// Check commit message
	if query != "" && strings.Contains(strings.ToLower(analysis.CommitMsg), query) {
		results = append(results, RecallResult{
			CommitHash: analysis.CommitHash,
			CommitMsg:  analysis.CommitMsg,
			Author:     analysis.Author,
			AnalyzedAt: analysis.AnalyzedAt,
			MatchType:  "commit",
			Snippet:    analysis.CommitMsg,
			Score:      0.8,
		})
	}

	// Check files and issues
	for _, file := range analysis.Files {
		if opts.FilePath != "" && !strings.Contains(file.Path, opts.FilePath) {
			continue
		}

		for _, issue := range file.Issues {
			if opts.Severity != "" && issue.Severity != opts.Severity {
				continue
			}

			issueText := strings.ToLower(issue.Message + " " + issue.Suggestion)
			if query == "" || strings.Contains(issueText, query) {
				results = append(results, RecallResult{
					CommitHash: analysis.CommitHash,
					CommitMsg:  analysis.CommitMsg,
					Author:     analysis.Author,
					AnalyzedAt: analysis.AnalyzedAt,
					FilePath:   file.Path,
					MatchType:  "issue",
					Snippet:    fmt.Sprintf("[%s] %s", issue.Severity, issue.Message),
					Score:      0.9,
				})
			}
		}
	}

	return results
}

// GetFileHistory returns the analysis history for a specific file.
func (cs *CommitStore) GetFileHistory(filePath string) (*CommitHistory, error) {
	summaries, err := cs.List()
	if err != nil {
		return nil, err
	}

	var relevantCommits []CommitSummary
	issuesByType := make(map[string]int)
	var totalIssues int

	for _, summary := range summaries {
		analysis, err := cs.Load(summary.Hash)
		if err != nil {
			continue
		}

		for _, file := range analysis.Files {
			if strings.Contains(file.Path, filePath) || strings.Contains(filePath, file.Path) {
				for _, issue := range file.Issues {
					issuesByType[issue.Type]++
					totalIssues++
				}
				relevantCommits = append(relevantCommits, summary)
				break
			}
		}
	}

	trend := "stable"
	if len(relevantCommits) >= 3 {
		recent := relevantCommits[0].IssueCount
		older := relevantCommits[len(relevantCommits)-1].IssueCount
		if recent < older {
			trend = "improving"
		} else if recent > older {
			trend = "worsening"
		}
	}

	return &CommitHistory{
		TotalCommits:    len(summaries),
		AnalyzedCommits: len(relevantCommits),
		Commits:         relevantCommits,
		IssueStats: IssueStats{
			TotalAnalyses:  len(relevantCommits),
			TotalIssues:    totalIssues,
			TrendDirection: trend,
		},
	}, nil
}

// Delete removes a commit analysis.
func (cs *CommitStore) Delete(commitHash string) error {
	shortHash := commitHash
	if len(shortHash) > 7 {
		shortHash = shortHash[:7]
	}
	return os.RemoveAll(filepath.Join(cs.baseDir, shortHash))
}

// Prune removes analyses older than the given duration.
func (cs *CommitStore) Prune(maxAge time.Duration) (int, error) {
	summaries, err := cs.List()
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-maxAge)
	pruned := 0

	for _, summary := range summaries {
		if summary.AnalyzedAt.Before(cutoff) {
			if err := cs.Delete(summary.Hash); err == nil {
				pruned++
			}
		}
	}

	return pruned, nil
}

// generateMarkdownSummary creates a human-readable markdown summary.
func (cs *CommitStore) generateMarkdownSummary(analysis *CommitAnalysis, path string) error {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Analysis: %s\n\n", analysis.CommitHash[:7]))
	sb.WriteString(fmt.Sprintf("**Commit:** %s\n", analysis.CommitMsg))
	sb.WriteString(fmt.Sprintf("**Author:** %s\n", analysis.Author))
	sb.WriteString(fmt.Sprintf("**Analyzed:** %s\n\n", analysis.AnalyzedAt.Format(time.RFC3339)))

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Files:** %d\n", analysis.Summary.TotalFiles))
	sb.WriteString(fmt.Sprintf("- **Issues:** %d\n", analysis.Summary.TotalIssues))
	sb.WriteString(fmt.Sprintf("- **Score:** %.1f/100\n", analysis.Summary.OverallScore))

	if len(analysis.Summary.BySeverity) > 0 {
		sb.WriteString("\n### By Severity\n")
		for sev, count := range analysis.Summary.BySeverity {
			sb.WriteString(fmt.Sprintf("- %s: %d\n", sev, count))
		}
	}

	if analysis.Summary.Recommendation != "" {
		sb.WriteString(fmt.Sprintf("\n### Recommendation\n%s\n", analysis.Summary.Recommendation))
	}

	sb.WriteString("\n## Files\n\n")
	for _, file := range analysis.Files {
		sb.WriteString(fmt.Sprintf("### %s\n", file.Path))
		sb.WriteString(fmt.Sprintf("- Language: %s\n", file.Language))
		sb.WriteString(fmt.Sprintf("- Changes: +%d/-%d\n", file.LinesAdded, file.LinesRemoved))

		if len(file.Issues) > 0 {
			sb.WriteString("\n**Issues:**\n")
			for _, issue := range file.Issues {
				sb.WriteString(fmt.Sprintf("- [%s] Line %d: %s\n", issue.Severity, issue.Line, issue.Message))
				if issue.Suggestion != "" {
					sb.WriteString(fmt.Sprintf("  - *Suggestion:* %s\n", issue.Suggestion))
				}
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Context\n\n")
	sb.WriteString(fmt.Sprintf("- Provider: %s\n", analysis.Context.Provider))
	sb.WriteString(fmt.Sprintf("- Model: %s\n", analysis.Context.Model))
	if analysis.Context.Personality != "" {
		sb.WriteString(fmt.Sprintf("- Personality: %s\n", analysis.Context.Personality))
	}
	if len(analysis.Context.Modes) > 0 {
		sb.WriteString(fmt.Sprintf("- Modes: %s\n", strings.Join(analysis.Context.Modes, ", ")))
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// formatAnalysisSummary formats an analysis for display.
func formatAnalysisSummary(analysis *CommitAnalysis) string {
	return fmt.Sprintf(
		"%s: %d issues (%d files) - Score: %.1f",
		analysis.CommitMsg,
		analysis.Summary.TotalIssues,
		analysis.Summary.TotalFiles,
		analysis.Summary.OverallScore,
	)
}

// GetCurrentBranch returns the current git branch name.
func GetCurrentBranch(repoRoot string) string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetCommitInfo retrieves commit information.
func GetCommitInfo(repoRoot, hash string) (msg, author, email string, err error) {
	cmd := exec.Command("git", "log", "-1", "--format=%s|%an|%ae", hash)
	cmd.Dir = repoRoot
	output, err := cmd.Output()
	if err != nil {
		return "", "", "", err
	}

	parts := strings.SplitN(strings.TrimSpace(string(output)), "|", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("unexpected git output")
	}

	return parts[0], parts[1], parts[2], nil
}
