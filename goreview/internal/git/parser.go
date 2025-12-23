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

// parseState holds state during regex-based diff parsing
type parseState struct {
	diff        *Diff
	currentFile *FileDiff
	currentHunk *Hunk
}

// ParseDiff parses a unified diff string into a Diff struct.
func ParseDiff(diffText string) (*Diff, error) {
	diff := &Diff{
		Files: make([]FileDiff, 0),
	}

	if strings.TrimSpace(diffText) == "" {
		return diff, nil
	}

	state := &parseState{diff: diff}
	lines := strings.Split(diffText, "\n")

	for i := 0; i < len(lines); i++ {
		state.processLine(lines[i])
	}

	state.finalizeFile()
	diff.CalculateStats()
	return diff, nil
}

// processLine handles a single line during parsing
func (s *parseState) processLine(line string) {
	// Check for new file diff
	if matches := diffHeaderRegex.FindStringSubmatch(line); matches != nil {
		s.startNewFile(matches)
		return
	}

	if s.currentFile == nil {
		return
	}

	// Check for file status indicators
	if s.checkFileStatus(line) {
		return
	}

	// Check for hunk header
	if matches := hunkHeaderRegex.FindStringSubmatch(line); matches != nil {
		s.startNewHunk(line, matches)
		return
	}

	// Parse diff lines
	s.addDiffLine(line)
}

// startNewFile begins parsing a new file
func (s *parseState) startNewFile(matches []string) {
	s.finalizeFile()

	s.currentFile = &FileDiff{
		Path:     matches[2],
		OldPath:  matches[1],
		Status:   FileModified,
		Language: detectLanguage(matches[2]),
		Hunks:    make([]Hunk, 0),
	}
	s.currentHunk = nil
}

// checkFileStatus checks and handles file status indicators
func (s *parseState) checkFileStatus(line string) bool {
	switch {
	case newFileRegex.MatchString(line):
		s.currentFile.Status = FileAdded
	case deletedRegex.MatchString(line):
		s.currentFile.Status = FileDeleted
	case renameFromRegex.MatchString(line):
		s.currentFile.Status = FileRenamed
	case binaryFileRegex.MatchString(line):
		s.currentFile.IsBinary = true
	default:
		return false
	}
	return true
}

// startNewHunk begins parsing a new hunk
func (s *parseState) startNewHunk(line string, matches []string) {
	if s.currentHunk != nil {
		s.currentFile.Hunks = append(s.currentFile.Hunks, *s.currentHunk)
	}

	s.currentHunk = &Hunk{
		Header:   line,
		OldStart: mustParseInt(matches[1]),
		OldLines: parseIntOrDefault(matches[2], 1),
		NewStart: mustParseInt(matches[3]),
		NewLines: parseIntOrDefault(matches[4], 1),
		Lines:    make([]Line, 0),
	}
}

// addDiffLine adds a content line to the current hunk
func (s *parseState) addDiffLine(line string) {
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
		return // "\ No newline at end of file" - skip
	}

	s.currentHunk.Lines = append(s.currentHunk.Lines, Line{
		Type:    lineType,
		Content: content,
	})
}

// finalizeFile saves the current file and hunk to the diff
func (s *parseState) finalizeFile() {
	if s.currentFile != nil {
		if s.currentHunk != nil {
			s.currentFile.Hunks = append(s.currentFile.Hunks, *s.currentHunk)
		}
		s.diff.Files = append(s.diff.Files, *s.currentFile)
	}
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
