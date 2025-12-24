package commands

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
	content, err := os.ReadFile(nameOrPath) //nolint:gosec // CLI tool loads user-specified template files
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

// APIFunctionData represents data for API documentation.
type APIFunctionData struct {
	Name        string
	Description string
	Language    string
	Signature   string
	Parameters  []ParameterData
	Returns     string
	Example     string
}

// ParameterData represents parameter documentation.
type ParameterData struct {
	Name        string
	Type        string
	Description string
}

// APIDocData represents data for API documentation template.
type APIDocData struct {
	Functions []APIFunctionData
}

// ReadmeSectionData represents data for readme section template.
type ReadmeSectionData struct {
	Title       string
	Description string
	Language    string
	Usage       string
	Config      []ConfigData
}

// ConfigData represents configuration documentation.
type ConfigData struct {
	Name        string
	Description string
	Default     string
}
