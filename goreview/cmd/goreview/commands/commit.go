package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/git"
	"github.com/JNZader/goreview/goreview/internal/providers"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate AI-powered commit messages",
	Long: `Generate commit messages using AI analysis of staged changes.

Examples:
  # Generate and show commit message
  goreview commit

  # Generate and commit directly
  goreview commit --execute

  # Generate with specific type
  goreview commit --type feat

  # Amend last commit with new message
  goreview commit --amend`,
	RunE: runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)

	// Execution flags
	commitCmd.Flags().BoolP("execute", "e", false, "Execute git commit with generated message")
	commitCmd.Flags().Bool("amend", false, "Amend the last commit")

	// Message customization
	commitCmd.Flags().StringP("type", "t", "", "Force commit type (feat, fix, docs, etc.)")
	commitCmd.Flags().StringP("scope", "s", "", "Force commit scope")
	commitCmd.Flags().Bool("breaking", false, "Mark as breaking change")
	commitCmd.Flags().StringP("body", "b", "", "Additional commit body")
	commitCmd.Flags().String("footer", "", "Commit footer (e.g., 'Closes #123')")

	// Output flags
	commitCmd.Flags().Bool("dry-run", false, "Show message without committing")
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize git repo
	gitRepo, err := git.NewRepo(".")
	if err != nil {
		return fmt.Errorf("initializing git: %w", err)
	}

	// Get staged diff
	diff, err := gitRepo.GetStagedDiff(ctx)
	if err != nil {
		return fmt.Errorf("getting staged diff: %w", err)
	}

	if len(diff.Files) == 0 {
		return fmt.Errorf("no staged changes found. Stage changes with 'git add' first")
	}

	// Initialize provider
	provider, err := providers.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("initializing provider: %w", err)
	}
	defer provider.Close()

	// Generate commit message
	if isVerbose() {
		fmt.Fprintf(os.Stderr, "Analyzing %d files...\n", len(diff.Files))
	}

	diffText := formatDiffForCommit(diff)
	message, err := provider.GenerateCommitMessage(ctx, diffText)
	if err != nil {
		return fmt.Errorf("generating commit message: %w", err)
	}

	// Apply overrides
	message = applyCommitOverrides(cmd, message)

	// Add body and footer
	body, _ := cmd.Flags().GetString("body")
	footer, _ := cmd.Flags().GetString("footer")
	if body != "" || footer != "" {
		message = buildFullMessage(message, body, footer)
	}

	// Dry run - just show message
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Println("Generated commit message:")
		fmt.Println("─────────────────────────")
		fmt.Println(message)
		fmt.Println("─────────────────────────")
		return nil
	}

	// Execute commit
	execute, _ := cmd.Flags().GetBool("execute")
	amend, _ := cmd.Flags().GetBool("amend")

	if execute || amend {
		return executeGitCommit(message, amend)
	}

	// Default: print message for user to copy
	fmt.Println(message)
	return nil
}

func formatDiffForCommit(diff *git.Diff) string {
	var sb strings.Builder
	for _, file := range diff.Files {
		sb.WriteString(fmt.Sprintf("File: %s (%s)\n", file.Path, file.Status))
		for _, hunk := range file.Hunks {
			sb.WriteString(hunk.Header + "\n")
			for _, line := range hunk.Lines {
				prefix := " "
				if line.Type == git.LineAddition {
					prefix = "+"
				} else if line.Type == git.LineDeletion {
					prefix = "-"
				}
				sb.WriteString(prefix + line.Content + "\n")
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func applyCommitOverrides(cmd *cobra.Command, message string) string {
	commitType, _ := cmd.Flags().GetString("type")
	scope, _ := cmd.Flags().GetString("scope")
	breaking, _ := cmd.Flags().GetBool("breaking")

	// Parse existing message
	parts := parseConventionalCommit(message)

	// Apply overrides
	if commitType != "" {
		parts.Type = commitType
	}
	if scope != "" {
		parts.Scope = scope
	}
	if breaking {
		parts.Breaking = true
	}

	// Rebuild message
	return parts.String()
}

type conventionalParts struct {
	Type        string
	Scope       string
	Breaking    bool
	Description string
}

func parseConventionalCommit(message string) *conventionalParts {
	parts := &conventionalParts{}

	// Simple parsing - find type(scope): description
	line := strings.Split(message, "\n")[0]

	if idx := strings.Index(line, ":"); idx > 0 {
		prefix := line[:idx]
		parts.Description = strings.TrimSpace(line[idx+1:])

		if strings.HasSuffix(prefix, "!") {
			parts.Breaking = true
			prefix = prefix[:len(prefix)-1]
		}

		if paren := strings.Index(prefix, "("); paren > 0 {
			parts.Type = prefix[:paren]
			parts.Scope = strings.Trim(prefix[paren:], "()")
		} else {
			parts.Type = prefix
		}
	} else {
		parts.Type = "chore"
		parts.Description = message
	}

	return parts
}

func (p *conventionalParts) String() string {
	var sb strings.Builder
	sb.WriteString(p.Type)
	if p.Scope != "" {
		sb.WriteString("(" + p.Scope + ")")
	}
	if p.Breaking {
		sb.WriteString("!")
	}
	sb.WriteString(": ")
	sb.WriteString(p.Description)
	return sb.String()
}

func buildFullMessage(subject, body, footer string) string {
	var parts []string
	parts = append(parts, subject)
	if body != "" {
		parts = append(parts, "", body)
	}
	if footer != "" {
		parts = append(parts, "", footer)
	}
	return strings.Join(parts, "\n")
}

func executeGitCommit(message string, amend bool) error {
	args := []string{"commit", "-m", message}
	if amend {
		args = append(args, "--amend")
	}

	gitCmd := exec.Command("git", args...)
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr
	gitCmd.Stdin = os.Stdin

	if err := gitCmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Commit created successfully\n")
	return nil
}
