package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/git"
	"github.com/JNZader/goreview/goreview/internal/providers"
	"github.com/JNZader/goreview/goreview/internal/review"
	"github.com/JNZader/goreview/goreview/internal/rules"
)

var fixCmd = &cobra.Command{
	Use:   "fix [files...]",
	Short: "Review and automatically fix code issues",
	Long: `Review code changes and automatically apply fixes for issues that have suggestions.

Examples:
  # Fix issues in staged changes (interactive mode)
  goreview fix --staged

  # Auto-fix all issues without confirmation
  goreview fix --staged --auto

  # Fix only specific issue types
  goreview fix --staged --types bug,security

  # Dry-run: show what would be fixed without applying
  goreview fix --staged --dry-run

  # Fix specific files
  goreview fix file1.go file2.go`,
	RunE: runFix,
}

func init() {
	rootCmd.AddCommand(fixCmd)

	// Mode flags (same as review)
	fixCmd.Flags().Bool("staged", false, "Fix issues in staged changes")
	fixCmd.Flags().String("commit", "", "Fix issues in a specific commit")
	fixCmd.Flags().String("branch", "", "Fix issues compared to branch")

	// Fix-specific flags
	fixCmd.Flags().Bool("auto", false, "Auto-apply all fixes without confirmation")
	fixCmd.Flags().Bool("dry-run", false, "Show what would be fixed without applying")
	fixCmd.Flags().StringSlice("types", nil, "Fix only these issue types (bug, security, performance, style)")
	fixCmd.Flags().StringSlice("severity", nil, "Fix only issues with these severities (info, warning, error, critical)")

	// Provider flags
	fixCmd.Flags().String("provider", "", "AI provider to use (ollama, openai)")
	fixCmd.Flags().String("model", "", "Model to use")
}

// FixableIssue represents an issue that can be fixed
type FixableIssue struct {
	FilePath  string
	Issue     providers.Issue
	FixedCode string
	StartLine int
	EndLine   int
}

func runFix(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateFixFlags(cmd, args); err != nil {
		return err
	}

	// Load configuration
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	applyFixFlagOverrides(cmd, cfg, args)

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Run review first
	fmt.Println("Analyzing code for fixable issues...")
	result, err := executeFixReview(ctx, cfg)
	if err != nil {
		return err
	}

	// Collect fixable issues
	fixableIssues := collectFixableIssues(cmd, result)
	if len(fixableIssues) == 0 {
		fmt.Println("No fixable issues found.")
		return nil
	}

	// Check for dry-run
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		return showDryRun(fixableIssues)
	}

	// Apply fixes
	autoFix, _ := cmd.Flags().GetBool("auto")
	return applyFixes(fixableIssues, autoFix)
}

func validateFixFlags(cmd *cobra.Command, args []string) error {
	staged, _ := cmd.Flags().GetBool("staged")
	commit, _ := cmd.Flags().GetString("commit")
	branch, _ := cmd.Flags().GetString("branch")

	modeCount := 0
	if staged {
		modeCount++
	}
	if commit != "" {
		modeCount++
	}
	if branch != "" {
		modeCount++
	}
	if len(args) > 0 {
		modeCount++
	}

	if modeCount == 0 {
		return fmt.Errorf("must specify mode: --staged, --commit, --branch, or file arguments")
	}
	if modeCount > 1 {
		return fmt.Errorf("only one mode allowed at a time")
	}

	return nil
}

func applyFixFlagOverrides(cmd *cobra.Command, cfg *config.Config, args []string) {
	mode, value := determineReviewMode(cmd, args)
	cfg.Review.Mode = mode

	switch mode {
	case "commit":
		cfg.Review.Commit = value.(string)
	case "branch":
		cfg.Git.BaseBranch = value.(string)
	case "files":
		cfg.Review.Files = value.([]string)
	}

	if provider, _ := cmd.Flags().GetString("provider"); provider != "" {
		cfg.Provider.Name = provider
	}
	if model, _ := cmd.Flags().GetString("model"); model != "" {
		cfg.Provider.Model = model
	}
}

func executeFixReview(ctx context.Context, cfg *config.Config) (*review.Result, error) {
	gitRepo, err := git.NewRepo(".")
	if err != nil {
		return nil, fmt.Errorf("initializing git: %w", err)
	}

	provider, err := providers.NewProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("initializing provider: %w", err)
	}
	defer func() { _ = provider.Close() }()

	if healthErr := provider.HealthCheck(ctx); healthErr != nil {
		return nil, fmt.Errorf("provider not available: %w", healthErr)
	}

	rulesLoader := rules.NewLoader(cfg.Rules.RulesDir)
	allRules, err := rulesLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("loading rules: %w", err)
	}

	presetConfig, err := rulesLoader.LoadPreset("standard")
	if err != nil {
		return nil, fmt.Errorf("loading preset: %w", err)
	}
	activeRules := rules.ApplyPreset(allRules, presetConfig)

	engine := review.NewEngine(cfg, gitRepo, provider, nil, activeRules)
	result, err := engine.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("review failed: %w", err)
	}
	return result, nil
}

func collectFixableIssues(cmd *cobra.Command, result *review.Result) []FixableIssue {
	typeFilter, _ := cmd.Flags().GetStringSlice("types")
	severityFilter, _ := cmd.Flags().GetStringSlice("severity")

	typeSet := make(map[string]bool)
	for _, t := range typeFilter {
		typeSet[strings.ToLower(t)] = true
	}

	severitySet := make(map[string]bool)
	for _, s := range severityFilter {
		severitySet[strings.ToLower(s)] = true
	}

	var fixable []FixableIssue

	for _, fileResult := range result.Files {
		if fileResult.Response == nil {
			continue
		}

		for _, issue := range fileResult.Response.Issues {
			// Skip if no fix available
			if issue.FixedCode == "" && issue.Suggestion == "" {
				continue
			}

			// Apply type filter
			if len(typeSet) > 0 && !typeSet[strings.ToLower(string(issue.Type))] {
				continue
			}

			// Apply severity filter
			if len(severitySet) > 0 && !severitySet[strings.ToLower(string(issue.Severity))] {
				continue
			}

			startLine := 0
			endLine := 0
			if issue.Location != nil {
				startLine = issue.Location.StartLine
				endLine = issue.Location.EndLine
			}

			fixedCode := issue.FixedCode
			if fixedCode == "" {
				// Use suggestion as a hint (may need manual review)
				fixedCode = "// TODO: " + issue.Suggestion
			}

			fixable = append(fixable, FixableIssue{
				FilePath:  fileResult.File,
				Issue:     issue,
				FixedCode: fixedCode,
				StartLine: startLine,
				EndLine:   endLine,
			})
		}
	}

	return fixable
}

func showDryRun(issues []FixableIssue) error {
	fmt.Printf("\nFound %d fixable issues:\n\n", len(issues))

	for i, fix := range issues {
		fmt.Printf("%d. [%s] %s\n", i+1, fix.Issue.Severity, fix.Issue.Message)
		fmt.Printf("   File: %s", fix.FilePath)
		if fix.StartLine > 0 {
			fmt.Printf(" (lines %d-%d)", fix.StartLine, fix.EndLine)
		}
		fmt.Println()
		if fix.Issue.Suggestion != "" {
			fmt.Printf("   Suggestion: %s\n", fix.Issue.Suggestion)
		}
		if fix.Issue.FixedCode != "" {
			fmt.Printf("   Fix available: Yes\n")
		}
		fmt.Println()
	}

	fmt.Println("Run without --dry-run to apply fixes.")
	return nil
}

func applyFixes(issues []FixableIssue, autoFix bool) error {
	applied := 0
	skipped := 0

	reader := bufio.NewReader(os.Stdin)

	for _, fix := range issues {
		fmt.Printf("\n[%s] %s\n", fix.Issue.Severity, fix.Issue.Message)
		fmt.Printf("File: %s", fix.FilePath)
		if fix.StartLine > 0 {
			fmt.Printf(" (lines %d-%d)", fix.StartLine, fix.EndLine)
		}
		fmt.Println()

		if fix.Issue.Suggestion != "" {
			fmt.Printf("Suggestion: %s\n", fix.Issue.Suggestion)
		}

		if fix.Issue.FixedCode != "" {
			fmt.Println("Proposed fix:")
			fmt.Println(strings.Repeat("-", 40))
			// Show first few lines of fix
			lines := strings.Split(fix.Issue.FixedCode, "\n")
			maxLines := 10
			if len(lines) <= maxLines {
				fmt.Println(fix.Issue.FixedCode)
			} else {
				for i := 0; i < maxLines; i++ {
					fmt.Println(lines[i])
				}
				fmt.Printf("... (%d more lines)\n", len(lines)-maxLines)
			}
			fmt.Println(strings.Repeat("-", 40))
		}

		shouldApply := autoFix
		if !autoFix {
			fmt.Print("Apply this fix? [y/n/q] ")
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(strings.ToLower(input))

			switch input {
			case "y", "yes":
				shouldApply = true
			case "q", "quit":
				fmt.Printf("\nApplied %d fixes, skipped %d\n", applied, skipped)
				return nil
			default:
				shouldApply = false
			}
		}

		if shouldApply && fix.Issue.FixedCode != "" && fix.StartLine > 0 {
			if err := applyFixToFile(fix); err != nil {
				fmt.Printf("Error applying fix: %v\n", err)
				skipped++
			} else {
				fmt.Println("Fix applied!")
				applied++
			}
		} else if shouldApply {
			fmt.Println("Cannot auto-apply: no line information or fixed code")
			skipped++
		} else {
			skipped++
		}
	}

	fmt.Printf("\nSummary: Applied %d fixes, skipped %d\n", applied, skipped)
	return nil
}

func applyFixToFile(fix FixableIssue) error {
	// Read the file
	absPath, err := filepath.Abs(fix.FilePath)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")

	// Validate line numbers
	if fix.StartLine < 1 || fix.StartLine > len(lines) {
		return fmt.Errorf("invalid start line %d", fix.StartLine)
	}
	if fix.EndLine < fix.StartLine || fix.EndLine > len(lines) {
		fix.EndLine = fix.StartLine
	}

	// Replace the lines
	fixedLines := strings.Split(fix.FixedCode, "\n")

	newLines := make([]string, 0, len(lines)-fix.EndLine+fix.StartLine-1+len(fixedLines))
	newLines = append(newLines, lines[:fix.StartLine-1]...)
	newLines = append(newLines, fixedLines...)
	newLines = append(newLines, lines[fix.EndLine:]...)

	// Write back
	newContent := strings.Join(newLines, "\n")
	return os.WriteFile(absPath, []byte(newContent), 0644)
}
