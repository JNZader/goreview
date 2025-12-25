package commands

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/git"
)

var changelogCmd = &cobra.Command{
	Use:   "changelog",
	Short: "Generate changelog from git commits",
	Long: `Generate a changelog from git commits following Conventional Commits format.

The changelog groups commits by type (feat, fix, refactor, etc.) and can be
generated for a specific version range or from the last tag.

Examples:
  # Generate changelog from last tag to HEAD
  goreview changelog

  # Generate changelog from a specific version
  goreview changelog --from=v1.0.0

  # Generate changelog between two versions
  goreview changelog --from=v1.0.0 --to=v1.1.0

  # Generate unreleased changes only
  goreview changelog --unreleased

  # Output to file
  goreview changelog --output=CHANGELOG.md

  # Append to existing changelog
  goreview changelog --append`,
	RunE: runChangelog,
}

func init() {
	rootCmd.AddCommand(changelogCmd)

	// Version range flags
	changelogCmd.Flags().String("from", "", "Start reference (tag, commit, or branch)")
	changelogCmd.Flags().String("to", "HEAD", "End reference (default: HEAD)")
	changelogCmd.Flags().Bool("unreleased", false, "Only show unreleased changes since last tag")

	// Output flags
	changelogCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	changelogCmd.Flags().Bool("append", false, "Append to existing changelog file")
	changelogCmd.Flags().String("version", "", "Version name for the changelog header")

	// Format flags
	changelogCmd.Flags().Bool("no-header", false, "Skip the version header")
	changelogCmd.Flags().Bool("no-date", false, "Skip the date in header")
	changelogCmd.Flags().Bool("no-links", false, "Skip commit links")
}

func runChangelog(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize git repo
	gitRepo, err := git.NewRepo(".")
	if err != nil {
		return fmt.Errorf("initializing git: %w", err)
	}

	// Get flags
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	unreleased, _ := cmd.Flags().GetBool("unreleased")
	output, _ := cmd.Flags().GetString("output")
	appendFile, _ := cmd.Flags().GetBool("append")
	version, _ := cmd.Flags().GetString("version")
	noHeader, _ := cmd.Flags().GetBool("no-header")
	noDate, _ := cmd.Flags().GetBool("no-date")
	noLinks, _ := cmd.Flags().GetBool("no-links")

	// If unreleased, get from latest tag
	if unreleased {
		latestTag, tagErr := gitRepo.GetLatestTag(ctx)
		if tagErr != nil {
			return fmt.Errorf("getting latest tag: %w", tagErr)
		}
		if latestTag != nil {
			from = latestTag.Name
		}
		if version == "" {
			version = "Unreleased"
		}
	}

	// If no from specified, try to get from latest tag
	if from == "" && !unreleased {
		latestTag, tagErr := gitRepo.GetLatestTag(ctx)
		if tagErr != nil {
			return fmt.Errorf("getting latest tag: %w", tagErr)
		}
		if latestTag != nil {
			from = latestTag.Name
		}
	}

	// Get commits
	commits, err := gitRepo.GetCommits(ctx, from, to)
	if err != nil {
		return fmt.Errorf("getting commits: %w", err)
	}

	if len(commits) == 0 {
		if !isQuiet() {
			fmt.Fprintln(os.Stderr, "No commits found in the specified range")
		}
		return nil
	}

	if isVerbose() {
		fmt.Fprintf(os.Stderr, "Found %d commits\n", len(commits))
	}

	// Parse and group commits
	grouped := groupCommitsByType(commits)

	// Generate changelog
	opts := changelogOptions{
		Version:  version,
		NoHeader: noHeader,
		NoDate:   noDate,
		NoLinks:  noLinks,
	}
	changelog := generateChangelog(grouped, opts)

	// Output
	if output != "" {
		return writeChangelog(output, changelog, appendFile)
	}

	fmt.Print(changelog)
	return nil
}

type changelogOptions struct {
	Version  string
	NoHeader bool
	NoDate   bool
	NoLinks  bool
}

type commitGroup struct {
	Type    string
	Title   string
	Commits []git.ConventionalCommit
}

// Conventional commit types and their display titles
var commitTypeOrder = []struct {
	Type  string
	Title string
}{
	{"feat", "Features"},
	{"fix", "Bug Fixes"},
	{"perf", "Performance Improvements"},
	{"refactor", "Code Refactoring"},
	{"docs", "Documentation"},
	{"test", "Tests"},
	{"build", "Build System"},
	{"ci", "CI/CD"},
	{"chore", "Chores"},
	{"style", "Styles"},
	{"revert", "Reverts"},
}

var conventionalCommitRegex = regexp.MustCompile(`^(\w+)(?:\(([^)]+)\))?(!)?:\s*(.+)$`)

func parseConventionalCommitMsg(commit git.Commit) git.ConventionalCommit {
	cc := git.ConventionalCommit{
		Hash:      commit.Hash,
		ShortHash: commit.ShortHash,
		Author:    commit.Author,
		Date:      commit.Date,
		Body:      commit.Body,
	}

	matches := conventionalCommitRegex.FindStringSubmatch(commit.Subject)
	if matches == nil {
		// Not a conventional commit, treat as "other"
		cc.Type = "other"
		cc.Description = commit.Subject
		return cc
	}

	cc.Type = strings.ToLower(matches[1])
	cc.Scope = matches[2]
	cc.Breaking = matches[3] == "!"
	cc.Description = matches[4]

	// Check for BREAKING CHANGE in body
	if strings.Contains(commit.Body, "BREAKING CHANGE") {
		cc.Breaking = true
	}

	return cc
}

func groupCommitsByType(commits []git.Commit) map[string][]git.ConventionalCommit {
	grouped := make(map[string][]git.ConventionalCommit)

	for _, commit := range commits {
		cc := parseConventionalCommitMsg(commit)
		grouped[cc.Type] = append(grouped[cc.Type], cc)
	}

	return grouped
}

func generateChangelog(grouped map[string][]git.ConventionalCommit, opts changelogOptions) string {
	var sb strings.Builder

	// Header
	if !opts.NoHeader {
		if opts.Version != "" {
			sb.WriteString("## ")
			sb.WriteString(opts.Version)
		} else {
			sb.WriteString("## Changelog")
		}

		if !opts.NoDate {
			sb.WriteString(" (")
			sb.WriteString(time.Now().Format("2006-01-02"))
			sb.WriteString(")")
		}
		sb.WriteString("\n\n")
	}

	// Breaking changes first
	breakingChanges := collectBreakingChanges(grouped)
	if len(breakingChanges) > 0 {
		sb.WriteString("### BREAKING CHANGES\n\n")
		for _, cc := range breakingChanges {
			writeCommitLine(&sb, cc, opts.NoLinks)
		}
		sb.WriteString("\n")
	}

	// Groups in order
	for _, typeInfo := range commitTypeOrder {
		commits, ok := grouped[typeInfo.Type]
		if !ok || len(commits) == 0 {
			continue
		}

		// Filter out breaking changes (already shown above)
		nonBreaking := filterNonBreaking(commits)
		if len(nonBreaking) == 0 {
			continue
		}

		sb.WriteString("### ")
		sb.WriteString(typeInfo.Title)
		sb.WriteString("\n\n")

		// Sort by scope
		sort.Slice(nonBreaking, func(i, j int) bool {
			return nonBreaking[i].Scope < nonBreaking[j].Scope
		})

		for _, cc := range nonBreaking {
			writeCommitLine(&sb, cc, opts.NoLinks)
		}
		sb.WriteString("\n")
	}

	// Other commits (non-conventional)
	if others, ok := grouped["other"]; ok && len(others) > 0 {
		sb.WriteString("### Other Changes\n\n")
		for _, cc := range others {
			writeCommitLine(&sb, cc, opts.NoLinks)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func writeCommitLine(sb *strings.Builder, cc git.ConventionalCommit, noLinks bool) {
	sb.WriteString("- ")

	if cc.Scope != "" {
		sb.WriteString("**")
		sb.WriteString(cc.Scope)
		sb.WriteString(":** ")
	}

	sb.WriteString(cc.Description)

	if !noLinks {
		sb.WriteString(" (")
		sb.WriteString(cc.ShortHash)
		sb.WriteString(")")
	}

	sb.WriteString("\n")
}

func collectBreakingChanges(grouped map[string][]git.ConventionalCommit) []git.ConventionalCommit {
	var breaking []git.ConventionalCommit
	for _, commits := range grouped {
		for _, cc := range commits {
			if cc.Breaking {
				breaking = append(breaking, cc)
			}
		}
	}
	return breaking
}

func filterNonBreaking(commits []git.ConventionalCommit) []git.ConventionalCommit {
	var result []git.ConventionalCommit
	for _, cc := range commits {
		if !cc.Breaking {
			result = append(result, cc)
		}
	}
	return result
}

func writeChangelog(filename, content string, appendToFile bool) error {
	var flag int
	if appendToFile {
		flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flag = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}

	file, err := os.OpenFile(filename, flag, 0600) //nolint:gosec // User-specified output file
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	if !isQuiet() {
		fmt.Fprintf(os.Stderr, "Changelog written to %s\n", filename)
	}

	return nil
}
