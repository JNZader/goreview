// Package history provides SQLite-based storage for review history.
// This complements the cognitive memory system (BadgerDB) by offering
// full-text search capabilities for CLI commands like `goreview search`.
//
// It also provides commit-level historical reasoning storage in
// .git/goreview/commits/ for recall and analysis over time.
package history

import "time"

// CommitAnalysis represents the detailed analysis of a single commit.
// Stored in .git/goreview/commits/<hash>/
type CommitAnalysis struct {
	CommitHash  string          `json:"commit_hash"`
	CommitMsg   string          `json:"commit_message"`
	Author      string          `json:"author"`
	AuthorEmail string          `json:"author_email"`
	AnalyzedAt  time.Time       `json:"analyzed_at"`
	Branch      string          `json:"branch,omitempty"`
	Files       []AnalyzedFile  `json:"files"`
	Summary     AnalysisSummary `json:"summary"`
	Context     AnalysisContext `json:"context"`
}

// AnalyzedFile represents analysis of a single file in a commit.
type AnalyzedFile struct {
	Path         string  `json:"path"`
	Language     string  `json:"language"`
	LinesAdded   int     `json:"lines_added"`
	LinesRemoved int     `json:"lines_removed"`
	Issues       []Issue `json:"issues"`
}

// Issue represents a detected issue in the analysis.
type Issue struct {
	ID         string     `json:"id"`
	Type       string     `json:"type"`
	Severity   string     `json:"severity"`
	Message    string     `json:"message"`
	Suggestion string     `json:"suggestion,omitempty"`
	Line       int        `json:"line,omitempty"`
	EndLine    int        `json:"end_line,omitempty"`
	RuleID     string     `json:"rule_id,omitempty"`
	RootCause  *RootCause `json:"root_cause,omitempty"`
}

// RootCause represents root cause tracing information.
type RootCause struct {
	Description string   `json:"description"`
	SourceLine  int      `json:"source_line"`
	Propagation []string `json:"propagation"`
}

// AnalysisSummary contains aggregate statistics.
type AnalysisSummary struct {
	TotalFiles     int            `json:"total_files"`
	TotalIssues    int            `json:"total_issues"`
	BySeverity     map[string]int `json:"by_severity"`
	ByType         map[string]int `json:"by_type"`
	OverallScore   float64        `json:"overall_score"`
	Recommendation string         `json:"recommendation,omitempty"`
}

// AnalysisContext contains the context used during analysis.
type AnalysisContext struct {
	Provider      string            `json:"provider"`
	Model         string            `json:"model"`
	Personality   string            `json:"personality,omitempty"`
	Modes         []string          `json:"modes,omitempty"`
	RAGSources    []string          `json:"rag_sources,omitempty"`
	CustomContext string            `json:"custom_context,omitempty"`
	Config        map[string]string `json:"config,omitempty"`
}

// RecallResult represents a search match in historical commit data.
type RecallResult struct {
	CommitHash string    `json:"commit_hash"`
	CommitMsg  string    `json:"commit_message"`
	Author     string    `json:"author"`
	AnalyzedAt time.Time `json:"analyzed_at"`
	FilePath   string    `json:"file_path,omitempty"`
	MatchType  string    `json:"match_type"` // "commit", "file", "issue", "content"
	Snippet    string    `json:"snippet"`
	Score      float64   `json:"score"`
}

// CommitHistory represents the analysis history of commits.
type CommitHistory struct {
	TotalCommits    int             `json:"total_commits"`
	AnalyzedCommits int             `json:"analyzed_commits"`
	Commits         []CommitSummary `json:"commits"`
	IssueStats      IssueStats      `json:"issue_stats"`
}

// CommitSummary is a brief summary of a commit for history views.
type CommitSummary struct {
	Hash       string         `json:"hash"`
	Message    string         `json:"message"`
	Author     string         `json:"author"`
	AnalyzedAt time.Time      `json:"analyzed_at"`
	IssueCount int            `json:"issue_count"`
	Severities map[string]int `json:"severities"`
}

// IssueStats aggregates issue statistics over time.
type IssueStats struct {
	TotalAnalyses   int              `json:"total_analyses"`
	TotalIssues     int              `json:"total_issues"`
	RecurringIssues []RecurringIssue `json:"recurring_issues"`
	TrendDirection  string           `json:"trend_direction"` // "improving", "stable", "worsening"
}

// RecurringIssue represents an issue that appears multiple times.
type RecurringIssue struct {
	Type         string    `json:"type"`
	Message      string    `json:"message"`
	Occurrences  int       `json:"occurrences"`
	FirstSeen    time.Time `json:"first_seen"`
	LastSeen     time.Time `json:"last_seen"`
	CommitHashes []string  `json:"commit_hashes"`
}

// RecallOptions configures the recall command.
type RecallOptions struct {
	Query      string    `json:"query,omitempty"`
	CommitHash string    `json:"commit_hash,omitempty"`
	FilePath   string    `json:"file_path,omitempty"`
	Author     string    `json:"author,omitempty"`
	Since      time.Time `json:"since,omitempty"`
	Until      time.Time `json:"until,omitempty"`
	Severity   string    `json:"severity,omitempty"`
	Limit      int       `json:"limit,omitempty"`
}

// ReviewRecord represents a stored review entry.
type ReviewRecord struct {
	ID          int64     `json:"id"`
	CommitHash  string    `json:"commit_hash"`
	FilePath    string    `json:"file_path"`
	IssueType   string    `json:"issue_type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Suggestion  string    `json:"suggestion,omitempty"`
	Line        int       `json:"line,omitempty"`
	Author      string    `json:"author,omitempty"`
	Branch      string    `json:"branch,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	Resolved    bool      `json:"resolved"`
	ResolvedAt  time.Time `json:"resolved_at,omitempty"`
	ReviewRound int       `json:"review_round"`
}

// SearchQuery represents a search query for review history.
type SearchQuery struct {
	// Text performs full-text search on message and suggestion
	Text string
	// File filters by file path (supports glob patterns)
	File string
	// Author filters by commit author
	Author string
	// Severity filters by issue severity
	Severity string
	// Type filters by issue type
	Type string
	// Branch filters by git branch
	Branch string
	// Since filters by creation date
	Since time.Time
	// Until filters by creation date
	Until time.Time
	// Resolved filters by resolved status (nil = all)
	Resolved *bool
	// Limit restricts result count
	Limit int
	// Offset for pagination
	Offset int
}

// SearchResult contains search results with metadata.
type SearchResult struct {
	Records    []ReviewRecord `json:"records"`
	TotalCount int64          `json:"total_count"`
	Query      SearchQuery    `json:"-"`
}

// FileHistory contains the review history for a specific file or directory.
type FileHistory struct {
	Path         string         `json:"path"`
	TotalIssues  int64          `json:"total_issues"`
	Resolved     int64          `json:"resolved"`
	Pending      int64          `json:"pending"`
	BySeverity   map[string]int `json:"by_severity"`
	ByType       map[string]int `json:"by_type"`
	FirstReview  time.Time      `json:"first_review"`
	LastReview   time.Time      `json:"last_review"`
	ReviewRounds int            `json:"review_rounds"`
}

// Stats contains aggregate statistics from the history database.
type Stats struct {
	TotalReviews   int64            `json:"total_reviews"`
	TotalIssues    int64            `json:"total_issues"`
	ResolvedIssues int64            `json:"resolved_issues"`
	BySeverity     map[string]int64 `json:"by_severity"`
	ByType         map[string]int64 `json:"by_type"`
	ByFile         map[string]int64 `json:"by_file"`
	TopAuthors     []AuthorStats    `json:"top_authors,omitempty"`
}

// AuthorStats contains statistics for an author.
type AuthorStats struct {
	Author       string  `json:"author"`
	TotalIssues  int64   `json:"total_issues"`
	ResolvedRate float64 `json:"resolved_rate"`
}
