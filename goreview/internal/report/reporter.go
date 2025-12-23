package report

import (
	"fmt"
	"io"

	"github.com/JNZader/goreview/goreview/internal/review"
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
		return &JSONReporter{Indent: true}, nil
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
