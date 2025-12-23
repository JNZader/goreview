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

	var currentFile *FileDiff
	var currentHunk *Hunk

	// Parse line by line without creating full slice
	start := 0
	for i := 0; i <= len(diffText); i++ {
		if i == len(diffText) || diffText[i] == '\n' {
			line := diffText[start:i]
			start = i + 1

			if len(line) == 0 {
				continue
			}

			// Check for diff header (most common, so check first)
			if strings.HasPrefix(line, "diff --git ") {
				// Save previous file
				if currentFile != nil {
					if currentHunk != nil {
						currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
					}
					diff.Files = append(diff.Files, *currentFile)
				}

				// Parse: "diff --git a/path b/path"
				oldPath, newPath := parseDiffGitLine(line)
				currentFile = &FileDiff{
					Path:     newPath,
					OldPath:  oldPath,
					Status:   FileModified,
					Language: detectLanguageOptimized(newPath),
					Hunks:    make([]Hunk, 0, 4),
				}
				currentHunk = nil
				continue
			}

			if currentFile == nil {
				continue
			}

			// Check for hunk header
			if strings.HasPrefix(line, "@@") {
				if currentHunk != nil {
					currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
				}
				currentHunk = parseHunkHeaderOptimized(line)
				continue
			}

			// Check for file status
			if strings.HasPrefix(line, "new file") {
				currentFile.Status = FileAdded
				continue
			}
			if strings.HasPrefix(line, "deleted file") {
				currentFile.Status = FileDeleted
				continue
			}
			if strings.HasPrefix(line, "rename from") {
				currentFile.Status = FileRenamed
				continue
			}
			if strings.HasPrefix(line, "Binary files") {
				currentFile.IsBinary = true
				continue
			}

			// Parse diff line content
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
					continue // No newline at end of file
				}

				currentHunk.Lines = append(currentHunk.Lines, Line{
					Type:    lineType,
					Content: content,
				})
			}
		}
	}

	// Add last file and hunk
	if currentFile != nil {
		if currentHunk != nil {
			currentFile.Hunks = append(currentFile.Hunks, *currentHunk)
		}
		diff.Files = append(diff.Files, *currentFile)
	}

	diff.CalculateStats()
	return diff, nil
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

// detectLanguageOptimized detects language from file extension with faster lookup
func detectLanguageOptimized(path string) string {
	// Get extension manually to avoid filepath.Ext allocation
	ext := ""
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			ext = strings.ToLower(path[i:])
			break
		}
		if path[i] == '/' || path[i] == '\\' {
			break
		}
	}

	if ext == "" {
		return "unknown"
	}

	// Use switch for faster lookup than map
	switch ext {
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".tsx":
		return "typescript"
	case ".jsx":
		return "javascript"
	case ".java":
		return "java"
	case ".rb":
		return "ruby"
	case ".rs":
		return "rust"
	case ".c":
		return "c"
	case ".cpp":
		return "cpp"
	case ".h":
		return "c"
	case ".hpp":
		return "cpp"
	case ".cs":
		return "csharp"
	case ".php":
		return "php"
	case ".swift":
		return "swift"
	case ".kt":
		return "kotlin"
	case ".scala":
		return "scala"
	case ".sh":
		return "shell"
	case ".bash":
		return "shell"
	case ".yaml":
		return "yaml"
	case ".yml":
		return "yaml"
	case ".json":
		return "json"
	case ".xml":
		return "xml"
	case ".html":
		return "html"
	case ".css":
		return "css"
	case ".scss":
		return "scss"
	case ".sql":
		return "sql"
	case ".md":
		return "markdown"
	default:
		return "unknown"
	}
}

// Unused function to satisfy the import
var _ = filepath.Ext
