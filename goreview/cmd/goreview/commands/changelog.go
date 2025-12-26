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

func runChangelog(cmd *cobra.Command, _ []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	gitRepo, err := git.NewRepo(".")
	if err != nil {
		return fmt.Errorf("initializing git: %w", err)
	}

	flags := parseChangelogFlags(cmd)
	from, version, err := resolveChangelogRange(ctx, gitRepo, flags)
	if err != nil {
		return err
	}

	commits, err := gitRepo.GetCommits(ctx, from, flags.to)
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

	grouped := groupCommitsByType(commits)
	opts := changelogOptions{
		Version:  version,
		NoHeader: flags.noHeader,
		NoDate:   flags.noDate,
		NoLinks:  flags.noLinks,
	}
	changelog := generateChangelog(grouped, opts)

	if flags.output != "" {
		return writeChangelog(flags.output, changelog, flags.appendFile)
	}

	fmt.Print(changelog)
	return nil
}

type changelogFlags struct {
	from       string
	to         string
	unreleased bool
	output     string
	appendFile bool
	version    string
	noHeader   bool
	noDate     bool
	noLinks    bool
}

func parseChangelogFlags(cmd *cobra.Command) changelogFlags {
	from, _ := cmd.Flags().GetString("from")
	to, _ := cmd.Flags().GetString("to")
	unreleased, _ := cmd.Flags().GetBool("unreleased")
	output, _ := cmd.Flags().GetString("output")
	appendFile, _ := cmd.Flags().GetBool("append")
	version, _ := cmd.Flags().GetString("version")
	noHeader, _ := cmd.Flags().GetBool("no-header")
	noDate, _ := cmd.Flags().GetBool("no-date")
	noLinks, _ := cmd.Flags().GetBool("no-links")

	return changelogFlags{
		from:       from,
		to:         to,
		unreleased: unreleased,
		output:     output,
		appendFile: appendFile,
		version:    version,
		noHeader:   noHeader,
		noDate:     noDate,
		noLinks:    noLinks,
	}
}

func resolveChangelogRange(ctx context.Context, gitRepo *git.Repo, flags changelogFlags) (from, version string, err error) {
	from = flags.from
	version = flags.version

	if flags.unreleased {
		latestTag, tagErr := gitRepo.GetLatestTag(ctx)
		if tagErr != nil {
			return "", "", fmt.Errorf("getting latest tag: %w", tagErr)
		}
		if latestTag != nil {
			from = latestTag.Name
		}
		if version == "" {
			version = "Unreleased"
		}
		return from, version, nil
	}

	if from == "" {
		latestTag, tagErr := gitRepo.GetLatestTag(ctx)
		if tagErr != nil {
			return "", "", fmt.Errorf("getting latest tag: %w", tagErr)
		}
		if latestTag != nil {
			from = latestTag.Name
		}
	}

	return from, version, nil
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

	writeChangelogHeader(&sb, opts)
	writeBreakingChangesSection(&sb, grouped, opts.NoLinks)
	writeTypeGroupSections(&sb, grouped, opts.NoLinks)
	writeOtherChangesSection(&sb, grouped, opts.NoLinks)

	return sb.String()
}

func writeChangelogHeader(sb *strings.Builder, opts changelogOptions) {
	if opts.NoHeader {
		return
	}

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

func writeBreakingChangesSection(sb *strings.Builder, grouped map[string][]git.ConventionalCommit, noLinks bool) {
	breakingChanges := collectBreakingChanges(grouped)
	if len(breakingChanges) == 0 {
		return
	}

	sb.WriteString("### BREAKING CHANGES\n\n")
	for _, cc := range breakingChanges {
		writeCommitLine(sb, cc, noLinks)
	}
	sb.WriteString("\n")
}

func writeTypeGroupSections(sb *strings.Builder, grouped map[string][]git.ConventionalCommit, noLinks bool) {
	for _, typeInfo := range commitTypeOrder {
		commits, ok := grouped[typeInfo.Type]
		if !ok || len(commits) == 0 {
			continue
		}

		nonBreaking := filterNonBreaking(commits)
		if len(nonBreaking) == 0 {
			continue
		}

		writeTypeSection(sb, typeInfo.Title, nonBreaking, noLinks)
	}
}

func writeTypeSection(sb *strings.Builder, title string, commits []git.ConventionalCommit, noLinks bool) {
	sb.WriteString("### ")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Scope < commits[j].Scope
	})

	for _, cc := range commits {
		writeCommitLine(sb, cc, noLinks)
	}
	sb.WriteString("\n")
}

func writeOtherChangesSection(sb *strings.Builder, grouped map[string][]git.ConventionalCommit, noLinks bool) {
	others, ok := grouped["other"]
	if !ok || len(others) == 0 {
		return
	}

	sb.WriteString("### Other Changes\n\n")
	for _, cc := range others {
		writeCommitLine(sb, cc, noLinks)
	}
	sb.WriteString("\n")
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
