# Iteracion 03: Integracion con Git

## Objetivos

Al completar esta iteracion tendras:
- Interface Repository para abstraer operaciones Git
- Implementacion GitRepository usando comandos git
- Parser de diffs completo
- Soporte para staged, commit, branch y files
- Tests completos

## Prerequisitos

- Iteracion 02 completada
- Git instalado en el sistema

## Tiempo Estimado: 8 horas

---

## Commit 3.1: Crear interface Repository

**Mensaje de commit:**
```
feat(git): add repository interface

- Define Repository interface for git operations
- Add Diff, FileDiff, and Hunk types
- Add DiffStats for summary information
- Define line types (addition, deletion, context)
```

### `goreview/internal/git/types.go`

```go
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
	Path       string     `json:"path"`
	OldPath    string     `json:"old_path,omitempty"` // For renames
	Status     FileStatus `json:"status"`
	Language   string     `json:"language"`
	IsBinary   bool       `json:"is_binary"`
	Hunks      []Hunk     `json:"hunks"`
	Additions  int        `json:"additions"`
	Deletions  int        `json:"deletions"`
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
	Header     string `json:"header"`      // @@ -start,count +start,count @@
	OldStart   int    `json:"old_start"`
	OldLines   int    `json:"old_lines"`
	NewStart   int    `json:"new_start"`
	NewLines   int    `json:"new_lines"`
	Lines      []Line `json:"lines"`
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
```

---

## Commit 3.2: Implementar GitRepository

**Mensaje de commit:**
```
feat(git): add git repository implementation

- Implement Repository interface using git commands
- Execute git via os/exec for portability
- Handle command errors gracefully
- Add timeout support via context
```

### `goreview/internal/git/repository.go`

```go
package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitRepository implements Repository using git commands.
type GitRepository struct {
	path string
}

// NewGitRepository creates a new GitRepository.
func NewGitRepository(path string) (*GitRepository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify it's a git repository
	repo := &GitRepository{path: absPath}
	if _, err := repo.GetRepoRoot(context.Background()); err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	return repo, nil
}

// runGit executes a git command and returns the output.
func (r *GitRepository) runGit(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Include stderr in error message for debugging
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return "", fmt.Errorf("git %s: %w: %s", args[0], err, errMsg)
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}

	return stdout.String(), nil
}

func (r *GitRepository) GetStagedDiff(ctx context.Context) (*Diff, error) {
	// Get staged diff
	output, err := r.runGit(ctx, "diff", "--cached", "--unified=3")
	if err != nil {
		return nil, err
	}

	diff, err := ParseDiff(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	return diff, nil
}

func (r *GitRepository) GetCommitDiff(ctx context.Context, sha string) (*Diff, error) {
	output, err := r.runGit(ctx, "show", sha, "--unified=3", "--format=")
	if err != nil {
		return nil, err
	}

	return ParseDiff(output)
}

func (r *GitRepository) GetBranchDiff(ctx context.Context, baseBranch string) (*Diff, error) {
	// Get merge base
	mergeBase, err := r.runGit(ctx, "merge-base", baseBranch, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to find merge base: %w", err)
	}

	mergeBase = strings.TrimSpace(mergeBase)
	output, err := r.runGit(ctx, "diff", mergeBase, "HEAD", "--unified=3")
	if err != nil {
		return nil, err
	}

	return ParseDiff(output)
}

func (r *GitRepository) GetFileDiff(ctx context.Context, files []string) (*Diff, error) {
	args := append([]string{"diff", "--unified=3", "--"}, files...)
	output, err := r.runGit(ctx, args...)
	if err != nil {
		return nil, err
	}

	return ParseDiff(output)
}

func (r *GitRepository) GetCurrentBranch(ctx context.Context) (string, error) {
	output, err := r.runGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *GitRepository) GetRepoRoot(ctx context.Context) (string, error) {
	output, err := r.runGit(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *GitRepository) IsClean(ctx context.Context) (bool, error) {
	output, err := r.runGit(ctx, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) == "", nil
}
```

---

## Commit 3.3: Implementar parser de diffs

**Mensaje de commit:**
```
feat(git): add diff parser

- Parse unified diff format
- Extract file paths and status
- Parse hunks and lines
- Detect file language from extension
- Handle binary files
```

### `goreview/internal/git/parser.go`

```go
package git

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Regex patterns for diff parsing
	diffHeaderRegex = regexp.MustCompile(`^diff --git a/(.*) b/(.*)$`)
	hunkHeaderRegex = regexp.MustCompile(`^@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@`)
	binaryFileRegex = regexp.MustCompile(`^Binary files .* differ$`)
	newFileRegex    = regexp.MustCompile(`^new file mode`)
	deletedRegex    = regexp.MustCompile(`^deleted file mode`)
	renameFromRegex = regexp.MustCompile(`^rename from (.*)$`)
	renameToRegex   = regexp.MustCompile(`^rename to (.*)$`)
)

// ParseDiff parses a unified diff string into a Diff struct.
func ParseDiff(diffText string) (*Diff, error) {
	diff := &Diff{
		Files: make([]FileDiff, 0),
	}

	if strings.TrimSpace(diffText) == "" {
		return diff, nil
	}

	lines := strings.Split(diffText, "\n")
	var currentFile *FileDiff
	var currentHunk *Hunk

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Check for new file diff
		if matches := diffHeaderRegex.FindStringSubmatch(line); matches != nil {
			// Save previous file if exists
			if currentFile != nil {
				if currentHunk != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
				}
				diff.Files = append(diff.Files, *currentFile)
			}

			// Start new file
			currentFile = &FileDiff{
				Path:     matches[2],
				OldPath:  matches[1],
				Status:   FileModified,
				Language: detectLanguage(matches[2]),
				Hunks:    make([]Hunk, 0),
			}
			currentHunk = nil
			continue
		}

		if currentFile == nil {
			continue
		}

		// Check for file status indicators
		if newFileRegex.MatchString(line) {
			currentFile.Status = FileAdded
			continue
		}
		if deletedRegex.MatchString(line) {
			currentFile.Status = FileDeleted
			continue
		}
		if renameFromRegex.MatchString(line) {
			currentFile.Status = FileRenamed
			continue
		}
		if binaryFileRegex.MatchString(line) {
			currentFile.IsBinary = true
			continue
		}

		// Check for hunk header
		if matches := hunkHeaderRegex.FindStringSubmatch(line); matches != nil {
			// Save previous hunk
			if currentHunk != nil {
				currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
			}

			currentHunk = &Hunk{
				Header:   line,
				OldStart: mustParseInt(matches[1]),
				OldLines: parseIntOrDefault(matches[2], 1),
				NewStart: mustParseInt(matches[3]),
				NewLines: parseIntOrDefault(matches[4], 1),
				Lines:    make([]Line, 0),
			}
			continue
		}

		// Parse diff lines
		if currentHunk != nil && len(line) > 0 {
			lineType := LineContext
			content := line

			switch line[0] {
			case '+':
				lineType = LineAddition
				content = line[1:]
				currentFile.Additions++
			case '-':
				lineType = LineDeletion
				content = line[1:]
				currentFile.Deletions++
			case ' ':
				content = line[1:]
			case '\\':
				// "\ No newline at end of file" - skip
				continue
			}

			currentHunk.Lines = append(currentHunk.Lines, Line{
				Type:    lineType,
				Content: content,
			})
		}
	}

	// Don't forget the last file and hunk
	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		diff.Files = append(diff.Files, *currentFile)
	}

	diff.CalculateStats()
	return diff, nil
}

// detectLanguage detects the programming language from file extension.
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))

	languages := map[string]string{
		".go":    "go",
		".py":    "python",
		".js":    "javascript",
		".ts":    "typescript",
		".tsx":   "typescript",
		".jsx":   "javascript",
		".java":  "java",
		".rb":    "ruby",
		".rs":    "rust",
		".c":     "c",
		".cpp":   "cpp",
		".h":     "c",
		".hpp":   "cpp",
		".cs":    "csharp",
		".php":   "php",
		".swift": "swift",
		".kt":    "kotlin",
		".scala": "scala",
		".sh":    "shell",
		".bash":  "shell",
		".yaml":  "yaml",
		".yml":   "yaml",
		".json":  "json",
		".xml":   "xml",
		".html":  "html",
		".css":   "css",
		".scss":  "scss",
		".sql":   "sql",
		".md":    "markdown",
	}

	if lang, ok := languages[ext]; ok {
		return lang
	}
	return "unknown"
}

func mustParseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func parseIntOrDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
```

---

## Commit 3.4: Agregar tests de Git

**Mensaje de commit:**
```
test(git): add repository and parser tests

- Test diff parsing with various formats
- Test language detection
- Test file status detection
- Add mock data for testing
```

### `goreview/internal/git/parser_test.go`

```go
package git

import (
	"testing"
)

func TestParseDiff(t *testing.T) {
	diffText := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,6 @@
 package main

+import "fmt"
+
 func main() {
-    println("hello")
+    fmt.Println("hello")
 }
`

	diff, err := ParseDiff(diffText)
	if err != nil {
		t.Fatalf("ParseDiff() error = %v", err)
	}

	if len(diff.Files) != 1 {
		t.Errorf("len(Files) = %d, want 1", len(diff.Files))
	}

	file := diff.Files[0]
	if file.Path != "main.go" {
		t.Errorf("Path = %v, want main.go", file.Path)
	}

	if file.Language != "go" {
		t.Errorf("Language = %v, want go", file.Language)
	}

	if file.Additions != 3 {
		t.Errorf("Additions = %d, want 3", file.Additions)
	}

	if file.Deletions != 1 {
		t.Errorf("Deletions = %d, want 1", file.Deletions)
	}
}

func TestParseDiffNewFile(t *testing.T) {
	diffText := `diff --git a/new.go b/new.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/new.go
@@ -0,0 +1,3 @@
+package main
+
+func new() {}
`

	diff, err := ParseDiff(diffText)
	if err != nil {
		t.Fatalf("ParseDiff() error = %v", err)
	}

	if diff.Files[0].Status != FileAdded {
		t.Errorf("Status = %v, want added", diff.Files[0].Status)
	}
}

func TestParseDiffDeleted(t *testing.T) {
	diffText := `diff --git a/old.go b/old.go
deleted file mode 100644
index 1234567..0000000
--- a/old.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func old() {}
`

	diff, err := ParseDiff(diffText)
	if err != nil {
		t.Fatalf("ParseDiff() error = %v", err)
	}

	if diff.Files[0].Status != FileDeleted {
		t.Errorf("Status = %v, want deleted", diff.Files[0].Status)
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "go"},
		{"script.py", "python"},
		{"app.ts", "typescript"},
		{"Component.tsx", "typescript"},
		{"style.css", "css"},
		{"unknown.xyz", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := detectLanguage(tt.path)
			if got != tt.want {
				t.Errorf("detectLanguage(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestEmptyDiff(t *testing.T) {
	diff, err := ParseDiff("")
	if err != nil {
		t.Fatalf("ParseDiff() error = %v", err)
	}

	if len(diff.Files) != 0 {
		t.Errorf("len(Files) = %d, want 0", len(diff.Files))
	}
}
```

---

## Resumen de la Iteracion 03

### Commits:
1. `feat(git): add repository interface`
2. `feat(git): add git repository implementation`
3. `feat(git): add diff parser`
4. `test(git): add repository and parser tests`

### Archivos:
```
goreview/internal/git/
├── types.go
├── repository.go
├── parser.go
└── parser_test.go
```

### Verificacion:
```bash
cd goreview
go test -v ./internal/git/...
```

---

## Siguiente Iteracion

Continua con: **[04-PROVIDERS-IA.md](04-PROVIDERS-IA.md)**
