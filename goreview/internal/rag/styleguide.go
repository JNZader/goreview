// Package rag provides Retrieval Augmented Generation for code review.
// It indexes style guides and coding standards for context-aware reviews.
package rag

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

// StyleGuide represents an indexed style guide document
type StyleGuide struct {
	Path     string        `json:"path"`
	Sections []RuleSection `json:"sections"`
	Hash     string        `json:"hash"`
}

// RuleSection represents a section of a style guide
type RuleSection struct {
	Title    string   `json:"title"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
	Priority int      `json:"priority"` // Higher = more important
}

// Index stores indexed style guides for retrieval
type Index struct {
	mu        sync.RWMutex
	guides    map[string]*StyleGuide
	sections  []indexedSection
	tagIndex  map[string][]int // tag -> section indices
	wordIndex map[string][]int // word -> section indices
}

type indexedSection struct {
	guideHash string
	section   RuleSection
}

// NewIndex creates a new RAG index
func NewIndex() *Index {
	return &Index{
		guides:    make(map[string]*StyleGuide),
		sections:  []indexedSection{},
		tagIndex:  make(map[string][]int),
		wordIndex: make(map[string][]int),
	}
}

// DefaultStyleGuidePatterns are common style guide file patterns
var DefaultStyleGuidePatterns = []string{
	"STYLEGUIDE.md",
	"STYLE_GUIDE.md",
	"CODING_STANDARDS.md",
	"CONTRIBUTING.md",
	".coding-standards.md",
	".styleguide.md",
	"docs/style-guide.md",
	"docs/coding-standards.md",
	".github/STYLE_GUIDE.md",
}

// skipDirs contains directories to skip during style guide search
var skipDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".git":         true,
}

// styleKeywords contains keywords indicating a style guide file
var styleKeywords = []string{"style", "standard", "convention", "guideline"}

// LoadFromDirectory loads style guides from a directory
func (idx *Index) LoadFromDirectory(dir string) error {
	if err := idx.loadFromPatterns(dir); err != nil {
		return err
	}

	_ = filepath.Walk(dir, idx.walkStyleGuides)
	return nil
}

// loadFromPatterns loads style guides from default patterns
func (idx *Index) loadFromPatterns(dir string) error {
	for _, pattern := range DefaultStyleGuidePatterns {
		path := filepath.Join(dir, pattern)
		if _, err := os.Stat(path); err == nil {
			if err := idx.LoadFile(path); err != nil {
				return err
			}
		}
	}
	return nil
}

// walkStyleGuides is the walk function for finding style guide files
func (idx *Index) walkStyleGuides(path string, info os.FileInfo, err error) error {
	if err != nil {
		return nil
	}

	if info.IsDir() {
		if skipDirs[info.Name()] {
			return filepath.SkipDir
		}
		return nil
	}

	if !strings.HasSuffix(path, ".md") {
		return nil
	}

	if idx.isStyleGuideFile(info.Name()) {
		_ = idx.LoadFile(path)
	}
	return nil
}

// isStyleGuideFile checks if a filename indicates a style guide
func (idx *Index) isStyleGuideFile(name string) bool {
	lowerName := strings.ToLower(name)
	for _, kw := range styleKeywords {
		if strings.Contains(lowerName, kw) {
			return true
		}
	}
	return false
}

// LoadFile loads and indexes a style guide file
func (idx *Index) LoadFile(path string) error {
	cleanPath := filepath.Clean(path)
	content, err := os.ReadFile(cleanPath) // #nosec G304 - path from trusted directory scan
	if err != nil {
		return err
	}

	return idx.LoadContent(path, string(content))
}

// LoadContent loads and indexes style guide content
func (idx *Index) LoadContent(path, content string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Calculate hash
	hash := sha256.Sum256([]byte(content))
	hashStr := hex.EncodeToString(hash[:8])

	// Check if already indexed with same hash
	if existing, ok := idx.guides[path]; ok && existing.Hash == hashStr {
		return nil
	}

	// Parse the markdown into sections
	sections := parseMarkdownSections(content)

	guide := &StyleGuide{
		Path:     path,
		Sections: sections,
		Hash:     hashStr,
	}

	idx.guides[path] = guide

	// Index sections
	for _, section := range sections {
		idx.indexSection(hashStr, section)
	}

	return nil
}

// parseMarkdownSections parses markdown into logical sections
func parseMarkdownSections(content string) []RuleSection {
	// Split by headers
	headerPattern := regexp.MustCompile(`(?m)^(#{1,6})\s+(.+)$`)
	codeBlockPattern := regexp.MustCompile("```[\\s\\S]*?```")

	// Remove code blocks temporarily (to not split on # in code)
	placeholders := make(map[string]string)
	placeholderIndex := 0
	contentClean := codeBlockPattern.ReplaceAllStringFunc(content, func(match string) string {
		placeholder := "<<CODEBLOCK_" + string(rune('A'+placeholderIndex)) + ">>"
		placeholders[placeholder] = match
		placeholderIndex++
		return placeholder
	})

	matches := headerPattern.FindAllStringSubmatchIndex(contentClean, -1)
	sections := make([]RuleSection, 0, len(matches))

	for i, match := range matches {
		if len(match) < 6 {
			continue
		}

		headerLevel := len(contentClean[match[2]:match[3]])
		title := contentClean[match[4]:match[5]]

		// Get content until next header
		contentStart := match[1]
		var contentEnd int
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		} else {
			contentEnd = len(contentClean)
		}

		sectionContent := strings.TrimSpace(contentClean[contentStart:contentEnd])

		// Restore code blocks
		for placeholder, original := range placeholders {
			sectionContent = strings.ReplaceAll(sectionContent, placeholder, original)
		}

		// Extract tags from content
		tags := extractTags(title, sectionContent)

		// Calculate priority based on header level and keywords
		priority := calculatePriority(headerLevel, title, sectionContent)

		sections = append(sections, RuleSection{
			Title:    title,
			Content:  sectionContent,
			Tags:     tags,
			Priority: priority,
		})
	}

	return sections
}

// extractTags extracts relevant tags from content
func extractTags(title, content string) []string {
	tagSet := make(map[string]bool)

	combined := strings.ToLower(title + " " + content)

	// Programming concepts
	concepts := []string{
		"naming", "convention", "variable", "function", "class", "method",
		"error", "exception", "handling", "logging", "testing", "test",
		"documentation", "comment", "import", "module", "package",
		"security", "performance", "memory", "concurrency", "async",
		"type", "interface", "struct", "enum", "constant",
		"formatting", "indentation", "spacing", "lint", "style",
		"git", "commit", "branch", "merge", "review",
		"dependency", "version", "deprecated", "breaking",
	}

	for _, concept := range concepts {
		if strings.Contains(combined, concept) {
			tagSet[concept] = true
		}
	}

	// Language-specific tags
	languages := []string{
		"go", "golang", "javascript", "typescript", "python",
		"java", "rust", "c++", "cpp", "c#", "csharp",
		"ruby", "php", "swift", "kotlin",
	}

	for _, lang := range languages {
		if strings.Contains(combined, lang) {
			tagSet[lang] = true
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags
}

// calculatePriority calculates section priority
func calculatePriority(headerLevel int, title, content string) int {
	priority := 10 - headerLevel // Higher level headers = higher priority

	lowerTitle := strings.ToLower(title)
	lowerContent := strings.ToLower(content)

	// Boost for important keywords
	importantKeywords := []string{
		"must", "required", "always", "never", "forbidden",
		"security", "critical", "important", "mandatory",
	}

	for _, kw := range importantKeywords {
		if strings.Contains(lowerTitle, kw) {
			priority += 5
		}
		if strings.Contains(lowerContent, kw) {
			priority += 2
		}
	}

	// Boost for actionable content (contains code examples)
	if strings.Contains(content, "```") {
		priority += 3
	}

	return priority
}

// indexSection adds a section to the indices
func (idx *Index) indexSection(guideHash string, section RuleSection) {
	sectionIdx := len(idx.sections)
	idx.sections = append(idx.sections, indexedSection{
		guideHash: guideHash,
		section:   section,
	})

	// Index by tags
	for _, tag := range section.Tags {
		idx.tagIndex[tag] = append(idx.tagIndex[tag], sectionIdx)
	}

	// Index by words in title
	words := tokenize(section.Title)
	for _, word := range words {
		if len(word) >= 3 { // Skip short words
			idx.wordIndex[word] = append(idx.wordIndex[word], sectionIdx)
		}
	}
}

// tokenize splits text into lowercase words
func tokenize(text string) []string {
	wordPattern := regexp.MustCompile(`\w+`)
	matches := wordPattern.FindAllString(strings.ToLower(text), -1)
	return matches
}

// RetrievalQuery represents a query for retrieving relevant rules
type RetrievalQuery struct {
	Language     string   // Programming language
	FilePath     string   // File being reviewed
	CodeContext  string   // Code snippet or diff
	FunctionName string   // Current function if known
	Tags         []string // Explicit tags to search for
}

// RetrievalResult represents a retrieved rule section
type RetrievalResult struct {
	Section RuleSection `json:"section"`
	Score   float64     `json:"score"`
	Source  string      `json:"source"` // File path
}

// scoredSection holds a section index with its relevance score
type scoredSection struct {
	index int
	score float64
}

// Retrieve retrieves relevant rule sections for a query
func (idx *Index) Retrieve(query RetrievalQuery, limit int) []RetrievalResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if limit <= 0 {
		limit = 5
	}

	scores := make(map[int]float64)
	idx.scoreByLanguage(scores, query.Language)
	idx.scoreByTags(scores, query.Tags)
	idx.scoreByCodeContext(scores, query.CodeContext)
	idx.scoreByWords(scores, query.FilePath, query.FunctionName)
	idx.addBasePriorityScores(scores)

	return idx.buildResults(scores, limit)
}

// scoreByLanguage adds score for language tag matches
func (idx *Index) scoreByLanguage(scores map[int]float64, language string) {
	if language == "" {
		return
	}
	langLower := strings.ToLower(language)
	for _, i := range idx.tagIndex[langLower] {
		scores[i] += 5.0
	}
}

// scoreByTags adds score for explicit tag matches
func (idx *Index) scoreByTags(scores map[int]float64, tags []string) {
	for _, tag := range tags {
		tagLower := strings.ToLower(tag)
		for _, i := range idx.tagIndex[tagLower] {
			scores[i] += 3.0
		}
	}
}

// scoreByCodeContext adds score based on inferred code context tags
func (idx *Index) scoreByCodeContext(scores map[int]float64, codeContext string) {
	if codeContext == "" {
		return
	}
	for _, tag := range inferTagsFromCode(codeContext) {
		for _, i := range idx.tagIndex[tag] {
			scores[i] += 2.0
		}
	}
}

// scoreByWords adds score for word matches in title
func (idx *Index) scoreByWords(scores map[int]float64, filePath, functionName string) {
	for _, word := range tokenize(filePath + " " + functionName) {
		if len(word) >= 3 {
			for _, i := range idx.wordIndex[word] {
				scores[i] += 1.0
			}
		}
	}
}

// addBasePriorityScores adds base score from section priority
func (idx *Index) addBasePriorityScores(scores map[int]float64) {
	for i := range idx.sections {
		scores[i] += float64(idx.sections[i].section.Priority) * 0.1
	}
}

// buildResults creates sorted results from scores
func (idx *Index) buildResults(scores map[int]float64, limit int) []RetrievalResult {
	scored := make([]scoredSection, 0, len(scores))
	for i, score := range scores {
		if score > 0 {
			scored = append(scored, scoredSection{i, score})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	results := make([]RetrievalResult, 0, min(len(scored), limit))
	for i := 0; i < len(scored) && i < limit; i++ {
		s := idx.sections[scored[i].index]
		results = append(results, RetrievalResult{
			Section: s.section,
			Score:   scored[i].score,
			Source:  idx.getSourcePath(s.guideHash),
		})
	}
	return results
}

func (idx *Index) getSourcePath(hash string) string {
	for path, guide := range idx.guides {
		if guide.Hash == hash {
			return path
		}
	}
	return ""
}

// inferTagsFromCode infers relevant tags from code content
func inferTagsFromCode(code string) []string {
	var tags []string
	lowerCode := strings.ToLower(code)

	tagPatterns := map[string][]string{
		"error":       {"error", "err", "exception", "panic", "throw"},
		"logging":     {"log", "logger", "print", "debug", "warn"},
		"testing":     {"test", "spec", "expect", "assert", "mock"},
		"async":       {"async", "await", "promise"},
		"naming":      {"func ", "function ", "def ", "class ", "var ", "const "},
		"security":    {"password", "secret", "token", "auth", "crypto"},
		"performance": {"cache", "optimize", "benchmark", "profile"},
		"concurrency": {"mutex", "lock", "sync", "atomic", "thread", "goroutine", "channel"},
	}

	for tag, patterns := range tagPatterns {
		for _, pattern := range patterns {
			if strings.Contains(lowerCode, pattern) {
				tags = append(tags, tag)
				break
			}
		}
	}

	return tags
}

// FormatForPrompt formats retrieved rules for inclusion in an LLM prompt
func FormatForPrompt(results []RetrievalResult, maxLength int) string {
	if len(results) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("## Relevant Style Guide Rules:\n\n")

	currentLength := sb.Len()

	for i, result := range results {
		section := result.Section

		// Format section
		formatted := "### " + section.Title + "\n"
		formatted += section.Content + "\n\n"

		// Check length
		if currentLength+len(formatted) > maxLength {
			if i == 0 {
				// At least include a truncated first result
				remaining := maxLength - currentLength - 50
				if remaining > 100 {
					formatted = formatted[:remaining] + "\n... (truncated)\n"
				} else {
					break
				}
			} else {
				sb.WriteString("... (more rules available)\n")
				break
			}
		}

		sb.WriteString(formatted)
		currentLength += len(formatted)
	}

	return sb.String()
}

// Stats returns index statistics
func (idx *Index) Stats() IndexStats {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return IndexStats{
		TotalGuides:   len(idx.guides),
		TotalSections: len(idx.sections),
		TotalTags:     len(idx.tagIndex),
		TotalWords:    len(idx.wordIndex),
	}
}

// IndexStats contains index statistics
type IndexStats struct {
	TotalGuides   int `json:"total_guides"`
	TotalSections int `json:"total_sections"`
	TotalTags     int `json:"total_tags"`
	TotalWords    int `json:"total_words"`
}
