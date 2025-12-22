package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var reviewCmd = &cobra.Command{
	Use:   "review [files...]",
	Short: "Review code changes using AI",
	Long: `Review code changes using AI-powered analysis.

Examples:
  # Review staged changes
  goreview review --staged

  # Review a specific commit
  goreview review --commit abc123

  # Review changes from current branch vs main
  goreview review --branch main

  # Review specific files
  goreview review file1.go file2.go

  # Output as JSON
  goreview review --staged --format json

  # Save report to file
  goreview review --staged -o report.md`,
	RunE: runReview,
}

func init() {
	rootCmd.AddCommand(reviewCmd)

	// Mode flags (mutually exclusive)
	reviewCmd.Flags().Bool("staged", false, "Review staged changes")
	reviewCmd.Flags().String("commit", "", "Review a specific commit")
	reviewCmd.Flags().String("branch", "", "Review changes compared to branch")

	// Output flags
	reviewCmd.Flags().StringP("format", "f", "markdown", "Output format (markdown, json, sarif)")
	reviewCmd.Flags().StringP("output", "o", "", "Write report to file")

	// Filter flags
	reviewCmd.Flags().StringSlice("include", nil, "Include only these file patterns")
	reviewCmd.Flags().StringSlice("exclude", nil, "Exclude these file patterns")

	// Provider flags
	reviewCmd.Flags().String("provider", "", "AI provider to use (ollama, openai)")
	reviewCmd.Flags().String("model", "", "Model to use")

	// Behavior flags
	reviewCmd.Flags().Int("concurrency", 0, "Max concurrent file reviews (0=auto)")
	reviewCmd.Flags().Bool("no-cache", false, "Disable caching")
	reviewCmd.Flags().String("preset", "standard", "Rule preset (minimal, standard, strict)")

	// Bind to viper
	_ = viper.BindPFlag("review.staged", reviewCmd.Flags().Lookup("staged"))
	_ = viper.BindPFlag("review.commit", reviewCmd.Flags().Lookup("commit"))
	_ = viper.BindPFlag("review.branch", reviewCmd.Flags().Lookup("branch"))
	_ = viper.BindPFlag("review.format", reviewCmd.Flags().Lookup("format"))
	_ = viper.BindPFlag("review.output", reviewCmd.Flags().Lookup("output"))
	_ = viper.BindPFlag("review.concurrency", reviewCmd.Flags().Lookup("concurrency"))
}

func runReview(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateReviewFlags(cmd, args); err != nil {
		return err
	}

	// Will be implemented in next commits
	fmt.Println("Review command - not yet implemented")
	return nil
}

func validateReviewFlags(cmd *cobra.Command, args []string) error {
	staged, _ := cmd.Flags().GetBool("staged")
	commit, _ := cmd.Flags().GetString("commit")
	branch, _ := cmd.Flags().GetString("branch")

	// Count active modes
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

	// Must have exactly one mode
	if modeCount == 0 {
		return fmt.Errorf("must specify review mode: --staged, --commit, --branch, or file arguments")
	}
	if modeCount > 1 {
		return fmt.Errorf("only one review mode allowed at a time")
	}

	// Validate format
	format, _ := cmd.Flags().GetString("format")
	validFormats := map[string]bool{"markdown": true, "json": true, "sarif": true}
	if !validFormats[format] {
		return fmt.Errorf("invalid format %q, must be: markdown, json, or sarif", format)
	}

	return nil
}

func determineReviewMode(cmd *cobra.Command, args []string) (string, interface{}) {
	if staged, _ := cmd.Flags().GetBool("staged"); staged {
		return "staged", nil
	}
	if commit, _ := cmd.Flags().GetString("commit"); commit != "" {
		return "commit", commit
	}
	if branch, _ := cmd.Flags().GetString("branch"); branch != "" {
		return "branch", branch
	}
	if len(args) > 0 {
		return "files", args
	}
	return "staged", nil // Default
}
