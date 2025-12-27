// Package export provides export functionality for review results to external systems.
package export

import (
	"time"

	"github.com/JNZader/goreview/goreview/internal/review"
)

// Exporter defines the interface for exporting review results.
type Exporter interface {
	// Export exports a review result to the target destination.
	Export(result *review.Result, metadata *ExportMetadata) error

	// Name returns the exporter name.
	Name() string
}

// ExportMetadata contains metadata for the export.
type ExportMetadata struct {
	// ProjectName is the name of the project being reviewed
	ProjectName string

	// Branch is the current git branch
	Branch string

	// CommitHash is the full commit hash
	CommitHash string

	// CommitShort is the short commit hash (7 chars)
	CommitShort string

	// Author is the commit author
	Author string

	// ReviewDate is when the review was performed
	ReviewDate time.Time

	// ReviewMode is the review mode used (staged, commit, branch, files)
	ReviewMode string

	// BaseBranch is the base branch for branch mode
	BaseBranch string
}

// ObsidianFrontmatter represents YAML frontmatter for Obsidian notes.
type ObsidianFrontmatter struct {
	// Date is the review date in ISO format
	Date string `yaml:"date"`

	// Project is the project name
	Project string `yaml:"project"`

	// Branch is the git branch
	Branch string `yaml:"branch"`

	// Commit is the full commit hash
	Commit string `yaml:"commit"`

	// CommitShort is the short commit hash
	CommitShort string `yaml:"commit_short"`

	// Author is the commit author
	Author string `yaml:"author,omitempty"`

	// ReviewMode is the review mode used
	ReviewMode string `yaml:"review_mode"`

	// FilesReviewed is the number of files reviewed
	FilesReviewed int `yaml:"files_reviewed"`

	// TotalIssues is the total number of issues found
	TotalIssues int `yaml:"total_issues"`

	// CriticalIssues is the count of critical severity issues
	CriticalIssues int `yaml:"critical_issues"`

	// ErrorIssues is the count of error severity issues
	ErrorIssues int `yaml:"error_issues"`

	// WarningIssues is the count of warning severity issues
	WarningIssues int `yaml:"warning_issues"`

	// InfoIssues is the count of info severity issues
	InfoIssues int `yaml:"info_issues"`

	// AverageScore is the average quality score (0-100)
	AverageScore int `yaml:"average_score"`

	// Duration is the review duration as string
	Duration string `yaml:"duration"`

	// Tags are the Obsidian tags for this note
	Tags []string `yaml:"tags"`

	// Aliases are alternative names for linking
	Aliases []string `yaml:"aliases,omitempty"`

	// RelatedReviews are links to related review notes
	RelatedReviews []string `yaml:"related,omitempty"`
}
