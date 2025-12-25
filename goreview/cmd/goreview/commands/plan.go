package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/providers"
)

var planCmd = &cobra.Command{
	Use:   "plan <document>",
	Short: "Review design documents and RFCs before implementation",
	Long: `Review design documents, RFCs, and implementation plans before writing code.

This command analyzes technical documents and provides feedback on:
- Feasibility and completeness of the design
- Potential architectural issues
- Security considerations
- Performance implications
- Missing edge cases
- Suggested improvements

Supported formats: Markdown (.md), plain text (.txt), reStructuredText (.rst)

Examples:
  # Review an RFC document
  goreview plan ./docs/RFC-001.md

  # Review with specific focus
  goreview plan ./docs/design.md --focus security

  # Review multiple documents
  goreview plan ./docs/RFC-001.md ./docs/RFC-002.md

  # Output as JSON
  goreview plan ./docs/design.md --format json`,
	Args: cobra.MinimumNArgs(1),
	RunE: runPlan,
}

var (
	planFocus     string
	planFormat    string
	planOutput    string
	planVerbose   bool
	planChecklist bool
)

func init() {
	rootCmd.AddCommand(planCmd)

	planCmd.Flags().StringVarP(&planFocus, "focus", "F", "", "Focus area: security, performance, scalability, maintainability, all (default: all)")
	planCmd.Flags().StringVarP(&planFormat, "format", "f", "markdown", "Output format: markdown, json")
	planCmd.Flags().StringVarP(&planOutput, "output", "o", "", "Write review to file")
	planCmd.Flags().BoolVarP(&planVerbose, "verbose", "V", false, "Include detailed analysis")
	planCmd.Flags().BoolVar(&planChecklist, "checklist", false, "Generate implementation checklist")
}

// PlanReview represents the review of a design document.
type PlanReview struct {
	Document    string          `json:"document"`
	ReviewedAt  time.Time       `json:"reviewed_at"`
	Summary     string          `json:"summary"`
	Score       PlanScore       `json:"score"`
	Strengths   []string        `json:"strengths"`
	Concerns    []PlanConcern   `json:"concerns"`
	Suggestions []string        `json:"suggestions"`
	Checklist   []ChecklistItem `json:"checklist,omitempty"`
}

// PlanScore represents the quality scores for a design.
type PlanScore struct {
	Overall      float64 `json:"overall"`
	Completeness float64 `json:"completeness"`
	Clarity      float64 `json:"clarity"`
	Feasibility  float64 `json:"feasibility"`
	Security     float64 `json:"security,omitempty"`
	Performance  float64 `json:"performance,omitempty"`
	Scalability  float64 `json:"scalability,omitempty"`
}

// PlanConcern represents a concern or potential issue in the design.
type PlanConcern struct {
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Suggestion  string `json:"suggestion,omitempty"`
	Section     string `json:"section,omitempty"`
}

// ChecklistItem represents an implementation checklist item.
type ChecklistItem struct {
	Task     string   `json:"task"`
	Priority string   `json:"priority"`
	Category string   `json:"category"`
	Depends  []string `json:"depends,omitempty"`
}

func runPlan(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	provider, err := providers.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("initializing provider: %w", err)
	}
	defer func() { _ = provider.Close() }()

	if healthErr := provider.HealthCheck(ctx); healthErr != nil {
		return fmt.Errorf("provider not available: %w", healthErr)
	}

	reviews := make([]*PlanReview, 0, len(args))
	for _, docPath := range args {
		review, reviewErr := reviewDocument(ctx, provider, docPath)
		if reviewErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to review %s: %v\n", docPath, reviewErr)
			continue
		}
		reviews = append(reviews, review)
	}

	if len(reviews) == 0 {
		return fmt.Errorf("no documents could be reviewed")
	}

	output, err := formatPlanOutput(reviews)
	if err != nil {
		return err
	}

	if planOutput != "" {
		if err := os.WriteFile(planOutput, []byte(output), 0600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Review written to %s\n", planOutput)
	} else {
		fmt.Print(output)
	}

	return nil
}

func reviewDocument(ctx context.Context, provider providers.Provider, docPath string) (*PlanReview, error) {
	content, err := os.ReadFile(docPath)
	if err != nil {
		return nil, fmt.Errorf("reading document: %w", err)
	}

	prompt := buildPlanPrompt(string(content), docPath)

	// Use GenerateDocumentation as it handles free-form text generation
	response, err := provider.GenerateDocumentation(ctx, string(content), prompt)
	if err != nil {
		return nil, fmt.Errorf("getting AI response: %w", err)
	}

	review, err := parsePlanResponse(response)
	if err != nil {
		// Fallback to basic review structure
		review = &PlanReview{
			Summary: response,
			Score: PlanScore{
				Overall: 70,
			},
		}
	}

	review.Document = docPath
	review.ReviewedAt = time.Now()

	return review, nil
}

func buildPlanPrompt(content, docPath string) string {
	docType := detectDocumentType(docPath)
	focusInstructions := getFocusInstructions()

	prompt := fmt.Sprintf(`You are a senior software architect reviewing a %s document.

Document path: %s

%s

Please analyze this document and provide a detailed review in the following JSON format:
{
  "summary": "Brief summary of the document and overall assessment",
  "score": {
    "overall": 75,
    "completeness": 80,
    "clarity": 70,
    "feasibility": 75,
    "security": 65,
    "performance": 70,
    "scalability": 75
  },
  "strengths": [
    "List of positive aspects of the design"
  ],
  "concerns": [
    {
      "category": "security|performance|scalability|complexity|clarity|missing",
      "severity": "critical|high|medium|low",
      "description": "Description of the concern",
      "suggestion": "How to address it",
      "section": "Which section of the document"
    }
  ],
  "suggestions": [
    "General improvement suggestions"
  ]%s
}

Focus on:
1. Design completeness - Are all necessary aspects covered?
2. Technical feasibility - Is the proposed solution realistic?
3. Security implications - Are there security risks?
4. Performance considerations - Will this scale?
5. Implementation clarity - Is this actionable?
6. Edge cases - What scenarios might be missed?
7. Dependencies - Are external dependencies well understood?

Document content:
---
%s
---

Provide your review as valid JSON only. No other text.`, docType, docPath, focusInstructions, getChecklistInstruction(), content)

	return prompt
}

func detectDocumentType(path string) string {
	base := strings.ToLower(filepath.Base(path))

	if strings.Contains(base, "rfc") {
		return "RFC (Request for Comments)"
	}
	if strings.Contains(base, "design") {
		return "Design Document"
	}
	if strings.Contains(base, "spec") {
		return "Specification"
	}
	if strings.Contains(base, "proposal") {
		return "Technical Proposal"
	}
	if strings.Contains(base, "adr") {
		return "Architecture Decision Record"
	}

	ext := filepath.Ext(path)
	switch ext {
	case ".md", ".markdown":
		return "Technical Document (Markdown)"
	case ".rst":
		return "Technical Document (reStructuredText)"
	default:
		return "Technical Document"
	}
}

func getFocusInstructions() string {
	if planFocus == "" || planFocus == "all" {
		return "Review all aspects of the design comprehensively."
	}

	focusMap := map[string]string{
		"security":        "Focus especially on security implications, vulnerabilities, and data protection.",
		"performance":     "Focus especially on performance implications, bottlenecks, and optimization opportunities.",
		"scalability":     "Focus especially on scalability concerns, growth patterns, and distributed systems considerations.",
		"maintainability": "Focus especially on code maintainability, modularity, and long-term sustainability.",
	}

	if instruction, ok := focusMap[planFocus]; ok {
		return instruction
	}
	return "Review all aspects of the design comprehensively."
}

func getChecklistInstruction() string {
	if !planChecklist {
		return ""
	}
	return `,
  "checklist": [
    {
      "task": "Implementation task description",
      "priority": "high|medium|low",
      "category": "setup|core|testing|deployment|documentation",
      "depends": ["list of prerequisite tasks"]
    }
  ]`
}

func parsePlanResponse(response string) (*PlanReview, error) {
	// Try to extract JSON from the response
	response = strings.TrimSpace(response)

	// Find JSON object in response
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")
	if start == -1 || end == -1 || end <= start {
		return nil, fmt.Errorf("no JSON object found in response")
	}

	jsonStr := response[start : end+1]

	var review PlanReview
	if err := json.Unmarshal([]byte(jsonStr), &review); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}

	return &review, nil
}

func formatPlanOutput(reviews []*PlanReview) (string, error) {
	switch planFormat {
	case "json":
		data, err := json.MarshalIndent(reviews, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil
	default:
		return formatPlanMarkdown(reviews), nil
	}
}

func formatPlanMarkdown(reviews []*PlanReview) string {
	var sb strings.Builder

	for i, review := range reviews {
		if i > 0 {
			sb.WriteString("\n---\n\n")
		}

		sb.WriteString(fmt.Sprintf("# Design Review: %s\n\n", filepath.Base(review.Document)))
		sb.WriteString(fmt.Sprintf("**Reviewed:** %s\n\n", review.ReviewedAt.Format(time.RFC3339)))

		// Summary
		sb.WriteString("## Summary\n\n")
		sb.WriteString(review.Summary)
		sb.WriteString("\n\n")

		// Score
		sb.WriteString("## Scores\n\n")
		sb.WriteString(fmt.Sprintf("| Aspect | Score |\n"))
		sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
		sb.WriteString(fmt.Sprintf("| **Overall** | %.0f/100 |\n", review.Score.Overall))
		sb.WriteString(fmt.Sprintf("| Completeness | %.0f/100 |\n", review.Score.Completeness))
		sb.WriteString(fmt.Sprintf("| Clarity | %.0f/100 |\n", review.Score.Clarity))
		sb.WriteString(fmt.Sprintf("| Feasibility | %.0f/100 |\n", review.Score.Feasibility))
		if review.Score.Security > 0 {
			sb.WriteString(fmt.Sprintf("| Security | %.0f/100 |\n", review.Score.Security))
		}
		if review.Score.Performance > 0 {
			sb.WriteString(fmt.Sprintf("| Performance | %.0f/100 |\n", review.Score.Performance))
		}
		if review.Score.Scalability > 0 {
			sb.WriteString(fmt.Sprintf("| Scalability | %.0f/100 |\n", review.Score.Scalability))
		}
		sb.WriteString("\n")

		// Strengths
		if len(review.Strengths) > 0 {
			sb.WriteString("## Strengths\n\n")
			for _, s := range review.Strengths {
				sb.WriteString(fmt.Sprintf("- %s\n", s))
			}
			sb.WriteString("\n")
		}

		// Concerns
		if len(review.Concerns) > 0 {
			sb.WriteString("## Concerns\n\n")
			for _, c := range review.Concerns {
				emoji := getConcernEmoji(c.Severity)
				sb.WriteString(fmt.Sprintf("### %s [%s] %s\n\n", emoji, c.Category, c.Description))
				if c.Section != "" {
					sb.WriteString(fmt.Sprintf("**Section:** %s\n\n", c.Section))
				}
				if c.Suggestion != "" {
					sb.WriteString(fmt.Sprintf("**Suggestion:** %s\n\n", c.Suggestion))
				}
			}
		}

		// Suggestions
		if len(review.Suggestions) > 0 {
			sb.WriteString("## Improvement Suggestions\n\n")
			for _, s := range review.Suggestions {
				sb.WriteString(fmt.Sprintf("- %s\n", s))
			}
			sb.WriteString("\n")
		}

		// Checklist
		if len(review.Checklist) > 0 {
			sb.WriteString("## Implementation Checklist\n\n")

			// Group by category
			categories := []string{"setup", "core", "testing", "deployment", "documentation"}
			for _, cat := range categories {
				var items []ChecklistItem
				for _, item := range review.Checklist {
					if item.Category == cat {
						items = append(items, item)
					}
				}
				if len(items) > 0 {
					sb.WriteString(fmt.Sprintf("### %s\n\n", titleCase(cat)))
					for _, item := range items {
						priority := getPriorityEmoji(item.Priority)
						sb.WriteString(fmt.Sprintf("- [ ] %s %s\n", priority, item.Task))
						if len(item.Depends) > 0 {
							sb.WriteString(fmt.Sprintf("  - *Depends on:* %s\n", strings.Join(item.Depends, ", ")))
						}
					}
					sb.WriteString("\n")
				}
			}
		}
	}

	return sb.String()
}

func getConcernEmoji(severity string) string {
	switch severity {
	case "critical":
		return "!!"
	case "high":
		return "!"
	case "medium":
		return "~"
	default:
		return "-"
	}
}

func getPriorityEmoji(priority string) string {
	switch priority {
	case "high":
		return "[P0]"
	case "medium":
		return "[P1]"
	default:
		return "[P2]"
	}
}

// titleCase capitalizes the first letter of a string.
func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
