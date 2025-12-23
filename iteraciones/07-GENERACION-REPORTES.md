# Iteracion 07: Generacion de Reportes

## Objetivos

- Interface Reporter para abstraccion
- Implementacion Markdown (human-readable)
- Implementacion JSON (machine-readable)
- Implementacion SARIF (estandar de analisis)

## Tiempo Estimado: 4 horas

---

## Commit 7.1: Crear interface Reporter

**Mensaje de commit:**
```
feat(report): add reporter interface

- Define Reporter interface
- Support multiple output formats
- Factory for creating reporters
```

### `goreview/internal/report/reporter.go`

```go
package report

import (
	"fmt"
	"io"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/review"
)

// Reporter defines the interface for generating review reports.
type Reporter interface {
	// Generate creates a report from review results.
	Generate(result *review.Result) (string, error)

	// Write writes the report to a writer.
	Write(result *review.Result, w io.Writer) error

	// Format returns the format name.
	Format() string
}

// NewReporter creates a reporter for the given format.
func NewReporter(format string) (Reporter, error) {
	switch format {
	case "markdown", "md":
		return &MarkdownReporter{}, nil
	case "json":
		return &JSONReporter{}, nil
	case "sarif":
		return &SARIFReporter{}, nil
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

// AvailableFormats returns the list of supported formats.
func AvailableFormats() []string {
	return []string{"markdown", "json", "sarif"}
}
```

---

## Commit 7.2: Implementar Markdown Reporter

**Mensaje de commit:**
```
feat(report): add markdown reporter

- Generate human-readable markdown
- Include issue details with code references
- Summary section with statistics
- Color-coded severity indicators
```

### `goreview/internal/report/markdown.go`

```go
package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/review"
)

// MarkdownReporter generates Markdown reports.
type MarkdownReporter struct{}

func (r *MarkdownReporter) Format() string { return "markdown" }

func (r *MarkdownReporter) Generate(result *review.Result) (string, error) {
	var sb strings.Builder
	r.Write(result, &sb)
	return sb.String(), nil
}

func (r *MarkdownReporter) Write(result *review.Result, w io.Writer) error {
	// Header
	fmt.Fprintf(w, "# Code Review Report\n\n")

	// Summary
	fmt.Fprintf(w, "## Summary\n\n")
	fmt.Fprintf(w, "- **Files Reviewed:** %d\n", len(result.Files))
	fmt.Fprintf(w, "- **Total Issues:** %d\n", result.TotalIssues)
	fmt.Fprintf(w, "- **Duration:** %s\n", result.Duration)
	fmt.Fprintf(w, "\n")

	if result.TotalIssues == 0 {
		fmt.Fprintf(w, "No issues found.\n\n")
		return nil
	}

	// Issues by file
	fmt.Fprintf(w, "## Issues\n\n")

	for _, file := range result.Files {
		if file.Error != nil {
			fmt.Fprintf(w, "### %s\n\n", file.File)
			fmt.Fprintf(w, "Error: %v\n\n", file.Error)
			continue
		}

		if file.Response == nil || len(file.Response.Issues) == 0 {
			continue
		}

		fmt.Fprintf(w, "### %s\n\n", file.File)

		if file.Cached {
			fmt.Fprintf(w, "_Cached result_\n\n")
		}

		for _, issue := range file.Response.Issues {
			r.writeIssue(w, issue)
		}
	}

	return nil
}

func (r *MarkdownReporter) writeIssue(w io.Writer, issue providers.Issue) {
	// Severity icon
	icon := r.severityIcon(issue.Severity)

	fmt.Fprintf(w, "#### %s [%s] %s\n\n", icon, issue.Type, issue.Message)

	if issue.Location != nil && issue.Location.StartLine > 0 {
		fmt.Fprintf(w, "**Location:** Line %d", issue.Location.StartLine)
		if issue.Location.EndLine > issue.Location.StartLine {
			fmt.Fprintf(w, "-%d", issue.Location.EndLine)
		}
		fmt.Fprintf(w, "\n\n")
	}

	if issue.Suggestion != "" {
		fmt.Fprintf(w, "**Suggestion:** %s\n\n", issue.Suggestion)
	}

	if issue.FixedCode != "" {
		fmt.Fprintf(w, "**Suggested Fix:**\n```\n%s\n```\n\n", issue.FixedCode)
	}

	fmt.Fprintf(w, "---\n\n")
}

func (r *MarkdownReporter) severityIcon(severity providers.Severity) string {
	switch severity {
	case providers.SeverityCritical:
		return "[CRITICAL]"
	case providers.SeverityError:
		return "[ERROR]"
	case providers.SeverityWarning:
		return "[WARNING]"
	default:
		return "[INFO]"
	}
}
```

---

## Commit 7.3: Implementar JSON Reporter

**Mensaje de commit:**
```
feat(report): add json reporter

- Generate machine-readable JSON
- Include all metadata
- Pretty-print with indentation
```

### `goreview/internal/report/json.go`

```go
package report

import (
	"encoding/json"
	"io"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/review"
)

// JSONReporter generates JSON reports.
type JSONReporter struct {
	Indent bool
}

func (r *JSONReporter) Format() string { return "json" }

func (r *JSONReporter) Generate(result *review.Result) (string, error) {
	var data []byte
	var err error

	if r.Indent {
		data, err = json.MarshalIndent(result, "", "  ")
	} else {
		data, err = json.Marshal(result)
	}

	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *JSONReporter) Write(result *review.Result, w io.Writer) error {
	encoder := json.NewEncoder(w)
	if r.Indent {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(result)
}
```

---

## Commit 7.4: Implementar SARIF Reporter

**Mensaje de commit:**
```
feat(report): add sarif reporter

- Generate SARIF 2.1.0 format
- Compatible with GitHub Code Scanning
- Include rule definitions
- Map severity levels
```

### `goreview/internal/report/sarif.go`

```go
package report

import (
	"encoding/json"
	"io"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/review"
)

// SARIFReporter generates SARIF 2.1.0 reports.
type SARIFReporter struct{}

func (r *SARIFReporter) Format() string { return "sarif" }

// SARIF types
type sarifReport struct {
	Schema  string      `json:"$schema"`
	Version string      `json:"version"`
	Runs    []sarifRun  `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Rules   []sarifRule `json:"rules,omitempty"`
}

type sarifRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description struct {
		Text string `json:"text"`
	} `json:"shortDescription"`
}

type sarifResult struct {
	RuleID    string           `json:"ruleId"`
	Level     string           `json:"level"`
	Message   sarifMessage     `json:"message"`
	Locations []sarifLocation  `json:"locations,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation struct {
		ArtifactLocation struct {
			URI string `json:"uri"`
		} `json:"artifactLocation"`
		Region *sarifRegion `json:"region,omitempty"`
	} `json:"physicalLocation"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	EndLine     int `json:"endLine,omitempty"`
	StartColumn int `json:"startColumn,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

func (r *SARIFReporter) Generate(result *review.Result) (string, error) {
	report := r.buildReport(result)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (r *SARIFReporter) Write(result *review.Result, w io.Writer) error {
	report := r.buildReport(result)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func (r *SARIFReporter) buildReport(result *review.Result) *sarifReport {
	report := &sarifReport{
		Schema:  "https://json.schemastore.org/sarif-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{{
			Tool: sarifTool{
				Driver: sarifDriver{
					Name:    "goreview",
					Version: "1.0.0",
				},
			},
			Results: []sarifResult{},
		}},
	}

	for _, file := range result.Files {
		if file.Response == nil {
			continue
		}

		for _, issue := range file.Response.Issues {
			sarifResult := sarifResult{
				RuleID:  string(issue.Type),
				Level:   r.mapLevel(issue.Severity),
				Message: sarifMessage{Text: issue.Message},
			}

			if issue.Location != nil {
				loc := sarifLocation{}
				loc.PhysicalLocation.ArtifactLocation.URI = file.File
				if issue.Location.StartLine > 0 {
					loc.PhysicalLocation.Region = &sarifRegion{
						StartLine: issue.Location.StartLine,
						EndLine:   issue.Location.EndLine,
					}
				}
				sarifResult.Locations = append(sarifResult.Locations, loc)
			}

			report.Runs[0].Results = append(report.Runs[0].Results, sarifResult)
		}
	}

	return report
}

func (r *SARIFReporter) mapLevel(severity providers.Severity) string {
	switch severity {
	case providers.SeverityCritical, providers.SeverityError:
		return "error"
	case providers.SeverityWarning:
		return "warning"
	default:
		return "note"
	}
}
```

---

## Resumen de la Iteracion 07

### Commits:
1. `feat(report): add reporter interface`
2. `feat(report): add markdown reporter`
3. `feat(report): add json reporter`
4. `feat(report): add sarif reporter`

### Archivos:
```
goreview/internal/report/
├── reporter.go
├── markdown.go
├── json.go
└── sarif.go
```

---

## Siguiente Iteracion

Continua con: **[08-SISTEMA-REGLAS.md](08-SISTEMA-REGLAS.md)**
