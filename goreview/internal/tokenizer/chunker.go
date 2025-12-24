package tokenizer

import (
	"regexp"
	"strings"
)

// Chunk represents a portion of code that can be reviewed independently
type Chunk struct {
	Content    string
	StartLine  int
	EndLine    int
	Type       ChunkType
	Name       string // Function/class name if applicable
	TokenCount int
}

// ChunkType represents the type of code chunk
type ChunkType int

const (
	ChunkTypeUnknown ChunkType = iota
	ChunkTypeFunction
	ChunkTypeClass
	ChunkTypeMethod
	ChunkTypeBlock
	ChunkTypeImports
	ChunkTypeRaw
)

func (t ChunkType) String() string {
	switch t {
	case ChunkTypeFunction:
		return "function"
	case ChunkTypeClass:
		return "class"
	case ChunkTypeMethod:
		return "method"
	case ChunkTypeBlock:
		return "block"
	case ChunkTypeImports:
		return "imports"
	case ChunkTypeRaw:
		return "raw"
	default:
		return "unknown"
	}
}

// ChunkerConfig configures the chunker
type ChunkerConfig struct {
	MaxChunkTokens int    // Maximum tokens per chunk
	Language       string // Programming language for better splitting
	Estimator      *Estimator
}

// Chunker splits code into reviewable chunks
type Chunker struct {
	config    ChunkerConfig
	estimator *Estimator
}

// NewChunker creates a new chunker
func NewChunker(cfg ChunkerConfig) *Chunker {
	if cfg.MaxChunkTokens <= 0 {
		cfg.MaxChunkTokens = 2000
	}
	if cfg.Estimator == nil {
		cfg.Estimator = NewEstimator()
	}
	return &Chunker{
		config:    cfg,
		estimator: cfg.Estimator,
	}
}

// ChunkDiff splits a diff into reviewable chunks
func (c *Chunker) ChunkDiff(diff string) []Chunk {
	lines := strings.Split(diff, "\n")

	// Try to split by function boundaries first
	chunks := c.splitByFunctions(lines)

	// If we couldn't find good boundaries, split by size
	if len(chunks) == 0 || (len(chunks) == 1 && c.estimator.EstimateTokens(chunks[0].Content) > c.config.MaxChunkTokens) {
		chunks = c.splitBySize(lines)
	}

	// Calculate token counts
	for i := range chunks {
		chunks[i].TokenCount = c.estimator.EstimateTokens(chunks[i].Content)
	}

	return chunks
}

// chunkState tracks state during function-based chunking
type chunkState struct {
	chunks       []Chunk
	currentChunk strings.Builder
	currentName  string
	currentStart int
	braceCount   int
	inFunction   bool
	lastType     ChunkType
}

// splitByFunctions attempts to split code at function boundaries
func (c *Chunker) splitByFunctions(lines []string) []Chunk {
	patterns := c.getFunctionPatterns()
	state := &chunkState{}

	for i, line := range lines {
		state.checkFunctionStart(patterns, line, i)
		state.updateBraceCount(line)
		state.currentChunk.WriteString(line)
		state.currentChunk.WriteString("\n")
		state.checkFunctionEnd(line, i)
	}

	state.finalizeRemaining(len(lines))
	return state.chunks
}

// checkFunctionStart checks if a line starts a new function and saves previous chunk
func (s *chunkState) checkFunctionStart(patterns []functionPattern, line string, lineNum int) {
	for _, p := range patterns {
		matches := p.pattern.FindStringSubmatch(line)
		if len(matches) == 0 {
			continue
		}
		if s.currentChunk.Len() > 0 && s.inFunction {
			s.saveCurrentChunk(lineNum-1, s.lastType)
		}
		s.inFunction = true
		s.currentStart = lineNum
		s.lastType = p.chunkType
		s.currentName = ""
		if len(matches) > 1 {
			s.currentName = matches[1]
		}
		break
	}
}

// updateBraceCount tracks brace balance for function end detection
func (s *chunkState) updateBraceCount(line string) {
	s.braceCount += strings.Count(line, "{") - strings.Count(line, "}")
}

// checkFunctionEnd checks if a function ended and saves the chunk
func (s *chunkState) checkFunctionEnd(line string, lineNum int) {
	if s.inFunction && s.braceCount == 0 && strings.Contains(line, "}") {
		s.saveCurrentChunk(lineNum, ChunkTypeFunction)
		s.inFunction = false
		s.currentName = ""
	}
}

// saveCurrentChunk saves the current chunk and resets the builder
func (s *chunkState) saveCurrentChunk(endLine int, chunkType ChunkType) {
	s.chunks = append(s.chunks, Chunk{
		Content:   s.currentChunk.String(),
		StartLine: s.currentStart,
		EndLine:   endLine,
		Type:      chunkType,
		Name:      s.currentName,
	})
	s.currentChunk.Reset()
}

// finalizeRemaining saves any remaining content as a chunk
func (s *chunkState) finalizeRemaining(totalLines int) {
	if s.currentChunk.Len() == 0 {
		return
	}
	chunkType := ChunkTypeRaw
	if s.inFunction {
		chunkType = ChunkTypeFunction
	}
	s.chunks = append(s.chunks, Chunk{
		Content:   s.currentChunk.String(),
		StartLine: s.currentStart,
		EndLine:   totalLines - 1,
		Type:      chunkType,
		Name:      s.currentName,
	})
}

type functionPattern struct {
	pattern   *regexp.Regexp
	chunkType ChunkType
}

func (c *Chunker) getFunctionPatterns() []functionPattern {
	lang := strings.ToLower(c.config.Language)

	switch lang {
	case "go", "golang":
		return []functionPattern{
			{regexp.MustCompile(`^\s*func\s+(?:\([^)]+\)\s+)?(\w+)\s*\(`), ChunkTypeFunction},
			{regexp.MustCompile(`^\s*type\s+(\w+)\s+struct\s*\{`), ChunkTypeClass},
			{regexp.MustCompile(`^\s*type\s+(\w+)\s+interface\s*\{`), ChunkTypeClass},
		}
	case "javascript", "typescript", "js", "ts":
		return []functionPattern{
			{regexp.MustCompile(`^\s*(?:export\s+)?(?:async\s+)?function\s+(\w+)\s*\(`), ChunkTypeFunction},
			{regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?\(`), ChunkTypeFunction},
			{regexp.MustCompile(`^\s*(?:export\s+)?(?:const|let|var)\s+(\w+)\s*=\s*(?:async\s+)?function`), ChunkTypeFunction},
			{regexp.MustCompile(`^\s*(?:export\s+)?class\s+(\w+)`), ChunkTypeClass},
			{regexp.MustCompile(`^\s*(\w+)\s*\([^)]*\)\s*(?::\s*\w+)?\s*\{`), ChunkTypeMethod},
		}
	case "python", "py":
		return []functionPattern{
			{regexp.MustCompile(`^\s*(?:async\s+)?def\s+(\w+)\s*\(`), ChunkTypeFunction},
			{regexp.MustCompile(`^\s*class\s+(\w+)`), ChunkTypeClass},
		}
	case "java", "kotlin":
		return []functionPattern{
			{regexp.MustCompile(`^\s*(?:public|private|protected)?\s*(?:static)?\s*\w+\s+(\w+)\s*\(`), ChunkTypeMethod},
			{regexp.MustCompile(`^\s*(?:public|private|protected)?\s*class\s+(\w+)`), ChunkTypeClass},
		}
	case "rust", "rs":
		return []functionPattern{
			{regexp.MustCompile(`^\s*(?:pub\s+)?(?:async\s+)?fn\s+(\w+)`), ChunkTypeFunction},
			{regexp.MustCompile(`^\s*(?:pub\s+)?struct\s+(\w+)`), ChunkTypeClass},
			{regexp.MustCompile(`^\s*(?:pub\s+)?impl\s+(?:<[^>]+>\s+)?(\w+)`), ChunkTypeClass},
		}
	case "c", "cpp", "c++":
		return []functionPattern{
			{regexp.MustCompile(`^\s*(?:\w+\s+)+(\w+)\s*\([^)]*\)\s*\{`), ChunkTypeFunction},
			{regexp.MustCompile(`^\s*class\s+(\w+)`), ChunkTypeClass},
			{regexp.MustCompile(`^\s*struct\s+(\w+)`), ChunkTypeClass},
		}
	default:
		// Generic patterns
		return []functionPattern{
			{regexp.MustCompile(`^\s*(?:func|function|def|fn)\s+(\w+)`), ChunkTypeFunction},
			{regexp.MustCompile(`^\s*class\s+(\w+)`), ChunkTypeClass},
		}
	}
}

// splitBySize splits content by token size when function splitting isn't effective
func (c *Chunker) splitBySize(lines []string) []Chunk {
	var chunks []Chunk
	var currentChunk strings.Builder
	currentStart := 0
	currentTokens := 0

	for i, line := range lines {
		lineTokens := c.estimator.EstimateTokens(line)

		// Check if adding this line would exceed the limit
		if currentTokens+lineTokens > c.config.MaxChunkTokens && currentChunk.Len() > 0 {
			// Try to find a good break point
			breakPoint := c.findBreakPoint(currentChunk.String())
			if breakPoint > 0 {
				content := currentChunk.String()
				chunks = append(chunks, Chunk{
					Content:   content[:breakPoint],
					StartLine: currentStart,
					EndLine:   i - 1,
					Type:      ChunkTypeBlock,
				})
				// Carry over the remaining content
				currentChunk.Reset()
				currentChunk.WriteString(content[breakPoint:])
				currentTokens = c.estimator.EstimateTokens(currentChunk.String())
			} else {
				chunks = append(chunks, Chunk{
					Content:   currentChunk.String(),
					StartLine: currentStart,
					EndLine:   i - 1,
					Type:      ChunkTypeBlock,
				})
				currentChunk.Reset()
				currentTokens = 0
			}
			currentStart = i
		}

		currentChunk.WriteString(line)
		currentChunk.WriteString("\n")
		currentTokens += lineTokens
	}

	// Don't forget remaining content
	if currentChunk.Len() > 0 {
		chunks = append(chunks, Chunk{
			Content:   currentChunk.String(),
			StartLine: currentStart,
			EndLine:   len(lines) - 1,
			Type:      ChunkTypeBlock,
		})
	}

	return chunks
}

// findBreakPoint finds a good point to break content (end of function, empty line, etc.)
func (c *Chunker) findBreakPoint(content string) int {
	// Look for function end (closing brace followed by newline)
	if idx := strings.LastIndex(content, "}\n"); idx > 0 {
		return idx + 2
	}

	// Look for double newline (paragraph break)
	if idx := strings.LastIndex(content, "\n\n"); idx > 0 {
		return idx + 2
	}

	// Look for single newline at reasonable point
	if idx := strings.LastIndex(content, "\n"); idx > len(content)/2 {
		return idx + 1
	}

	return 0
}

// PrioritizeFiles sorts files by review priority
func PrioritizeFiles(files []FileInfo) []FileInfo {
	// Priority order:
	// 1. Source code files
	// 2. Test files
	// 3. Configuration files
	// 4. Documentation

	priorities := make(map[int][]FileInfo)

	for _, f := range files {
		priority := getFilePriority(f)
		priorities[priority] = append(priorities[priority], f)
	}

	var result []FileInfo
	for p := 1; p <= 4; p++ {
		result = append(result, priorities[p]...)
	}

	return result
}

// FileInfo contains information about a file for prioritization
type FileInfo struct {
	Path       string
	Language   string
	TokenCount int
	IsTest     bool
}

func getFilePriority(f FileInfo) int {
	// Test files are lower priority
	if f.IsTest || strings.Contains(f.Path, "_test") || strings.Contains(f.Path, ".test.") || strings.Contains(f.Path, ".spec.") {
		return 2
	}

	// Documentation is lowest priority
	ext := strings.ToLower(getExtension(f.Path))
	if ext == "md" || ext == "txt" || ext == "rst" || ext == "adoc" {
		return 4
	}

	// Config files are lower priority
	if ext == "json" || ext == "yaml" || ext == "yml" || ext == "toml" || ext == "ini" {
		return 3
	}

	// Source code is highest priority
	return 1
}

func getExtension(path string) string {
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		return path[idx+1:]
	}
	return ""
}
