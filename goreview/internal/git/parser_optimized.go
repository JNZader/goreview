package git

import (
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

// Pool of Line slices to reduce allocations
var lineSlicePool = sync.Pool{
	New: func() interface{} {
		s := make([]Line, 0, 64)
		return &s
	},
}

// getLineSlice gets a Line slice from the pool
func getLineSlice() *[]Line {
	return lineSlicePool.Get().(*[]Line)
}

// putLineSlice returns a Line slice to the pool
func putLineSlice(s *[]Line) {
	*s = (*s)[:0] // Reset length but keep capacity
	if cap(*s) < 1024 {
		lineSlicePool.Put(s)
	}
}

// diffParseState holds the state during diff parsing
type diffParseState struct {
	diff        *Diff
	currentFile *FileDiff
	currentHunk *Hunk
}

// ParseDiffOptimized parses a unified diff string with optimizations.
// It uses less memory allocations and avoids regex where possible.
func ParseDiffOptimized(diffText string) (*Diff, error) {
	diff := &Diff{}

	if len(diffText) == 0 || strings.TrimSpace(diffText) == "" {
		diff.Files = make([]FileDiff, 0)
		return diff, nil
	}

	// Estimate number of files
	estimatedFiles := strings.Count(diffText, "diff --git")
	if estimatedFiles == 0 {
		estimatedFiles = 1
	}
	diff.Files = make([]FileDiff, 0, estimatedFiles)

	state := &diffParseState{diff: diff}

	// Parse line by line without creating full slice
	start := 0
	for i := 0; i <= len(diffText); i++ {
		if i == len(diffText) || diffText[i] == '\n' {
			line := diffText[start:i]
			start = i + 1

			if len(line) == 0 {
				continue
			}

			state.parseLine(line)
		}
	}

	// Add last file and hunk
	state.finalize()

	diff.CalculateStats()
	return diff, nil
}

// parseLine handles parsing a single line of the diff
func (s *diffParseState) parseLine(line string) {
	// Check for diff header (most common, so check first)
	if strings.HasPrefix(line, "diff --git ") {
		s.handleNewFile(line)
		return
	}

	if s.currentFile == nil {
		return
	}

	// Check for hunk header
	if strings.HasPrefix(line, "@@") {
		s.handleHunkHeader(line)
		return
	}

	// Check for file status
	if s.handleFileStatus(line) {
		return
	}

	// Parse diff line content
	s.handleDiffLine(line)
}

// handleNewFile processes a new file header
func (s *diffParseState) handleNewFile(line string) {
	// Save previous file
	if s.currentFile != nil {
		if s.currentHunk != nil {
			s.currentFile.Hunks = append(s.currentFile.Hunks, *s.currentHunk)
		}
		s.diff.Files = append(s.diff.Files, *s.currentFile)
	}

	// Parse: "diff --git a/path b/path"
	oldPath, newPath := parseDiffGitLine(line)
	s.currentFile = &FileDiff{
		Path:     newPath,
		OldPath:  oldPath,
		Status:   FileModified,
		Language: detectLanguageOptimized(newPath),
		Hunks:    make([]Hunk, 0, 4),
	}
	s.currentHunk = nil
}

// handleHunkHeader processes a hunk header line
func (s *diffParseState) handleHunkHeader(line string) {
	if s.currentHunk != nil {
		s.currentFile.Hunks = append(s.currentFile.Hunks, *s.currentHunk)
	}
	s.currentHunk = parseHunkHeaderOptimized(line)
}

// handleFileStatus checks and handles file status lines
func (s *diffParseState) handleFileStatus(line string) bool {
	switch {
	case strings.HasPrefix(line, "new file"):
		s.currentFile.Status = FileAdded
	case strings.HasPrefix(line, "deleted file"):
		s.currentFile.Status = FileDeleted
	case strings.HasPrefix(line, "rename from"):
		s.currentFile.Status = FileRenamed
	case strings.HasPrefix(line, "Binary files"):
		s.currentFile.IsBinary = true
	default:
		return false
	}
	return true
}

// handleDiffLine processes a diff content line
func (s *diffParseState) handleDiffLine(line string) {
	if s.currentHunk == nil || len(line) == 0 {
		return
	}

	lineType := LineContext
	content := line

	switch line[0] {
	case '+':
		lineType = LineAddition
		content = line[1:]
		s.currentFile.Additions++
	case '-':
		lineType = LineDeletion
		content = line[1:]
		s.currentFile.Deletions++
	case ' ':
		content = line[1:]
	case '\\':
		return // No newline at end of file
	}

	s.currentHunk.Lines = append(s.currentHunk.Lines, Line{
		Type:    lineType,
		Content: content,
	})
}

// finalize adds the last file and hunk to the diff
func (s *diffParseState) finalize() {
	if s.currentFile != nil {
		if s.currentHunk != nil {
			s.currentFile.Hunks = append(s.currentFile.Hunks, *s.currentHunk)
		}
		s.diff.Files = append(s.diff.Files, *s.currentFile)
	}
}

// parseDiffGitLine extracts paths from "diff --git a/path b/path"
func parseDiffGitLine(line string) (oldPath, newPath string) {
	// Skip "diff --git "
	const prefix = "diff --git "
	if !strings.HasPrefix(line, prefix) {
		return "", ""
	}

	rest := line[len(prefix):]

	// Find " b/" separator
	idx := strings.Index(rest, " b/")
	if idx == -1 {
		return "", ""
	}

	// Extract paths (skip "a/" and " b/")
	if len(rest) > 2 && rest[0] == 'a' && rest[1] == '/' {
		oldPath = rest[2:idx]
	} else {
		oldPath = rest[:idx]
	}
	newPath = rest[idx+3:]

	return oldPath, newPath
}

// parseHunkHeaderOptimized parses "@@ -1,10 +1,12 @@" without regex
func parseHunkHeaderOptimized(line string) *Hunk {
	hunk := &Hunk{
		Header:   line,
		Lines:    make([]Line, 0, 32),
		OldLines: 1, // Default
		NewLines: 1, // Default
	}

	// Find first @@ and second @@
	if !strings.HasPrefix(line, "@@ ") {
		return hunk
	}

	// Parse: @@ -oldStart,oldLines +newStart,newLines @@
	// Skip "@@ "
	rest := line[3:]

	// Find the closing "@@"
	endIdx := strings.Index(rest, " @@")
	if endIdx == -1 {
		endIdx = len(rest)
	}
	rangeStr := rest[:endIdx]

	// Split on space to get old and new ranges
	parts := strings.Split(rangeStr, " ")
	if len(parts) >= 2 {
		// Parse old range (-1,10)
		oldRange := parts[0]
		if strings.HasPrefix(oldRange, "-") {
			parseRange(oldRange[1:], &hunk.OldStart, &hunk.OldLines)
		}

		// Parse new range (+1,12)
		newRange := parts[1]
		if strings.HasPrefix(newRange, "+") {
			parseRange(newRange[1:], &hunk.NewStart, &hunk.NewLines)
		}
	}

	return hunk
}

// parseRange parses "start,count" or "start"
func parseRange(s string, start, count *int) {
	idx := strings.Index(s, ",")
	if idx == -1 {
		if v, err := strconv.Atoi(s); err == nil {
			*start = v
		}
		*count = 1
	} else {
		if v, err := strconv.Atoi(s[:idx]); err == nil {
			*start = v
		}
		if v, err := strconv.Atoi(s[idx+1:]); err == nil {
			*count = v
		}
	}
}

// extToLanguage maps file extensions to language names
var extToLanguage = map[string]string{
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

// detectLanguageOptimized detects language from file extension with faster lookup
func detectLanguageOptimized(path string) string {
	ext := extractExtension(path)
	if ext == "" {
		return "unknown"
	}

	if lang, ok := extToLanguage[ext]; ok {
		return lang
	}
	return "unknown"
}

// extractExtension extracts the lowercase extension from a path
func extractExtension(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return strings.ToLower(path[i:])
		}
		if path[i] == '/' || path[i] == '\\' {
			break
		}
	}
	return ""
}

// Unused function to satisfy the import
var _ = filepath.Ext
