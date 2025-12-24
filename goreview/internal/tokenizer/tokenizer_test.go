package tokenizer

import (
	"strings"
	"testing"
)

func TestEstimateTokens(t *testing.T) {
	e := NewEstimator()

	tests := []struct {
		name     string
		input    string
		minToken int
		maxToken int
	}{
		{
			name:     "empty string",
			input:    "",
			minToken: 0,
			maxToken: 0,
		},
		{
			name:     "simple text",
			input:    "Hello, world!",
			minToken: 2,
			maxToken: 10,
		},
		{
			name:     "code snippet",
			input:    "func main() { fmt.Println(\"Hello\") }",
			minToken: 8,
			maxToken: 20,
		},
		{
			name:     "longer code",
			input:    strings.Repeat("func test() { return nil }\n", 10),
			minToken: 50,
			maxToken: 150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := e.EstimateTokens(tt.input)
			if tokens < tt.minToken || tokens > tt.maxToken {
				t.Errorf("EstimateTokens() = %d, want between %d and %d", tokens, tt.minToken, tt.maxToken)
			}
		})
	}
}

func TestEstimatorForModel(t *testing.T) {
	models := []string{
		"gpt-4",
		"gpt-3.5-turbo",
		"claude-3-opus",
		"gemini-pro",
		"llama-2-70b",
		"mistral-7b",
		"qwen-72b",
		"unknown-model",
	}

	text := "This is a test string with some code: func main() { return 42 }"

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			e := NewEstimatorForModel(model)
			tokens := e.EstimateTokens(text)
			if tokens <= 0 {
				t.Errorf("EstimateTokens for %s returned %d, want > 0", model, tokens)
			}
		})
	}
}

func TestBudget(t *testing.T) {
	b := NewBudget(1000, 200)

	if b.Available() != 800 {
		t.Errorf("Available() = %d, want 800", b.Available())
	}

	b.Use(300)
	if b.Available() != 500 {
		t.Errorf("Available() after Use = %d, want 500", b.Available())
	}

	if !b.CanFit(400) {
		t.Error("CanFit(400) should return true")
	}

	if b.CanFit(600) {
		t.Error("CanFit(600) should return false")
	}

	b.Reset()
	if b.Available() != 800 {
		t.Errorf("Available() after Reset = %d, want 800", b.Available())
	}
}

func TestGetModelMaxTokens(t *testing.T) {
	tests := []struct {
		model    string
		expected int
	}{
		{"gpt-4-turbo", GPT4TurboMaxTokens},
		{"gpt-4o-mini", GPT4TurboMaxTokens},
		{"gpt-4", GPT4MaxTokens},
		{"gpt-3.5-turbo", GPT35TurboMaxTokens},
		{"claude-3-opus", ClaudeMaxTokens},
		{"gemini-1.5-pro", GeminiProMaxTokens},
		{"gemini-1.5-flash", GeminiFlashMaxTokens},
		{"llama-3-70b", LlamaMaxTokens},
		{"mistral-large", MistralMaxTokens},
		{"qwen-72b-chat", QwenMaxTokens},
		{"unknown", DefaultMaxTokens},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			if got := GetModelMaxTokens(tt.model); got != tt.expected {
				t.Errorf("GetModelMaxTokens(%s) = %d, want %d", tt.model, got, tt.expected)
			}
		})
	}
}

func TestChunker(t *testing.T) {
	cfg := ChunkerConfig{
		MaxChunkTokens: 100,
		Language:       "go",
	}
	c := NewChunker(cfg)

	goCode := `package main

func Hello() string {
	return "Hello"
}

func World() string {
	return "World"
}

func main() {
	fmt.Println(Hello(), World())
}`

	chunks := c.ChunkDiff(goCode)

	if len(chunks) == 0 {
		t.Error("Expected at least one chunk")
	}

	// Verify all content is preserved
	var combined strings.Builder
	for _, chunk := range chunks {
		combined.WriteString(chunk.Content)
	}

	// Content should be preserved (allow for minor whitespace differences)
	if len(combined.String()) < int(float64(len(goCode))*0.9) {
		t.Error("Chunking lost too much content")
	}
}

func TestChunkerLanguages(t *testing.T) {
	languages := []struct {
		lang string
		code string
	}{
		{
			lang: "javascript",
			code: `function hello() {
	return "hello";
}

const world = () => {
	return "world";
};`,
		},
		{
			lang: "python",
			code: `def hello():
    return "hello"

class Greeter:
    def greet(self):
        return "world"`,
		},
		{
			lang: "rust",
			code: `fn hello() -> &'static str {
    "hello"
}

pub struct Greeter;

impl Greeter {
    pub fn greet(&self) -> &'static str {
        "world"
    }
}`,
		},
	}

	for _, tt := range languages {
		t.Run(tt.lang, func(t *testing.T) {
			c := NewChunker(ChunkerConfig{
				MaxChunkTokens: 50,
				Language:       tt.lang,
			})
			chunks := c.ChunkDiff(tt.code)
			if len(chunks) == 0 {
				t.Errorf("Expected chunks for %s", tt.lang)
			}
		})
	}
}

func TestPrioritizeFiles(t *testing.T) {
	files := []FileInfo{
		{Path: "README.md", Language: "markdown"},
		{Path: "main.go", Language: "go"},
		{Path: "config.yaml", Language: "yaml"},
		{Path: "main_test.go", Language: "go", IsTest: true},
		{Path: "utils.go", Language: "go"},
	}

	prioritized := PrioritizeFiles(files)

	// Source files should come first
	if prioritized[0].Path != "main.go" && prioritized[0].Path != "utils.go" {
		t.Errorf("Expected source file first, got %s", prioritized[0].Path)
	}

	// Documentation should come last
	if prioritized[len(prioritized)-1].Path != "README.md" {
		t.Errorf("Expected documentation last, got %s", prioritized[len(prioritized)-1].Path)
	}
}

func BenchmarkEstimateTokens(b *testing.B) {
	e := NewEstimator()
	code := strings.Repeat("func test() { return nil }\n", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.EstimateTokens(code)
	}
}

func BenchmarkChunkDiff(b *testing.B) {
	c := NewChunker(ChunkerConfig{
		MaxChunkTokens: 500,
		Language:       "go",
	})
	code := strings.Repeat(`func test() {
	x := 1
	y := 2
	return x + y
}

`, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.ChunkDiff(code)
	}
}
