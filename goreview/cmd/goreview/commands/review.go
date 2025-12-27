package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/JNZader/goreview/goreview/internal/cache"
	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/export"
	"github.com/JNZader/goreview/goreview/internal/git"
	"github.com/JNZader/goreview/goreview/internal/profiler"
	"github.com/JNZader/goreview/goreview/internal/providers"
	"github.com/JNZader/goreview/goreview/internal/report"
	"github.com/JNZader/goreview/goreview/internal/review"
	"github.com/JNZader/goreview/goreview/internal/rules"
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
	reviewCmd.Flags().String("personality", "default", "Reviewer personality (default, senior, strict, friendly, security-expert)")
	reviewCmd.Flags().String("mode", "default", "Review focus mode (default, security, perf, clean, docs, tests). Combine with commas: security,perf")

	// TDD workflow flags
	reviewCmd.Flags().Bool("require-tests", false, "Fail if reviewed code lacks corresponding tests")
	reviewCmd.Flags().Float64("min-coverage", 0, "Minimum test coverage percentage required (0=disabled)")

	// Analysis flags
	reviewCmd.Flags().Bool("trace", false, "Enable root cause tracing for each issue")

	// Profiling flags
	reviewCmd.Flags().String("cpuprofile", "", "Write CPU profile to file")
	reviewCmd.Flags().String("memprofile", "", "Write memory profile to file")
	reviewCmd.Flags().String("pprof-addr", "", "Enable pprof HTTP server (e.g., :6060)")

	// Export flags
	reviewCmd.Flags().Bool("export-obsidian", false, "Export results to Obsidian vault")
	reviewCmd.Flags().String("obsidian-vault", "", "Override Obsidian vault path")

	// Bind to viper
	_ = viper.BindPFlag("review.staged", reviewCmd.Flags().Lookup("staged"))
	_ = viper.BindPFlag("review.commit", reviewCmd.Flags().Lookup("commit"))
	_ = viper.BindPFlag("review.branch", reviewCmd.Flags().Lookup("branch"))
	_ = viper.BindPFlag("review.format", reviewCmd.Flags().Lookup("format"))
	_ = viper.BindPFlag("review.output", reviewCmd.Flags().Lookup("output"))
	_ = viper.BindPFlag("review.concurrency", reviewCmd.Flags().Lookup("concurrency"))
}

func runReview(cmd *cobra.Command, args []string) error {
	if err := validateReviewFlags(cmd, args); err != nil {
		return err
	}

	// Initialize profiler if requested
	cleanupProfiler, err := setupProfiler(cmd)
	if err != nil {
		return err
	}
	if cleanupProfiler != nil {
		defer cleanupProfiler()
	}

	// Load configuration
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	applyFlagOverrides(cmd, cfg, args)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Initialize dependencies
	result, err := executeReview(ctx, cmd, cfg)
	if err != nil {
		return err
	}

	// Check TDD requirements
	requireTests, _ := cmd.Flags().GetBool("require-tests")
	if requireTests {
		if err := checkTestCoverage(result); err != nil {
			return err
		}
	}

	// Generate and write report
	if err := outputReport(cmd, result); err != nil {
		return err
	}

	// Export to Obsidian if requested
	exportObsidian, _ := cmd.Flags().GetBool("export-obsidian")
	if exportObsidian || cfg.Export.Obsidian.Enabled {
		if err := exportToObsidian(ctx, cmd, cfg, result); err != nil {
			// Non-fatal - log warning but don't fail
			fmt.Fprintf(os.Stderr, "Warning: Obsidian export failed: %v\n", err)
		}
	}

	// Exit with error code if critical issues found
	checkCriticalIssues(result)
	return nil
}

// setupProfiler initializes profiler if flags are set, returns cleanup function
func setupProfiler(cmd *cobra.Command) (func(), error) {
	cpuProfile, _ := cmd.Flags().GetString("cpuprofile")
	memProfile, _ := cmd.Flags().GetString("memprofile")
	pprofAddr, _ := cmd.Flags().GetString("pprof-addr")

	if cpuProfile == "" && memProfile == "" && pprofAddr == "" {
		return nil, nil
	}

	prof, err := profiler.New(profiler.Config{
		CPUProfile: cpuProfile,
		MemProfile: memProfile,
		HTTPAddr:   pprofAddr,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start profiler: %w", err)
	}

	if isVerbose() {
		_, _ = fmt.Fprintf(os.Stderr, "Profiler started - Initial memory stats: %s\n", profiler.Stats().String())
	}

	return func() {
		if isVerbose() {
			_, _ = fmt.Fprintf(os.Stderr, "Profiler stopping - Final memory stats: %s\n", profiler.Stats().String())
		}
		if stopErr := prof.Stop(); stopErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to stop profiler: %v\n", stopErr)
		}
	}, nil
}

// executeReview initializes dependencies and runs the review
func executeReview(ctx context.Context, cmd *cobra.Command, cfg *config.Config) (*review.Result, error) {
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

	reviewCache := initCache(cmd, cfg)
	activeRules, err := loadActiveRules(cmd, cfg)
	if err != nil {
		return nil, err
	}

	engine := review.NewEngine(cfg, gitRepo, provider, reviewCache, activeRules)
	result, err := engine.Run(ctx)
	if err != nil {
		return nil, fmt.Errorf("review failed: %w", err)
	}
	return result, nil
}

// initCache creates a cache if enabled
func initCache(cmd *cobra.Command, cfg *config.Config) cache.Cache {
	noCache, _ := cmd.Flags().GetBool("no-cache")
	if noCache || !cfg.Cache.Enabled {
		return nil
	}
	return cache.NewLRUCache(cfg.Cache.MaxEntries, cfg.Cache.TTL)
}

// loadActiveRules loads and applies rule preset
func loadActiveRules(cmd *cobra.Command, cfg *config.Config) ([]rules.Rule, error) {
	rulesLoader := rules.NewLoader(cfg.Rules.RulesDir)
	allRules, err := rulesLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("loading rules: %w", err)
	}

	preset, _ := cmd.Flags().GetString("preset")
	presetConfig, err := rulesLoader.LoadPreset(preset)
	if err != nil {
		return nil, fmt.Errorf("loading preset: %w", err)
	}
	return rules.ApplyPreset(allRules, presetConfig), nil
}

// outputReport generates and writes the review report
func outputReport(cmd *cobra.Command, result *review.Result) error {
	format, _ := cmd.Flags().GetString("format")
	reporter, err := report.NewReporter(format)
	if err != nil {
		return err
	}

	output, err := reporter.Generate(result)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	outputFile, _ := cmd.Flags().GetString("output")
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0600); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		_, _ = fmt.Fprintf(os.Stderr, "Report written to %s\n", outputFile)
	} else {
		fmt.Print(output)
	}
	return nil
}

// checkCriticalIssues exits with code 1 if critical issues found
func checkCriticalIssues(result *review.Result) {
	if result.TotalIssues == 0 {
		return
	}
	for _, f := range result.Files {
		if f.Response == nil {
			continue
		}
		for _, issue := range f.Response.Issues {
			if issue.Severity == providers.SeverityCritical {
				os.Exit(1)
			}
		}
	}
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

func applyFlagOverrides(cmd *cobra.Command, cfg *config.Config, args []string) {
	mode, value := determineReviewMode(cmd, args)
	cfg.Review.Mode = mode

	switch mode {
	case "commit":
		//nolint:errcheck // determineReviewMode returns string for commit mode
		cfg.Review.Commit = value.(string)
	case "branch":
		//nolint:errcheck // determineReviewMode returns string for branch mode
		cfg.Git.BaseBranch = value.(string)
	case "files":
		//nolint:errcheck // determineReviewMode returns []string for files mode
		cfg.Review.Files = value.([]string)
	}

	if provider, _ := cmd.Flags().GetString("provider"); provider != "" {
		cfg.Provider.Name = provider
	}
	if model, _ := cmd.Flags().GetString("model"); model != "" {
		cfg.Provider.Model = model
	}
	if concurrency, _ := cmd.Flags().GetInt("concurrency"); concurrency > 0 {
		cfg.Review.MaxConcurrency = concurrency
	}
	if personality, _ := cmd.Flags().GetString("personality"); personality != "" {
		cfg.Review.Personality = personality
	}
	if mode, _ := cmd.Flags().GetString("mode"); mode != "" {
		cfg.Review.Modes = mode
	}
	if trace, _ := cmd.Flags().GetBool("trace"); trace {
		cfg.Review.RootCauseTracing = true
	}

	// Include/exclude patterns
	if includes, _ := cmd.Flags().GetStringSlice("include"); len(includes) > 0 {
		// Store in config for later use
		_ = includes
	}
	if excludes, _ := cmd.Flags().GetStringSlice("exclude"); len(excludes) > 0 {
		cfg.Git.IgnorePatterns = append(cfg.Git.IgnorePatterns, excludes...)
	}
}

// checkTestCoverage ensures all reviewed files have corresponding tests
func checkTestCoverage(result *review.Result) error {
	var filesWithoutTests []string

	for _, f := range result.Files {
		if isTestFile(f.File) {
			continue
		}

		if !hasCorrespondingTest(f.File) {
			filesWithoutTests = append(filesWithoutTests, f.File)
		}
	}

	if len(filesWithoutTests) > 0 {
		fmt.Fprintln(os.Stderr, "\n❌ TDD Violation: The following files lack corresponding tests:")
		for _, f := range filesWithoutTests {
			expectedTest := getExpectedTestPath(f)
			fmt.Fprintf(os.Stderr, "   • %s (expected: %s)\n", f, expectedTest)
		}
		fmt.Fprintln(os.Stderr)
		return fmt.Errorf("--require-tests: %d file(s) missing tests", len(filesWithoutTests))
	}

	fmt.Fprintln(os.Stderr, "\n✅ TDD Check: All reviewed files have corresponding tests")
	return nil
}

// isTestFile checks if the file is a test file
func isTestFile(path string) bool {
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Go tests
	if strings.HasSuffix(base, "_test.go") {
		return true
	}

	// JavaScript/TypeScript tests
	jsTestSuffixes := []string{".test", ".spec", "_test", "_spec"}
	for _, suffix := range jsTestSuffixes {
		if strings.HasSuffix(nameWithoutExt, suffix) {
			return true
		}
	}

	// Test directories - normalize path separators for cross-platform
	normalizedPath := filepath.ToSlash(path)
	testDirs := []string{"test", "tests", "__tests__", "spec", "specs"}
	for _, d := range testDirs {
		// Check if directory is in path
		if strings.Contains(normalizedPath, "/"+d+"/") ||
			strings.HasPrefix(normalizedPath, d+"/") ||
			strings.Contains(normalizedPath, d+"/") {
			return true
		}
	}

	return false
}

// hasCorrespondingTest checks if a source file has a corresponding test file
func hasCorrespondingTest(path string) bool {
	testPaths := getTestPathVariants(path)

	for _, testPath := range testPaths {
		if _, err := os.Stat(testPath); err == nil {
			return true
		}
	}

	return false
}

// getTestPathVariants returns possible test file paths for a source file
func getTestPathVariants(path string) []string {
	dir := filepath.Dir(path)
	base := filepath.Base(path)
	ext := filepath.Ext(path)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	var variants []string

	switch ext {
	case ".go":
		// Go: file.go -> file_test.go
		variants = append(variants, filepath.Join(dir, nameWithoutExt+"_test.go"))

	case ".js", ".jsx", ".ts", ".tsx":
		// JS/TS: file.js -> file.test.js, file.spec.js
		variants = append(variants, filepath.Join(dir, nameWithoutExt+".test"+ext))
		variants = append(variants, filepath.Join(dir, nameWithoutExt+".spec"+ext))
		variants = append(variants, filepath.Join(dir, "__tests__", base))
		// Also check for .ts tests if .js file
		if ext == ".js" || ext == ".jsx" {
			tsExt := strings.Replace(ext, ".js", ".ts", 1)
			variants = append(variants, filepath.Join(dir, nameWithoutExt+".test"+tsExt))
			variants = append(variants, filepath.Join(dir, nameWithoutExt+".spec"+tsExt))
		}

	case ".py":
		// Python: file.py -> test_file.py, file_test.py
		variants = append(variants, filepath.Join(dir, "test_"+base))
		variants = append(variants, filepath.Join(dir, nameWithoutExt+"_test.py"))
		variants = append(variants, filepath.Join(dir, "tests", "test_"+base))

	case ".java":
		// Java: File.java -> FileTest.java
		variants = append(variants, filepath.Join(dir, nameWithoutExt+"Test.java"))
		variants = append(variants, strings.Replace(path, "/main/", "/test/", 1))

	case ".rs":
		// Rust: usually in same file or mod tests
		variants = append(variants, filepath.Join(dir, nameWithoutExt+"_test.rs"))

	case ".rb":
		// Ruby: file.rb -> file_spec.rb, file_test.rb
		variants = append(variants, filepath.Join(dir, nameWithoutExt+"_spec.rb"))
		variants = append(variants, filepath.Join(dir, nameWithoutExt+"_test.rb"))
		variants = append(variants, filepath.Join(dir, "spec", base))

	default:
		// Generic: try _test suffix
		variants = append(variants, filepath.Join(dir, nameWithoutExt+"_test"+ext))
	}

	return variants
}

// getExpectedTestPath returns the most likely expected test path
func getExpectedTestPath(path string) string {
	variants := getTestPathVariants(path)
	if len(variants) > 0 {
		return variants[0]
	}
	return path + " (test file)"
}

// exportToObsidian exports the review result to an Obsidian vault
func exportToObsidian(ctx context.Context, cmd *cobra.Command, cfg *config.Config, result *review.Result) error {
	// Override vault path from flag if provided
	if vaultPath, _ := cmd.Flags().GetString("obsidian-vault"); vaultPath != "" {
		cfg.Export.Obsidian.VaultPath = vaultPath
	}

	if cfg.Export.Obsidian.VaultPath == "" {
		return fmt.Errorf("obsidian vault path not configured (use --obsidian-vault or config)")
	}

	// Create exporter
	exporter, err := export.NewObsidianExporter(&cfg.Export.Obsidian)
	if err != nil {
		return err
	}

	// Build metadata
	metadata, err := buildExportMetadata(ctx, cfg)
	if err != nil {
		return err
	}

	// Export
	if err := exporter.Export(result, metadata); err != nil {
		return err
	}

	outputPath := exporter.GetOutputPath(metadata)
	fmt.Fprintf(os.Stderr, "Exported to Obsidian: %s\n", outputPath)

	return nil
}

// buildExportMetadata builds metadata for the export
func buildExportMetadata(ctx context.Context, cfg *config.Config) (*export.ExportMetadata, error) {
	gitRepo, err := git.NewRepo(".")
	if err != nil {
		// Not in a git repo - use defaults
		cwd, _ := os.Getwd()
		return &export.ExportMetadata{
			ProjectName: filepath.Base(cwd),
			Branch:      "unknown",
			ReviewDate:  time.Now(),
			ReviewMode:  cfg.Review.Mode,
		}, nil
	}

	branch, _ := gitRepo.GetCurrentBranch(ctx)
	repoRoot, _ := gitRepo.GetRepoRoot(ctx)
	projectName := filepath.Base(repoRoot)

	// Get current commit info using git log
	commitHash := getGitCommitHash()
	commitShort := ""
	if len(commitHash) >= 7 {
		commitShort = commitHash[:7]
	}

	author := getGitAuthor()

	return &export.ExportMetadata{
		ProjectName: projectName,
		Branch:      branch,
		CommitHash:  commitHash,
		CommitShort: commitShort,
		Author:      author,
		ReviewDate:  time.Now(),
		ReviewMode:  cfg.Review.Mode,
		BaseBranch:  cfg.Git.BaseBranch,
	}, nil
}

// getGitCommitHash returns the current HEAD commit hash
func getGitCommitHash() string {
	out, err := runGitCommand("rev-parse", "HEAD")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// getGitAuthor returns the author of the HEAD commit
func getGitAuthor() string {
	out, err := runGitCommand("log", "-1", "--format=%an")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// runGitCommand executes a git command and returns the output
func runGitCommand(args ...string) (string, error) {
	cmd := exec.Command("git", args...) // #nosec G204 - git command with controlled args
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
