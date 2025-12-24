package rag

import (
	"strings"
	"testing"
)

func TestParseMarkdownSections(t *testing.T) {
	content := `# Style Guide

## Naming Conventions

Always use descriptive names for variables.

### Function Names

Functions should be verbs or verb phrases.

## Error Handling

Never ignore errors. Always handle or propagate them.

` + "```go" + `
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
` + "```"

	sections := parseMarkdownSections(content)

	if len(sections) < 3 {
		t.Errorf("Expected at least 3 sections, got %d", len(sections))
	}

	// Check that sections were parsed correctly
	foundNaming := false
	foundError := false

	for _, s := range sections {
		if strings.Contains(s.Title, "Naming") {
			foundNaming = true
		}
		if strings.Contains(s.Title, "Error") {
			foundError = true
			// Should contain code block
			if !strings.Contains(s.Content, "```") {
				t.Error("Error handling section should contain code block")
			}
		}
	}

	if !foundNaming {
		t.Error("Should have found Naming section")
	}
	if !foundError {
		t.Error("Should have found Error handling section")
	}
}

func TestExtractTags(t *testing.T) {
	title := "Error Handling Guidelines"
	content := "Always use proper error handling in Go functions."

	tags := extractTags(title, content)

	expectedTags := map[string]bool{
		"error":    true,
		"handling": true,
		"go":       true,
		"function": true,
	}

	for expected := range expectedTags {
		found := false
		for _, tag := range tags {
			if tag == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tag '%s' not found in %v", expected, tags)
		}
	}
}

func TestIndexLoadContent(t *testing.T) {
	idx := NewIndex()

	content := `# Coding Standards

## Security

Never store passwords in plain text.

## Performance

Always use caching for expensive operations.
`

	err := idx.LoadContent("STYLEGUIDE.md", content)
	if err != nil {
		t.Fatalf("LoadContent failed: %v", err)
	}

	stats := idx.Stats()
	if stats.TotalGuides != 1 {
		t.Errorf("Expected 1 guide, got %d", stats.TotalGuides)
	}

	if stats.TotalSections == 0 {
		t.Error("Expected sections to be indexed")
	}

	if stats.TotalTags == 0 {
		t.Error("Expected tags to be indexed")
	}
}

func TestRetrieve(t *testing.T) {
	idx := NewIndex()

	content := `# Go Style Guide

## Error Handling

Always check errors in Go:

` + "```go" + `
if err != nil {
    return err
}
` + "```" + `

## Naming Conventions

Use camelCase for variable names and PascalCase for exported functions.

## Testing

Write unit tests for all public functions.

## Security

Never hardcode credentials. Use environment variables.
`

	_ = idx.LoadContent("GO_STYLE.md", content)

	// Query for error handling
	results := idx.Retrieve(RetrievalQuery{
		Language:    "go",
		CodeContext: "if err != nil { return err }",
		Tags:        []string{"error"},
	}, 3)

	if len(results) == 0 {
		t.Fatal("Expected results for error handling query")
	}

	// First result should be about error handling
	found := false
	for _, r := range results {
		if strings.Contains(r.Section.Title, "Error") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected error handling section in top results")
	}

	// Query for security
	securityResults := idx.Retrieve(RetrievalQuery{
		CodeContext: "password := os.Getenv(\"PASSWORD\")",
		Tags:        []string{"security"},
	}, 2)

	if len(securityResults) == 0 {
		t.Fatal("Expected results for security query")
	}
}

func TestFormatForPrompt(t *testing.T) {
	results := []RetrievalResult{
		{
			Section: RuleSection{
				Title:   "Error Handling",
				Content: "Always handle errors properly.",
			},
			Score: 10.0,
		},
		{
			Section: RuleSection{
				Title:   "Naming",
				Content: "Use descriptive names.",
			},
			Score: 5.0,
		},
	}

	formatted := FormatForPrompt(results, 1000)

	if !strings.Contains(formatted, "Style Guide Rules") {
		t.Error("Should contain header")
	}

	if !strings.Contains(formatted, "Error Handling") {
		t.Error("Should contain first section title")
	}

	if !strings.Contains(formatted, "Naming") {
		t.Error("Should contain second section title")
	}
}

func TestFormatForPromptTruncation(t *testing.T) {
	results := []RetrievalResult{
		{
			Section: RuleSection{
				Title:   "Long Section",
				Content: strings.Repeat("This is a very long content. ", 100),
			},
			Score: 10.0,
		},
	}

	formatted := FormatForPrompt(results, 200)

	if len(formatted) > 300 { // Some buffer for truncation text
		t.Errorf("Formatted result too long: %d chars", len(formatted))
	}

	if !strings.Contains(formatted, "truncated") {
		t.Error("Should indicate truncation")
	}
}

func TestInferTagsFromCode(t *testing.T) {
	tests := []struct {
		code         string
		expectedTags []string
	}{
		{
			code:         "if err != nil { return err }",
			expectedTags: []string{"error"},
		},
		{
			code:         "log.Printf(\"debug: %v\", data)",
			expectedTags: []string{"logging"},
		},
		{
			code:         "func TestSomething(t *testing.T) {}",
			expectedTags: []string{"testing"},
		},
		{
			code:         "go func() { channel <- data }()",
			expectedTags: []string{"concurrency"},
		},
		{
			code:         "password := decrypt(secret)",
			expectedTags: []string{"security"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.code[:20], func(t *testing.T) {
			tags := inferTagsFromCode(tt.code)

			for _, expected := range tt.expectedTags {
				found := false
				for _, tag := range tags {
					if tag == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag '%s' not found in %v for code: %s",
						expected, tags, tt.code)
				}
			}
		})
	}
}

func TestCalculatePriority(t *testing.T) {
	// Higher level header = higher priority
	p1 := calculatePriority(1, "Important Rule", "content")
	p2 := calculatePriority(3, "Sub Rule", "content")

	if p1 <= p2 {
		t.Error("H1 should have higher priority than H3")
	}

	// Keywords boost priority
	pMust := calculatePriority(2, "Must Do This", "This is required.")
	pMay := calculatePriority(2, "May Do This", "This is optional.")

	if pMust <= pMay {
		t.Error("Section with 'must'/'required' should have higher priority")
	}

	// Code examples boost priority
	pWithCode := calculatePriority(2, "Example", "```code```")
	pWithoutCode := calculatePriority(2, "Example", "no code here")

	if pWithCode <= pWithoutCode {
		t.Error("Section with code should have higher priority")
	}
}

func BenchmarkRetrieve(b *testing.B) {
	idx := NewIndex()

	// Load a substantial style guide
	content := `# Complete Style Guide

` + strings.Repeat(`
## Section

This is content about programming best practices.

`, 50)

	_ = idx.LoadContent("guide.md", content)

	query := RetrievalQuery{
		Language:    "go",
		CodeContext: "func test() error { return nil }",
		Tags:        []string{"error", "testing"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx.Retrieve(query, 5)
	}
}
