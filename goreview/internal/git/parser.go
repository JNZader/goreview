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
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
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
