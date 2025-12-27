package export

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/providers"
	"github.com/JNZader/goreview/goreview/internal/review"
)

// ObsidianExporter exports review results to an Obsidian vault.
type ObsidianExporter struct {
	cfg      *config.ObsidianExportConfig
	template *template.Template
}

// obsidianTemplateData holds all data passed to the template.
type obsidianTemplateData struct {
	Frontmatter    *ObsidianFrontmatter
	Result         *review.Result
	Metadata       *ExportMetadata
	Config         *config.ObsidianExportConfig
	RelatedReviews []string
}

// NewObsidianExporter creates a new Obsidian exporter.
func NewObsidianExporter(cfg *config.ObsidianExportConfig) (*ObsidianExporter, error) {
	e := &ObsidianExporter{cfg: cfg}

	if err := e.validate(); err != nil {
		return nil, err
	}

	if err := e.loadTemplate(); err != nil {
		return nil, err
	}

	return e, nil
}

// Name returns the exporter name.
func (e *ObsidianExporter) Name() string {
	return "obsidian"
}

// validate validates the exporter configuration.
func (e *ObsidianExporter) validate() error {
	if e.cfg.VaultPath == "" {
		return fmt.Errorf("obsidian vault path is required")
	}

	// Expand ~ in path
	vaultPath := expandPath(e.cfg.VaultPath)
	e.cfg.VaultPath = vaultPath

	// Check vault exists
	if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
		return fmt.Errorf("obsidian vault not found: %s", vaultPath)
	}

	return nil
}

// loadTemplate loads the export template.
func (e *ObsidianExporter) loadTemplate() error {
	// Use custom template if specified
	if e.cfg.TemplateFile != "" {
		tmplPath := expandPath(e.cfg.TemplateFile)
		content, err := os.ReadFile(tmplPath) // #nosec G304 - user-provided template path
		if err != nil {
			return fmt.Errorf("loading custom template: %w", err)
		}
		tmpl, err := template.New("obsidian").Funcs(e.templateFuncs()).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parsing custom template: %w", err)
		}
		e.template = tmpl
		return nil
	}

	// Use default template
	tmpl, err := template.New("obsidian").Funcs(e.templateFuncs()).Parse(defaultObsidianTemplate)
	if err != nil {
		return fmt.Errorf("parsing default template: %w", err)
	}
	e.template = tmpl
	return nil
}

// templateFuncs returns the template functions.
func (e *ObsidianExporter) templateFuncs() template.FuncMap {
	return template.FuncMap{
		"severityIcon":    severityIcon,
		"severityCallout": severityCallout,
		"formatTags":      formatTags,
		"wikiLink":        wikiLink,
		"formatTime":      formatTime,
	}
}

// Export exports the review result to the Obsidian vault.
func (e *ObsidianExporter) Export(result *review.Result, metadata *ExportMetadata) error {
	// Ensure project directory exists
	projectDir := filepath.Join(e.cfg.VaultPath, e.cfg.FolderName, sanitizeFilename(metadata.ProjectName))
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("creating project directory: %w", err)
	}

	// Generate filename
	filename := e.generateFilename(projectDir, metadata)

	// Build frontmatter
	frontmatter := e.buildFrontmatter(result, metadata)

	// Find related reviews for linking
	var relatedReviews []string
	if e.cfg.LinkToPreviousReviews {
		relatedReviews = e.findRelatedReviews(projectDir, filename)
	}

	// Prepare template data
	data := &obsidianTemplateData{
		Frontmatter:    frontmatter,
		Result:         result,
		Metadata:       metadata,
		Config:         e.cfg,
		RelatedReviews: relatedReviews,
	}

	// Execute template
	var sb strings.Builder
	if err := e.template.Execute(&sb, data); err != nil {
		return fmt.Errorf("executing template: %w", err)
	}

	// Write file
	outputPath := filepath.Join(projectDir, filename)
	if err := os.WriteFile(outputPath, []byte(sb.String()), 0644); err != nil {
		return fmt.Errorf("writing export file: %w", err)
	}

	return nil
}

// GetOutputPath returns the path where the export file will be written.
func (e *ObsidianExporter) GetOutputPath(metadata *ExportMetadata) string {
	projectDir := filepath.Join(e.cfg.VaultPath, e.cfg.FolderName, sanitizeFilename(metadata.ProjectName))
	filename := e.generateFilename(projectDir, metadata)
	return filepath.Join(projectDir, filename)
}

// generateFilename generates the filename for the export.
func (e *ObsidianExporter) generateFilename(projectDir string, metadata *ExportMetadata) string {
	// Find next review number
	reviewNum := e.findNextReviewNumber(projectDir)

	// Format: review-001-2024-01-15.md
	date := metadata.ReviewDate.Format("2006-01-02")
	return fmt.Sprintf("review-%03d-%s.md", reviewNum, date)
}

// findNextReviewNumber finds the next review number in the directory.
func (e *ObsidianExporter) findNextReviewNumber(projectDir string) int {
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return 1
	}

	maxNum := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "review-") {
			continue
		}

		var num int
		if _, err := fmt.Sscanf(entry.Name(), "review-%d-", &num); err == nil {
			if num > maxNum {
				maxNum = num
			}
		}
	}

	return maxNum + 1
}

// buildFrontmatter builds the Obsidian frontmatter from the review result.
func (e *ObsidianExporter) buildFrontmatter(result *review.Result, metadata *ExportMetadata) *ObsidianFrontmatter {
	fm := &ObsidianFrontmatter{
		Date:          metadata.ReviewDate.Format(time.RFC3339),
		Project:       metadata.ProjectName,
		Branch:        metadata.Branch,
		Commit:        metadata.CommitHash,
		CommitShort:   metadata.CommitShort,
		Author:        metadata.Author,
		ReviewMode:    metadata.ReviewMode,
		FilesReviewed: len(result.Files),
		TotalIssues:   result.TotalIssues,
		Duration:      result.Duration.String(),
	}

	// Count issues by severity
	severityCounts := e.countBySeverity(result)
	fm.CriticalIssues = severityCounts["critical"]
	fm.ErrorIssues = severityCounts["error"]
	fm.WarningIssues = severityCounts["warning"]
	fm.InfoIssues = severityCounts["info"]

	// Calculate average score
	fm.AverageScore = e.calculateAverageScore(result)

	// Build tags
	fm.Tags = e.buildTags(result)

	return fm
}

// countBySeverity counts issues by severity.
func (e *ObsidianExporter) countBySeverity(result *review.Result) map[string]int {
	counts := make(map[string]int)
	for _, file := range result.Files {
		if file.Response == nil {
			continue
		}
		for _, issue := range file.Response.Issues {
			counts[string(issue.Severity)]++
		}
	}
	return counts
}

// calculateAverageScore calculates the average quality score.
func (e *ObsidianExporter) calculateAverageScore(result *review.Result) int {
	var total, count int
	for _, file := range result.Files {
		if file.Response != nil && file.Response.Score > 0 {
			total += file.Response.Score
			count++
		}
	}
	if count == 0 {
		return 100 // No issues = perfect score
	}
	return total / count
}

// buildTags builds the tags for the Obsidian note.
func (e *ObsidianExporter) buildTags(result *review.Result) []string {
	tags := []string{"goreview", "code-review"}

	// Add severity tags
	severityCounts := e.countBySeverity(result)
	if severityCounts["critical"] > 0 {
		tags = append(tags, "critical")
	}
	if severityCounts["error"] > 0 {
		tags = append(tags, "has-errors")
	}

	// Add issue type tags
	typeSet := make(map[string]bool)
	for _, file := range result.Files {
		if file.Response == nil {
			continue
		}
		for _, issue := range file.Response.Issues {
			typeSet[strings.ToLower(string(issue.Type))] = true
		}
	}
	for t := range typeSet {
		tags = append(tags, t)
	}

	// Add custom tags
	tags = append(tags, e.cfg.CustomTags...)

	// Sort and deduplicate
	sort.Strings(tags)
	return unique(tags)
}

// findRelatedReviews finds previous reviews in the same project directory.
func (e *ObsidianExporter) findRelatedReviews(projectDir string, currentFilename string) []string {
	var related []string
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return related
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		// Skip current file
		if entry.Name() == currentFilename {
			continue
		}
		// Create wiki link without .md extension
		name := strings.TrimSuffix(entry.Name(), ".md")
		related = append(related, name)
	}

	// Return most recent reviews first (they're sorted by name which includes date)
	sort.Sort(sort.Reverse(sort.StringSlice(related)))

	// Limit to last 5 reviews
	if len(related) > 5 {
		related = related[:5]
	}

	return related
}

// Template helper functions

// severityIcon returns an emoji icon for the severity level.
func severityIcon(severity providers.Severity) string {
	switch severity {
	case providers.SeverityCritical:
		return ":red_circle:"
	case providers.SeverityError:
		return ":orange_circle:"
	case providers.SeverityWarning:
		return ":yellow_circle:"
	default:
		return ":blue_circle:"
	}
}

// severityCallout returns the Obsidian callout type for the severity.
func severityCallout(severity providers.Severity) string {
	switch severity {
	case providers.SeverityCritical:
		return "danger"
	case providers.SeverityError:
		return "warning"
	case providers.SeverityWarning:
		return "caution"
	default:
		return "info"
	}
}

// formatTags formats tags as Obsidian hashtags.
func formatTags(tags []string) string {
	var result []string
	for _, t := range tags {
		result = append(result, "#"+t)
	}
	return strings.Join(result, " ")
}

// wikiLink creates an Obsidian wiki link.
func wikiLink(name string) string {
	return "[[" + name + "]]"
}

// formatTime formats a time for display.
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// Utility functions

// sanitizeFilename removes invalid characters from a filename.
func sanitizeFilename(name string) string {
	// Replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, c := range invalid {
		result = strings.ReplaceAll(result, c, "-")
	}
	return result
}

// unique returns a slice with duplicate strings removed.
func unique(s []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// expandPath expands ~ to the user's home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
