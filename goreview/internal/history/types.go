// Package history provides SQLite-based storage for review history.
// This complements the cognitive memory system (BadgerDB) by offering
// full-text search capabilities for CLI commands like `goreview search`.
package history

import "time"

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
	Author       string `json:"author"`
	TotalIssues  int64  `json:"total_issues"`
	ResolvedRate float64 `json:"resolved_rate"`
}
