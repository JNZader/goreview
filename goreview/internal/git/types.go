// Package git provides Git repository operations for goreview.
package git

import "context"

// Repository defines the interface for Git operations.
// This abstraction allows for testing with mock implementations.
type Repository interface {
	// GetStagedDiff returns the diff of staged changes.
	GetStagedDiff(ctx context.Context) (*Diff, error)

	// GetCommitDiff returns the diff of a specific commit.
	GetCommitDiff(ctx context.Context, sha string) (*Diff, error)

	// GetBranchDiff returns the diff between current branch and base branch.
	GetBranchDiff(ctx context.Context, baseBranch string) (*Diff, error)

	// GetFileDiff returns the diff for specific files.
	GetFileDiff(ctx context.Context, files []string) (*Diff, error)

	// GetCurrentBranch returns the current branch name.
	GetCurrentBranch(ctx context.Context) (string, error)

	// GetRepoRoot returns the root directory of the repository.
	GetRepoRoot(ctx context.Context) (string, error)

	// IsClean returns true if there are no uncommitted changes.
	IsClean(ctx context.Context) (bool, error)
}

// Diff represents a complete diff with multiple files.
type Diff struct {
	Files []FileDiff `json:"files"`
	Stats DiffStats  `json:"stats"`
}

// FileDiff represents the diff for a single file.
type FileDiff struct {
	Path      string     `json:"path"`
	OldPath   string     `json:"old_path,omitempty"` // For renames
	Status    FileStatus `json:"status"`
	Language  string     `json:"language"`
	IsBinary  bool       `json:"is_binary"`
	Hunks     []Hunk     `json:"hunks"`
	Additions int        `json:"additions"`
	Deletions int        `json:"deletions"`
}

// FileStatus represents the status of a file in the diff.
type FileStatus string

const (
	FileAdded    FileStatus = "added"
	FileModified FileStatus = "modified"
	FileDeleted  FileStatus = "deleted"
	FileRenamed  FileStatus = "renamed"
	FileCopied   FileStatus = "copied"
)

// Hunk represents a section of changes in a file.
type Hunk struct {
	Header   string `json:"header"` // @@ -start,count +start,count @@
	OldStart int    `json:"old_start"`
	OldLines int    `json:"old_lines"`
	NewStart int    `json:"new_start"`
	NewLines int    `json:"new_lines"`
	Lines    []Line `json:"lines"`
}

// Line represents a single line in a hunk.
type Line struct {
	Type      LineType `json:"type"`
	Content   string   `json:"content"`
	OldNumber int      `json:"old_number,omitempty"`
	NewNumber int      `json:"new_number,omitempty"`
}

// LineType represents the type of a diff line.
type LineType string

const (
	LineContext  LineType = "context"
	LineAddition LineType = "addition"
	LineDeletion LineType = "deletion"
)

// DiffStats contains summary statistics about a diff.
type DiffStats struct {
	FilesChanged int `json:"files_changed"`
	Additions    int `json:"additions"`
	Deletions    int `json:"deletions"`
}

// CalculateStats calculates statistics from the diff.
func (d *Diff) CalculateStats() {
	d.Stats = DiffStats{
		FilesChanged: len(d.Files),
	}
	for _, f := range d.Files {
		d.Stats.Additions += f.Additions
		d.Stats.Deletions += f.Deletions
	}
}

// Commit represents a git commit for changelog generation.
type Commit struct {
	Hash        string `json:"hash"`
	ShortHash   string `json:"short_hash"`
	Subject     string `json:"subject"`
	Body        string `json:"body,omitempty"`
	Author      string `json:"author"`
	AuthorEmail string `json:"author_email"`
	Date        string `json:"date"`
}

// Tag represents a git tag.
type Tag struct {
	Name   string `json:"name"`
	Hash   string `json:"hash"`
	Date   string `json:"date"`
	Tagger string `json:"tagger,omitempty"`
}

// ConventionalCommit represents a parsed conventional commit.
type ConventionalCommit struct {
	Type        string `json:"type"`
	Scope       string `json:"scope,omitempty"`
	Breaking    bool   `json:"breaking"`
	Description string `json:"description"`
	Body        string `json:"body,omitempty"`
	Footer      string `json:"footer,omitempty"`
	Hash        string `json:"hash"`
	ShortHash   string `json:"short_hash"`
	Author      string `json:"author"`
	Date        string `json:"date"`
}
