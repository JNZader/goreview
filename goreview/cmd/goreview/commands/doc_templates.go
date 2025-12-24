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

// sectionKeywords maps changelog section names to their detection keywords
var sectionKeywords = map[string][]string{
	"added":   {"added", "new"},
	"changed": {"changed", "updated"},
	"fixed":   {"fixed", "bug"},
	"removed": {"removed", "deleted"},
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

		if section := detectSection(line); section != "" {
			currentSection = section
			continue
		}

		item := cleanItem(line)
		data.addItem(currentSection, item)
	}

	return data
}

// detectSection checks if a line is a section header
func detectSection(line string) string {
	lower := strings.ToLower(line)
	for section, keywords := range sectionKeywords {
		for _, kw := range keywords {
			if strings.Contains(lower, kw) {
				return section
			}
		}
	}
	return ""
}

// cleanItem removes common list prefixes from an item
func cleanItem(line string) string {
	item := strings.TrimPrefix(line, "- ")
	return strings.TrimPrefix(item, "* ")
}

// addItem adds an item to the appropriate changelog section
func (d *ChangelogData) addItem(section, item string) {
	switch section {
	case "added":
		d.Added = append(d.Added, item)
	case "changed":
		d.Changed = append(d.Changed, item)
	case "fixed":
		d.Fixed = append(d.Fixed, item)
	case "removed":
		d.Removed = append(d.Removed, item)
	}
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
