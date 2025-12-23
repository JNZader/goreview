package memory

import (
	"hash/fnv"
	"math"
	"regexp"
	"strings"
	"unicode"
)

// EmbeddingDim is the dimension of the embedding vectors.
const EmbeddingDim = 256

// Embedder generates vector embeddings for text content.
type Embedder struct {
	// Stopwords to filter out
	stopwords map[string]bool

	// Token pattern
	tokenPattern *regexp.Regexp
}

// NewEmbedder creates a new embedder instance.
func NewEmbedder() *Embedder {
	return &Embedder{
		stopwords:    defaultStopwords(),
		tokenPattern: regexp.MustCompile(`[a-zA-Z_][a-zA-Z0-9_]*|[0-9]+`),
	}
}

// Embed generates an embedding vector for the given text.
// Uses a combination of:
// - Bag of words with feature hashing
// - Character n-grams for subword information
// - Positional weighting for important tokens
func (e *Embedder) Embed(text string) []float32 {
	if text == "" {
		return make([]float32, EmbeddingDim)
	}

	embedding := make([]float32, EmbeddingDim)

	// Extract tokens
	tokens := e.tokenize(text)
	if len(tokens) == 0 {
		return embedding
	}

	// Calculate token weights (TF-style)
	tokenCounts := make(map[string]int)
	for _, token := range tokens {
		tokenCounts[token]++
	}

	// Add word features
	for token, count := range tokenCounts {
		if e.isStopword(token) {
			continue
		}

		// TF weight (log normalization)
		weight := float32(1 + math.Log(float64(count)))

		// Hash token to embedding dimension
		h := e.hash(token)
		idx := h % uint64(EmbeddingDim)

		// Use positive/negative based on second hash (for sparse coding)
		sign := float32(1.0)
		if e.hash(token+"_sign")%2 == 1 {
			sign = -1.0
		}

		embedding[idx] += sign * weight
	}

	// Add character n-gram features for subword information
	for token := range tokenCounts {
		if len(token) >= 3 {
			for i := 0; i <= len(token)-3; i++ {
				ngram := token[i : i+3]
				h := e.hash("ngram:" + ngram)
				idx := h % uint64(EmbeddingDim)
				embedding[idx] += 0.3 // Smaller weight for n-grams
			}
		}
	}

	// Add code-specific features
	e.addCodeFeatures(text, embedding)

	// L2 normalize
	e.normalize(embedding)

	return embedding
}

// EmbedBatch generates embeddings for multiple texts.
func (e *Embedder) EmbedBatch(texts []string) [][]float32 {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embeddings[i] = e.Embed(text)
	}
	return embeddings
}

// Similarity calculates cosine similarity between two embeddings.
func (e *Embedder) Similarity(a, b []float32) float64 {
	return cosineSimilarity(a, b)
}

// Internal methods

func (e *Embedder) tokenize(text string) []string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Find all tokens
	matches := e.tokenPattern.FindAllString(text, -1)

	// Filter and normalize
	tokens := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) >= 2 { // Skip single characters
			tokens = append(tokens, match)
		}
	}

	return tokens
}

func (e *Embedder) isStopword(token string) bool {
	return e.stopwords[token]
}

func (e *Embedder) hash(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

func (e *Embedder) normalize(embedding []float32) {
	var norm float64
	for _, v := range embedding {
		norm += float64(v) * float64(v)
	}

	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range embedding {
			embedding[i] = float32(float64(embedding[i]) / norm)
		}
	}
}

// codePatterns contains precompiled regex patterns for code feature detection.
var codePatterns = map[string]*regexp.Regexp{
	"func":     regexp.MustCompile(`(?i)func\s+\w+`),
	"class":    regexp.MustCompile(`(?i)class\s+\w+`),
	"import":   regexp.MustCompile(`(?i)import\s+`),
	"return":   regexp.MustCompile(`(?i)return\s+`),
	"error":    regexp.MustCompile(`(?i)error|Error|err\b`),
	"test":     regexp.MustCompile(`(?i)test|Test|_test\.`),
	"security": regexp.MustCompile(`(?i)security|unsafe|injection|xss|sql`),
	"bug":      regexp.MustCompile(`(?i)bug|fix|issue|problem`),
	"perf":     regexp.MustCompile(`(?i)performance|slow|fast|optimize`),
}

func (e *Embedder) addCodeFeatures(text string, embedding []float32) {
	// Detect code patterns and add features
	for name, pattern := range codePatterns {
		if pattern.MatchString(text) {
			h := e.hash("code_feature:" + name)
			idx := h % uint64(EmbeddingDim)
			embedding[idx] += 0.5
		}
	}

	// Count special characters (code indicators)
	specialCounts := map[rune]string{
		'{': "braces",
		'(': "parens",
		'[': "brackets",
		';': "semicolons",
		':': "colons",
		'.': "dots",
	}

	for _, r := range text {
		if name, ok := specialCounts[r]; ok {
			h := e.hash("special:" + name)
			idx := h % uint64(EmbeddingDim)
			embedding[idx] += 0.1
		}
	}

	// Detect common programming language keywords
	languages := map[string][]string{
		"go":         {"func", "package", "import", "defer", "goroutine", "chan"},
		"python":     {"def", "class", "import", "from", "self", "lambda"},
		"javascript": {"const", "let", "var", "function", "async", "await"},
		"rust":       {"fn", "impl", "struct", "enum", "match", "mut"},
		"java":       {"public", "private", "class", "interface", "extends"},
	}

	lowerText := strings.ToLower(text)
	for lang, keywords := range languages {
		count := 0
		for _, kw := range keywords {
			if strings.Contains(lowerText, kw) {
				count++
			}
		}
		if count >= 2 {
			h := e.hash("lang:" + lang)
			idx := h % uint64(EmbeddingDim)
			embedding[idx] += float32(count) * 0.2
		}
	}
}

// CamelCaseToWords splits camelCase into words.
func CamelCaseToWords(s string) []string {
	var words []string
	var current strings.Builder

	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
		}
		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, strings.ToLower(current.String()))
	}

	return words
}

// SnakeCaseToWords splits snake_case into words.
func SnakeCaseToWords(s string) []string {
	parts := strings.Split(s, "_")
	words := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			words = append(words, strings.ToLower(p))
		}
	}
	return words
}

// defaultStopwords returns common English stopwords.
func defaultStopwords() map[string]bool {
	words := []string{
		"a", "an", "and", "are", "as", "at", "be", "by", "for",
		"from", "has", "he", "in", "is", "it", "its", "of", "on",
		"or", "that", "the", "to", "was", "were", "will", "with",
		"this", "but", "they", "have", "had", "what", "when", "where",
		"who", "which", "why", "how", "all", "each", "every", "both",
		"few", "more", "most", "other", "some", "such", "no", "nor",
		"not", "only", "own", "same", "so", "than", "too", "very",
		"can", "just", "should", "now", "if", "then", "else",
	}

	m := make(map[string]bool)
	for _, w := range words {
		m[w] = true
	}
	return m
}

// SemanticIndex provides fast semantic search using embeddings.
type SemanticIndex struct {
	embedder *Embedder
	entries  map[string]*indexEntry
}

type indexEntry struct {
	ID        string
	Embedding []float32
}

// NewSemanticIndex creates a new semantic index.
func NewSemanticIndex() *SemanticIndex {
	return &SemanticIndex{
		embedder: NewEmbedder(),
		entries:  make(map[string]*indexEntry),
	}
}

// Index adds an entry to the semantic index.
func (s *SemanticIndex) Index(id, content string) {
	embedding := s.embedder.Embed(content)
	s.entries[id] = &indexEntry{
		ID:        id,
		Embedding: embedding,
	}
}

// Remove removes an entry from the index.
func (s *SemanticIndex) Remove(id string) {
	delete(s.entries, id)
}

// Search finds the most similar entries to the query.
func (s *SemanticIndex) Search(query string, limit int) []SemanticResult {
	queryEmbedding := s.embedder.Embed(query)

	results := make([]SemanticResult, 0, len(s.entries))

	for _, entry := range s.entries {
		similarity := s.embedder.Similarity(queryEmbedding, entry.Embedding)
		if similarity > 0 {
			results = append(results, SemanticResult{
				ID:         entry.ID,
				Similarity: similarity,
			})
		}
	}

	// Sort by similarity (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Similarity > results[i].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results
}

// SearchByEmbedding finds the most similar entries to the given embedding.
func (s *SemanticIndex) SearchByEmbedding(embedding []float32, limit int) []SemanticResult {
	results := make([]SemanticResult, 0, len(s.entries))

	for _, entry := range s.entries {
		similarity := s.embedder.Similarity(embedding, entry.Embedding)
		if similarity > 0 {
			results = append(results, SemanticResult{
				ID:         entry.ID,
				Similarity: similarity,
			})
		}
	}

	// Sort by similarity (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Similarity > results[i].Similarity {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results
}

// SemanticResult represents a semantic search result.
type SemanticResult struct {
	ID         string
	Similarity float64
}

// GetEmbedding returns the embedding for an entry.
func (s *SemanticIndex) GetEmbedding(id string) ([]float32, bool) {
	entry, ok := s.entries[id]
	if !ok {
		return nil, false
	}
	return entry.Embedding, true
}

// Size returns the number of indexed entries.
func (s *SemanticIndex) Size() int {
	return len(s.entries)
}
