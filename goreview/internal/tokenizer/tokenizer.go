// Package tokenizer provides token estimation and budget management for LLM requests.
package tokenizer

import (
	"strings"
	"unicode/utf8"
)

// Model token limits (approximate)
const (
	// Context window sizes
	GPT4TurboMaxTokens   = 128000
	GPT4MaxTokens        = 8192
	GPT35TurboMaxTokens  = 16384
	ClaudeMaxTokens      = 200000
	GeminiProMaxTokens   = 32768
	GeminiFlashMaxTokens = 1048576
	LlamaMaxTokens       = 8192
	MistralMaxTokens     = 32768
	QwenMaxTokens        = 32768

	// Default budget (conservative)
	DefaultMaxTokens = 8000

	// Reserve tokens for response
	DefaultResponseReserve = 2000
)

// Estimator estimates token counts for text
type Estimator struct {
	// Average characters per token (varies by model, ~4 for English)
	charsPerToken float64

	// Model-specific adjustments
	modelMultiplier float64
}

// NewEstimator creates a new token estimator
func NewEstimator() *Estimator {
	return &Estimator{
		charsPerToken:   4.0,
		modelMultiplier: 1.0,
	}
}

// NewEstimatorForModel creates an estimator tuned for a specific model
func NewEstimatorForModel(model string) *Estimator {
	e := NewEstimator()

	switch {
	case strings.Contains(model, "gpt-4"):
		e.charsPerToken = 4.0
	case strings.Contains(model, "gpt-3.5"):
		e.charsPerToken = 4.0
	case strings.Contains(model, "claude"):
		e.charsPerToken = 3.5
	case strings.Contains(model, "gemini"):
		e.charsPerToken = 4.0
	case strings.Contains(model, "llama"):
		e.charsPerToken = 4.5
	case strings.Contains(model, "mistral"):
		e.charsPerToken = 4.0
	case strings.Contains(model, "qwen"):
		e.charsPerToken = 3.0 // CJK characters use more tokens
	default:
		e.charsPerToken = 4.0
	}

	return e
}

// EstimateTokens estimates the token count for a given text
func (e *Estimator) EstimateTokens(text string) int {
	if text == "" {
		return 0
	}

	// Count characters (rune count for proper Unicode handling)
	charCount := utf8.RuneCountInString(text)

	// Base estimate
	estimate := float64(charCount) / e.charsPerToken

	// Adjust for code (code tends to have more tokens due to punctuation)
	if isCodeLike(text) {
		estimate *= 1.3
	}

	// Adjust for whitespace-heavy content
	whitespaceRatio := countWhitespace(text) / float64(charCount)
	if whitespaceRatio > 0.3 {
		estimate *= 0.9
	}

	return int(estimate * e.modelMultiplier)
}

// EstimateTokensForDiff estimates tokens for a diff with context
func (e *Estimator) EstimateTokensForDiff(diff, language, filePath string) int {
	// Diff content
	diffTokens := e.EstimateTokens(diff)

	// Metadata overhead (language, path, instructions)
	metadataTokens := e.EstimateTokens(language) + e.EstimateTokens(filePath) + 50

	// System prompt overhead (approximate)
	systemPromptTokens := 200

	return diffTokens + metadataTokens + systemPromptTokens
}

// isCodeLike checks if text appears to be code
func isCodeLike(text string) bool {
	codeIndicators := []string{
		"func ", "function ", "def ", "class ",
		"if ", "for ", "while ", "switch ",
		"import ", "require(", "from ",
		"return ", "const ", "let ", "var ",
		"{", "}", "(", ")", "[", "]",
		"=>", "->", "::", "//", "/*",
	}

	lowerText := strings.ToLower(text)
	indicatorCount := 0
	for _, indicator := range codeIndicators {
		if strings.Contains(lowerText, indicator) {
			indicatorCount++
		}
	}

	return indicatorCount >= 3
}

// countWhitespace counts whitespace characters
func countWhitespace(text string) float64 {
	count := 0
	for _, r := range text {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			count++
		}
	}
	return float64(count)
}

// Budget manages token budgets for requests
type Budget struct {
	maxTokens       int
	responseReserve int
	used            int
}

// NewBudget creates a new token budget
func NewBudget(maxTokens, responseReserve int) *Budget {
	if maxTokens <= 0 {
		maxTokens = DefaultMaxTokens
	}
	if responseReserve <= 0 {
		responseReserve = DefaultResponseReserve
	}
	return &Budget{
		maxTokens:       maxTokens,
		responseReserve: responseReserve,
		used:            0,
	}
}

// Available returns the available tokens for input
func (b *Budget) Available() int {
	return b.maxTokens - b.responseReserve - b.used
}

// Use marks tokens as used
func (b *Budget) Use(tokens int) {
	b.used += tokens
}

// CanFit checks if a number of tokens can fit in the budget
func (b *Budget) CanFit(tokens int) bool {
	return tokens <= b.Available()
}

// Reset resets the budget
func (b *Budget) Reset() {
	b.used = 0
}

// GetModelMaxTokens returns the max tokens for a model
func GetModelMaxTokens(model string) int {
	switch {
	case strings.Contains(model, "gpt-4-turbo"), strings.Contains(model, "gpt-4o"):
		return GPT4TurboMaxTokens
	case strings.Contains(model, "gpt-4"):
		return GPT4MaxTokens
	case strings.Contains(model, "gpt-3.5"):
		return GPT35TurboMaxTokens
	case strings.Contains(model, "claude-3"):
		return ClaudeMaxTokens
	case strings.Contains(model, "gemini-1.5-flash"):
		return GeminiFlashMaxTokens
	case strings.Contains(model, "gemini"):
		return GeminiProMaxTokens
	case strings.Contains(model, "llama"):
		return LlamaMaxTokens
	case strings.Contains(model, "mistral"):
		return MistralMaxTokens
	case strings.Contains(model, "qwen"):
		return QwenMaxTokens
	default:
		return DefaultMaxTokens
	}
}
