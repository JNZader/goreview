# Iteracion 11: Comando Doc

## Objetivos

- Implementar comando `goreview doc`
- Generar documentacion automatica
- Soportar multiples formatos (MD, JSDoc, GoDoc)
- Integracion con contexto del proyecto

## Tiempo Estimado: 4 horas

---

## Commit 11.1: Crear estructura del comando doc

**Mensaje de commit:**
```
feat(cli): add doc command structure

- Add doc.go command file
- Define documentation flags
- Support multiple output types
```

### `goreview/cmd/doc.go`

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var docCmd = &cobra.Command{
	Use:   "doc [files...]",
	Short: "Generate documentation from code changes",
	Long: `Generate documentation for code changes using AI analysis.

Examples:
  # Generate docs for staged changes
  goreview doc --staged

  # Generate docs for specific files
  goreview doc src/main.go src/utils.go

  # Generate changelog entry
  goreview doc --staged --type changelog

  # Generate API documentation
  goreview doc --files "**/*.go" --type api

  # Output to file
  goreview doc --staged -o CHANGELOG.md`,
	RunE: runDoc,
}

func init() {
	rootCmd.AddCommand(docCmd)

	// Input flags
	docCmd.Flags().Bool("staged", false, "Document staged changes")
	docCmd.Flags().String("commit", "", "Document a specific commit")
	docCmd.Flags().String("range", "", "Document commit range (from..to)")

	// Type flags
	docCmd.Flags().StringP("type", "t", "changes", "Documentation type (changes, changelog, api, readme)")
	docCmd.Flags().String("style", "markdown", "Output style (markdown, jsdoc, godoc)")

	// Context flags
	docCmd.Flags().String("context", "", "Additional context for generation")
	docCmd.Flags().String("template", "", "Custom template file")

	// Output flags
	docCmd.Flags().StringP("output", "o", "", "Write to file")
	docCmd.Flags().Bool("append", false, "Append to existing file")
	docCmd.Flags().Bool("prepend", false, "Prepend to existing file")
}

func runDoc(cmd *cobra.Command, args []string) error {
	fmt.Println("Doc command - implementation follows")
	return nil
}
```

---

## Commit 11.2: Implementar generacion de documentacion

**Mensaje de commit:**
```
feat(cli): implement documentation generation

- Analyze code changes
- Generate documentation based on type
- Support different styles
- Handle context and templates
```

### Actualizar `goreview/cmd/doc.go`:

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/git"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

func runDoc(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Initialize git repo
	gitRepo, err := git.NewRepository(".")
	if err != nil {
		return fmt.Errorf("initializing git: %w", err)
	}

	// Get diff based on mode
	diff, err := getDocDiff(cmd, args, gitRepo, ctx)
	if err != nil {
		return err
	}

	if len(diff.Files) == 0 {
		return fmt.Errorf("no changes found to document")
	}

	// Initialize provider
	provider, err := providers.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("initializing provider: %w", err)
	}
	defer provider.Close()

	// Build documentation context
	docType, _ := cmd.Flags().GetString("type")
	style, _ := cmd.Flags().GetString("style")
	customContext, _ := cmd.Flags().GetString("context")

	docContext := buildDocContext(diff, docType, style, customContext)

	// Generate documentation
	diffText := formatDiffForDoc(diff)
	documentation, err := provider.GenerateDocumentation(ctx, diffText, docContext)
	if err != nil {
		return fmt.Errorf("generating documentation: %w", err)
	}

	// Format output
	output := formatDocOutput(documentation, style)

	// Write output
	outputFile, _ := cmd.Flags().GetString("output")
	append, _ := cmd.Flags().GetBool("append")
	prepend, _ := cmd.Flags().GetBool("prepend")

	if outputFile != "" {
		return writeDocOutput(outputFile, output, append, prepend)
	}

	fmt.Print(output)
	return nil
}

func getDocDiff(cmd *cobra.Command, args []string, repo git.Repository, ctx context.Context) (*git.Diff, error) {
	if staged, _ := cmd.Flags().GetBool("staged"); staged {
		return repo.GetStagedDiff(ctx)
	}

	if commit, _ := cmd.Flags().GetString("commit"); commit != "" {
		return repo.GetCommitDiff(ctx, commit)
	}

	if len(args) > 0 {
		return repo.GetFileDiff(ctx, args)
	}

	return nil, fmt.Errorf("specify --staged, --commit, or file arguments")
}

func buildDocContext(diff *git.Diff, docType, style, customContext string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generate %s documentation in %s format.\n\n", docType, style))

	switch docType {
	case "changelog":
		sb.WriteString("Format as a CHANGELOG.md entry with:\n")
		sb.WriteString("- Version header (use [Unreleased])\n")
		sb.WriteString("- Grouped by: Added, Changed, Fixed, Removed\n")
		sb.WriteString("- Each item as a bullet point\n")
	case "api":
		sb.WriteString("Generate API documentation including:\n")
		sb.WriteString("- Function signatures\n")
		sb.WriteString("- Parameter descriptions\n")
		sb.WriteString("- Return values\n")
		sb.WriteString("- Example usage\n")
	case "readme":
		sb.WriteString("Generate README content including:\n")
		sb.WriteString("- Feature description\n")
		sb.WriteString("- Usage examples\n")
		sb.WriteString("- Configuration options\n")
	default: // changes
		sb.WriteString("Summarize the changes:\n")
		sb.WriteString("- What was changed\n")
		sb.WriteString("- Why it was changed\n")
		sb.WriteString("- How to use the new features\n")
	}

	if customContext != "" {
		sb.WriteString("\nAdditional context:\n")
		sb.WriteString(customContext)
	}

	// Add file summary
	sb.WriteString("\n\nFiles changed:\n")
	for _, f := range diff.Files {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", f.Path, f.Status))
	}

	return sb.String()
}

func formatDiffForDoc(diff *git.Diff) string {
	var sb strings.Builder

	for _, file := range diff.Files {
		sb.WriteString(fmt.Sprintf("\n=== %s ===\n", file.Path))
		for _, hunk := range file.Hunks {
			for _, line := range hunk.Lines {
				if line.Type == git.LineAddition {
					sb.WriteString("+ " + line.Content + "\n")
				}
			}
		}
	}

	return sb.String()
}

func formatDocOutput(doc, style string) string {
	switch style {
	case "jsdoc":
		return formatAsJSDoc(doc)
	case "godoc":
		return formatAsGoDoc(doc)
	default:
		return doc
	}
}

func formatAsJSDoc(doc string) string {
	lines := strings.Split(doc, "\n")
	var result []string
	result = append(result, "/**")
	for _, line := range lines {
		result = append(result, " * "+line)
	}
	result = append(result, " */")
	return strings.Join(result, "\n")
}

func formatAsGoDoc(doc string) string {
	lines := strings.Split(doc, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, "// "+line)
	}
	return strings.Join(result, "\n")
}

func writeDocOutput(path, content string, appendMode, prependMode bool) error {
	if appendMode {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString("\n" + content)
		return err
	}

	if prependMode {
		existing, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		content = content + "\n" + string(existing)
	}

	return os.WriteFile(path, []byte(content), 0644)
}
```

---

## Commit 11.3: Agregar templates de documentacion

**Mensaje de commit:**
```
feat(cli): add documentation templates

- Changelog template
- API documentation template
- README template
- Custom template support
```

### `goreview/cmd/doc_templates.go`

```go
package cmd

import (
	"os"
	"strings"
	"text/template"
	"time"
)

// DocTemplate represents a documentation template.
type DocTemplate struct {
	Name    string
	Content string
}

var defaultTemplates = map[string]*DocTemplate{
	"changelog": {
		Name: "changelog",
		Content: `## [Unreleased] - {{.Date}}

### Added
{{range .Added}}
- {{.}}
{{end}}

### Changed
{{range .Changed}}
- {{.}}
{{end}}

### Fixed
{{range .Fixed}}
- {{.}}
{{end}}

### Removed
{{range .Removed}}
- {{.}}
{{end}}
`,
	},
	"api": {
		Name: "api",
		Content: `# API Documentation

{{range .Functions}}
## {{.Name}}

{{.Description}}

### Signature

` + "```" + `{{.Language}}
{{.Signature}}
` + "```" + `

### Parameters

{{range .Parameters}}
- ` + "`{{.Name}}`" + ` ({{.Type}}): {{.Description}}
{{end}}

### Returns

{{.Returns}}

### Example

` + "```" + `{{.Language}}
{{.Example}}
` + "```" + `

{{end}}
`,
	},
	"readme-section": {
		Name: "readme-section",
		Content: `## {{.Title}}

{{.Description}}

### Usage

` + "```" + `{{.Language}}
{{.Usage}}
` + "```" + `

### Configuration

{{range .Config}}
- ` + "`{{.Name}}`" + `: {{.Description}} (default: {{.Default}})
{{end}}
`,
	},
}

// LoadTemplate loads a template by name or from file.
func LoadTemplate(nameOrPath string) (*template.Template, error) {
	// Check if it's a built-in template
	if tmpl, ok := defaultTemplates[nameOrPath]; ok {
		return template.New(tmpl.Name).Parse(tmpl.Content)
	}

	// Try to load from file
	content, err := os.ReadFile(nameOrPath)
	if err != nil {
		return nil, err
	}

	return template.New("custom").Parse(string(content))
}

// ChangelogData represents data for changelog template.
type ChangelogData struct {
	Date    string
	Added   []string
	Changed []string
	Fixed   []string
	Removed []string
}

// NewChangelogData creates changelog data from documentation.
func NewChangelogData(doc string) *ChangelogData {
	data := &ChangelogData{
		Date: time.Now().Format("2006-01-02"),
	}

	lines := strings.Split(doc, "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		lower := strings.ToLower(line)
		if strings.Contains(lower, "added") || strings.Contains(lower, "new") {
			currentSection = "added"
			continue
		}
		if strings.Contains(lower, "changed") || strings.Contains(lower, "updated") {
			currentSection = "changed"
			continue
		}
		if strings.Contains(lower, "fixed") || strings.Contains(lower, "bug") {
			currentSection = "fixed"
			continue
		}
		if strings.Contains(lower, "removed") || strings.Contains(lower, "deleted") {
			currentSection = "removed"
			continue
		}

		// Add to current section
		item := strings.TrimPrefix(line, "- ")
		item = strings.TrimPrefix(item, "* ")

		switch currentSection {
		case "added":
			data.Added = append(data.Added, item)
		case "changed":
			data.Changed = append(data.Changed, item)
		case "fixed":
			data.Fixed = append(data.Fixed, item)
		case "removed":
			data.Removed = append(data.Removed, item)
		}
	}

	return data
}
```

---

## Commit 11.4: Tests del comando doc

**Mensaje de commit:**
```
test(cli): add doc command tests

- Test documentation generation
- Test template loading
- Test output formatting
```

### `goreview/cmd/doc_test.go`

```go
package cmd

import (
	"strings"
	"testing"
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
				t.Errorf("buildDocContext() should contain %q", tt.wantContain)
			}
		})
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

func TestNewChangelogData(t *testing.T) {
	doc := `Added:
- New feature A
- New feature B

Changed:
- Updated feature C

Fixed:
- Bug fix D

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
```

---

## Resumen de la Iteracion 11

### Commits:
1. `feat(cli): add doc command structure`
2. `feat(cli): implement documentation generation`
3. `feat(cli): add documentation templates`
4. `test(cli): add doc command tests`

### Archivos:
```
goreview/cmd/
├── doc.go
├── doc_templates.go
└── doc_test.go
```

---

## Siguiente Iteracion

Continua con: **[12-COMANDO-INIT.md](12-COMANDO-INIT.md)**
