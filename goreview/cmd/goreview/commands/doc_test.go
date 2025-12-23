package commands

import (
	"strings"
	"testing"

	"github.com/JNZader/goreview/goreview/internal/git"
)

func TestBuildDocContext(t *testing.T) {
	diff := &git.Diff{
		Files: []git.FileDiff{
			{Path: "main.go", Status: git.FileModified},
			{Path: "utils.go", Status: git.FileAdded},
		},
	}

	tests := []struct {
		docType     string
		style       string
		wantContain string
	}{
		{"changelog", "markdown", "CHANGELOG.md"},
		{"api", "markdown", "API documentation"},
		{"readme", "markdown", "README content"},
		{"changes", "markdown", "Summarize the changes"},
	}

	for _, tt := range tests {
		t.Run(tt.docType, func(t *testing.T) {
			result := buildDocContext(diff, tt.docType, tt.style, "")
			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("buildDocContext() should contain %q, got %q", tt.wantContain, result)
			}
		})
	}
}

func TestBuildDocContextWithCustomContext(t *testing.T) {
	diff := &git.Diff{
		Files: []git.FileDiff{
			{Path: "main.go", Status: git.FileModified},
		},
	}

	customCtx := "This is custom context for testing"
	result := buildDocContext(diff, "changes", "markdown", customCtx)

	if !strings.Contains(result, "Additional context:") {
		t.Error("Should contain 'Additional context:' section")
	}
	if !strings.Contains(result, customCtx) {
		t.Errorf("Should contain custom context %q", customCtx)
	}
}

func TestBuildDocContextFileSummary(t *testing.T) {
	diff := &git.Diff{
		Files: []git.FileDiff{
			{Path: "main.go", Status: git.FileModified},
			{Path: "new.go", Status: git.FileAdded},
			{Path: "old.go", Status: git.FileDeleted},
		},
	}

	result := buildDocContext(diff, "changes", "markdown", "")

	if !strings.Contains(result, "Files changed:") {
		t.Error("Should contain 'Files changed:' section")
	}
	if !strings.Contains(result, "main.go") {
		t.Error("Should list main.go")
	}
	if !strings.Contains(result, "new.go") {
		t.Error("Should list new.go")
	}
	if !strings.Contains(result, "old.go") {
		t.Error("Should list old.go")
	}
}

func TestFormatAsJSDoc(t *testing.T) {
	input := "This is a test\nSecond line"
	result := formatAsJSDoc(input)

	if !strings.HasPrefix(result, "/**") {
		t.Error("JSDoc should start with /**")
	}
	if !strings.HasSuffix(result, " */") {
		t.Error("JSDoc should end with */")
	}
	if !strings.Contains(result, " * This is a test") {
		t.Error("JSDoc should contain formatted lines")
	}
}

func TestFormatAsGoDoc(t *testing.T) {
	input := "This is a test\nSecond line"
	result := formatAsGoDoc(input)

	lines := strings.Split(result, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "// ") {
			t.Errorf("GoDoc line should start with '// ': %q", line)
		}
	}
}

func TestFormatDocOutput(t *testing.T) {
	tests := []struct {
		style       string
		wantPrefix  string
	}{
		{"jsdoc", "/**"},
		{"godoc", "// "},
		{"markdown", "Test content"},
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			result := formatDocOutput("Test content", tt.style)
			if !strings.HasPrefix(result, tt.wantPrefix) {
				t.Errorf("formatDocOutput(%q) should start with %q, got %q", tt.style, tt.wantPrefix, result)
			}
		})
	}
}

func TestNewChangelogData(t *testing.T) {
	doc := `Added:
- Feature A
- Feature B

Changed:
- Modified feature C

Fixed:
- Issue D resolved

Removed:
- Old feature E`

	data := NewChangelogData(doc)

	if len(data.Added) != 2 {
		t.Errorf("Added = %d, want 2", len(data.Added))
	}
	if len(data.Changed) != 1 {
		t.Errorf("Changed = %d, want 1", len(data.Changed))
	}
	if len(data.Fixed) != 1 {
		t.Errorf("Fixed = %d, want 1", len(data.Fixed))
	}
	if len(data.Removed) != 1 {
		t.Errorf("Removed = %d, want 1", len(data.Removed))
	}
}

func TestNewChangelogDataAlternativeKeywords(t *testing.T) {
	doc := `New features:
- Feature A

Updated:
- Feature B

Bug fixes:
- Issue C

Deleted:
- Feature D`

	data := NewChangelogData(doc)

	if len(data.Added) != 1 {
		t.Errorf("Added = %d, want 1 (using 'new' keyword)", len(data.Added))
	}
	if len(data.Changed) != 1 {
		t.Errorf("Changed = %d, want 1 (using 'updated' keyword)", len(data.Changed))
	}
	if len(data.Fixed) != 1 {
		t.Errorf("Fixed = %d, want 1 (using 'bug' keyword)", len(data.Fixed))
	}
	if len(data.Removed) != 1 {
		t.Errorf("Removed = %d, want 1 (using 'deleted' keyword)", len(data.Removed))
	}
}

func TestNewChangelogDataDate(t *testing.T) {
	data := NewChangelogData("Added:\n- Test item")

	if data.Date == "" {
		t.Error("Date should not be empty")
	}
	// Date format should be YYYY-MM-DD
	if len(data.Date) != 10 || data.Date[4] != '-' || data.Date[7] != '-' {
		t.Errorf("Date should be in YYYY-MM-DD format, got %q", data.Date)
	}
}

func TestLoadTemplate(t *testing.T) {
	// Test built-in templates
	builtins := []string{"changelog", "api", "readme-section"}

	for _, name := range builtins {
		t.Run(name, func(t *testing.T) {
			tmpl, err := LoadTemplate(name)
			if err != nil {
				t.Errorf("LoadTemplate(%q) error = %v", name, err)
			}
			if tmpl == nil {
				t.Error("Template should not be nil")
			}
		})
	}
}

func TestLoadTemplateNotFound(t *testing.T) {
	_, err := LoadTemplate("nonexistent-template-file.txt")
	if err == nil {
		t.Error("LoadTemplate should return error for nonexistent file")
	}
}

func TestFormatDiffForDoc(t *testing.T) {
	diff := &git.Diff{
		Files: []git.FileDiff{
			{
				Path: "test.go",
				Hunks: []git.Hunk{
					{
						Lines: []git.Line{
							{Type: git.LineAddition, Content: "new line 1"},
							{Type: git.LineAddition, Content: "new line 2"},
							{Type: git.LineDeletion, Content: "deleted line"},
							{Type: git.LineContext, Content: "context line"},
						},
					},
				},
			},
		},
	}

	result := formatDiffForDoc(diff)

	if !strings.Contains(result, "=== test.go ===") {
		t.Error("Should contain file header")
	}
	if !strings.Contains(result, "+ new line 1") {
		t.Error("Should contain additions with + prefix")
	}
	if !strings.Contains(result, "+ new line 2") {
		t.Error("Should contain additions with + prefix")
	}
	if strings.Contains(result, "deleted line") {
		t.Error("Should not contain deletions")
	}
	if strings.Contains(result, "context line") {
		t.Error("Should not contain context lines")
	}
}
