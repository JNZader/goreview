package report

import (
	"fmt"
	"io"
	"strings"

	"github.com/JNZader/goreview/goreview/internal/providers"
	"github.com/JNZader/goreview/goreview/internal/review"
)

// MarkdownReporter generates Markdown reports.
type MarkdownReporter struct{}

func (r *MarkdownReporter) Format() string { return "markdown" }

func (r *MarkdownReporter) Generate(result *review.Result) (string, error) {
	var sb strings.Builder
	_ = r.Write(result, &sb)
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
