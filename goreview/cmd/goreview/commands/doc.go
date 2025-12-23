package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/git"
	"github.com/JNZader/goreview/goreview/internal/providers"
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
	// Load configuration
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Initialize git repo
	gitRepo, err := git.NewGitRepository(".")
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
	appendMode, _ := cmd.Flags().GetBool("append")
	prependMode, _ := cmd.Flags().GetBool("prepend")

	if outputFile != "" {
		return writeDocOutput(outputFile, output, appendMode, prependMode)
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
	result := make([]string, 0, len(lines)+2)
	result = append(result, "/**")
	for _, line := range lines {
		result = append(result, " * "+line)
	}
	result = append(result, " */")
	return strings.Join(result, "\n")
}

func formatAsGoDoc(doc string) string {
	lines := strings.Split(doc, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		result = append(result, "// "+line)
	}
	return strings.Join(result, "\n")
}

func writeDocOutput(path, content string, appendMode, prependMode bool) error {
	if appendMode {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString("\n" + content)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Appended to: %s\n", path)
		return nil
	}

	if prependMode {
		existing, err := os.ReadFile(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		content = content + "\n" + string(existing)
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return err
	}

	if prependMode {
		fmt.Fprintf(os.Stderr, "Prepended to: %s\n", path)
	} else {
		fmt.Fprintf(os.Stderr, "Written to: %s\n", path)
	}
	return nil
}
