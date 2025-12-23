# Iteracion 09: Comando Review

## Objetivos

- Implementar comando `goreview review`
- Integrar todos los modulos anteriores
- Soportar multiples modos (staged, commit, branch, files)
- Seleccion de formato de salida
- Escritura a archivo

## Tiempo Estimado: 6 horas

---

## Commit 9.1: Crear estructura del comando review

**Mensaje de commit:**
```
feat(cli): add review command structure

- Add review.go command file
- Define all flags and options
- Setup command hierarchy
```

### `goreview/cmd/review.go`

```go
package cmd

import (
	"fmt"
	"os"

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
	viper.BindPFlag("review.staged", reviewCmd.Flags().Lookup("staged"))
	viper.BindPFlag("review.commit", reviewCmd.Flags().Lookup("commit"))
	viper.BindPFlag("review.branch", reviewCmd.Flags().Lookup("branch"))
	viper.BindPFlag("review.format", reviewCmd.Flags().Lookup("format"))
	viper.BindPFlag("review.output", reviewCmd.Flags().Lookup("output"))
	viper.BindPFlag("review.concurrency", reviewCmd.Flags().Lookup("concurrency"))
}

func runReview(cmd *cobra.Command, args []string) error {
	// Will be implemented in next commits
	fmt.Println("Review command - not yet implemented")
	return nil
}
```

---

## Commit 9.2: Implementar validacion de flags

**Mensaje de commit:**
```
feat(cli): add review flag validation

- Validate mutually exclusive mode flags
- Validate output format
- Check provider availability
```

### Agregar a `goreview/cmd/review.go`:

```go
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
```

---

## Commit 9.3: Integrar motor de review

**Mensaje de commit:**
```
feat(cli): integrate review engine

- Load configuration
- Initialize provider, cache, rules
- Execute review engine
- Handle errors gracefully
```

### Actualizar `goreview/cmd/review.go`:

```go
package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/cache"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/git"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/report"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/review"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/rules"
)

func runReview(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateReviewFlags(cmd, args); err != nil {
		return err
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Apply flag overrides
	applyFlagOverrides(cmd, cfg, args)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Initialize components
	gitRepo, err := git.NewRepository(".")
	if err != nil {
		return fmt.Errorf("initializing git: %w", err)
	}

	provider, err := providers.NewProvider(cfg)
	if err != nil {
		return fmt.Errorf("initializing provider: %w", err)
	}
	defer provider.Close()

	// Health check provider
	if err := provider.HealthCheck(ctx); err != nil {
		return fmt.Errorf("provider not available: %w", err)
	}

	// Initialize cache (optional)
	var reviewCache cache.Cache
	if noCache, _ := cmd.Flags().GetBool("no-cache"); !noCache {
		reviewCache = cache.NewLRUCache(100, time.Hour)
	}

	// Load rules
	rulesLoader := rules.NewLoader(cfg.Rules.CustomDir)
	allRules, err := rulesLoader.Load()
	if err != nil {
		return fmt.Errorf("loading rules: %w", err)
	}

	// Apply preset
	preset, _ := cmd.Flags().GetString("preset")
	presetConfig, err := rulesLoader.LoadPreset(preset)
	if err != nil {
		return fmt.Errorf("loading preset: %w", err)
	}
	activeRules := rules.ApplyPreset(allRules, presetConfig)

	// Create and run engine
	engine := review.NewEngine(cfg, gitRepo, provider, reviewCache, activeRules)
	result, err := engine.Run(ctx)
	if err != nil {
		return fmt.Errorf("review failed: %w", err)
	}

	// Generate report
	format, _ := cmd.Flags().GetString("format")
	reporter, err := report.NewReporter(format)
	if err != nil {
		return err
	}

	output, err := reporter.Generate(result)
	if err != nil {
		return fmt.Errorf("generating report: %w", err)
	}

	// Write output
	outputFile, _ := cmd.Flags().GetString("output")
	if outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Printf("Report written to %s\n", outputFile)
	} else {
		fmt.Print(output)
	}

	// Exit with error code if critical issues found
	if result.TotalIssues > 0 {
		for _, f := range result.Files {
			if f.Response != nil {
				for _, issue := range f.Response.Issues {
					if issue.Severity == providers.SeverityCritical {
						os.Exit(1)
					}
				}
			}
		}
	}

	return nil
}

func applyFlagOverrides(cmd *cobra.Command, cfg *config.Config, args []string) {
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
	if concurrency, _ := cmd.Flags().GetInt("concurrency"); concurrency > 0 {
		cfg.Review.MaxConcurrency = concurrency
	}

	// Include/exclude patterns
	if includes, _ := cmd.Flags().GetStringSlice("include"); len(includes) > 0 {
		cfg.Review.IncludePatterns = includes
	}
	if excludes, _ := cmd.Flags().GetStringSlice("exclude"); len(excludes) > 0 {
		cfg.Git.IgnorePatterns = append(cfg.Git.IgnorePatterns, excludes...)
	}
}
```

---

## Commit 9.4: Agregar barra de progreso

**Mensaje de commit:**
```
feat(cli): add progress indicator

- Show file processing progress
- Display spinner during LLM calls
- Print summary statistics
```

### `goreview/cmd/progress.go`

```go
package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// ProgressReporter handles CLI progress output.
type ProgressReporter struct {
	total     int
	current   int
	mu        sync.Mutex
	startTime time.Time
	spinner   int
	done      chan struct{}
}

// NewProgressReporter creates a new progress reporter.
func NewProgressReporter(total int) *ProgressReporter {
	return &ProgressReporter{
		total:     total,
		startTime: time.Now(),
		done:      make(chan struct{}),
	}
}

// Start begins the progress display.
func (p *ProgressReporter) Start() {
	go func() {
		spinnerChars := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-p.done:
				return
			case <-ticker.C:
				p.mu.Lock()
				spinner := spinnerChars[p.spinner%len(spinnerChars)]
				p.spinner++
				progress := float64(p.current) / float64(p.total) * 100
				bar := p.renderBar(progress)
				fmt.Fprintf(os.Stderr, "\r%c Reviewing files %s %.0f%% (%d/%d)",
					spinner, bar, progress, p.current, p.total)
				p.mu.Unlock()
			}
		}
	}()
}

// Increment advances the progress by one.
func (p *ProgressReporter) Increment(filename string) {
	p.mu.Lock()
	p.current++
	p.mu.Unlock()
}

// Finish completes the progress display.
func (p *ProgressReporter) Finish() {
	close(p.done)
	duration := time.Since(p.startTime)
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", 80))
	fmt.Fprintf(os.Stderr, "Reviewed %d files in %s\n", p.total, duration.Round(time.Millisecond))
}

func (p *ProgressReporter) renderBar(progress float64) string {
	width := 20
	filled := int(progress / 100 * float64(width))
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

// PrintSummary prints a summary of the review results.
func PrintSummary(totalIssues int, files int, duration time.Duration) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "  Review Complete\n")
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Fprintf(os.Stderr, "  Files reviewed: %d\n", files)
	fmt.Fprintf(os.Stderr, "  Issues found:   %d\n", totalIssues)
	fmt.Fprintf(os.Stderr, "  Duration:       %s\n", duration.Round(time.Millisecond))
	fmt.Fprintf(os.Stderr, "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}
```

---

## Commit 9.5: Agregar soporte para archivo de salida

**Mensaje de commit:**
```
feat(cli): add file output support

- Write report to specified file
- Auto-detect format from extension
- Create parent directories if needed
```

### Agregar funciones de output:

```go
// En goreview/cmd/output.go

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteOutput writes the report to a file or stdout.
func WriteOutput(content, outputPath string) error {
	if outputPath == "" {
		fmt.Print(content)
		return nil
	}

	// Create parent directories
	dir := filepath.Dir(outputPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Write file
	if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Report written to: %s\n", outputPath)
	return nil
}

// DetectFormatFromPath infers the output format from file extension.
func DetectFormatFromPath(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "json"
	case ".sarif":
		return "sarif"
	case ".md", ".markdown":
		return "markdown"
	default:
		return ""
	}
}
```

---

## Commit 9.6: Tests de integracion del comando review

**Mensaje de commit:**
```
test(cli): add review command integration tests

- Test flag validation
- Test mode detection
- Test output generation
- Mock provider for testing
```

### `goreview/cmd/review_test.go`

```go
package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
)

func TestValidateReviewFlags(t *testing.T) {
	tests := []struct {
		name    string
		flags   map[string]interface{}
		args    []string
		wantErr bool
	}{
		{
			name:    "no mode specified",
			flags:   map[string]interface{}{},
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "staged mode",
			flags:   map[string]interface{}{"staged": true},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "commit mode",
			flags:   map[string]interface{}{"commit": "abc123"},
			args:    []string{},
			wantErr: false,
		},
		{
			name:    "files mode",
			flags:   map[string]interface{}{},
			args:    []string{"file.go"},
			wantErr: false,
		},
		{
			name:    "multiple modes",
			flags:   map[string]interface{}{"staged": true, "commit": "abc123"},
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "invalid format",
			flags:   map[string]interface{}{"staged": true, "format": "xml"},
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("staged", false, "")
			cmd.Flags().String("commit", "", "")
			cmd.Flags().String("branch", "", "")
			cmd.Flags().String("format", "markdown", "")

			for k, v := range tt.flags {
				switch val := v.(type) {
				case bool:
					cmd.Flags().Set(k, "true")
				case string:
					cmd.Flags().Set(k, val)
				}
			}

			err := validateReviewFlags(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateReviewFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetermineReviewMode(t *testing.T) {
	tests := []struct {
		name     string
		flags    map[string]interface{}
		args     []string
		wantMode string
	}{
		{
			name:     "staged",
			flags:    map[string]interface{}{"staged": true},
			wantMode: "staged",
		},
		{
			name:     "commit",
			flags:    map[string]interface{}{"commit": "abc123"},
			wantMode: "commit",
		},
		{
			name:     "branch",
			flags:    map[string]interface{}{"branch": "main"},
			wantMode: "branch",
		},
		{
			name:     "files",
			args:     []string{"file1.go", "file2.go"},
			wantMode: "files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			cmd.Flags().Bool("staged", false, "")
			cmd.Flags().String("commit", "", "")
			cmd.Flags().String("branch", "", "")

			for k, v := range tt.flags {
				switch val := v.(type) {
				case bool:
					cmd.Flags().Set(k, "true")
				case string:
					cmd.Flags().Set(k, val)
				}
			}

			mode, _ := determineReviewMode(cmd, tt.args)
			if mode != tt.wantMode {
				t.Errorf("determineReviewMode() = %v, want %v", mode, tt.wantMode)
			}
		})
	}
}

func TestDetectFormatFromPath(t *testing.T) {
	tests := []struct {
		path   string
		format string
	}{
		{"report.json", "json"},
		{"report.sarif", "sarif"},
		{"report.md", "markdown"},
		{"report.markdown", "markdown"},
		{"report.txt", ""},
		{"report", ""},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := DetectFormatFromPath(tt.path)
			if got != tt.format {
				t.Errorf("DetectFormatFromPath(%q) = %q, want %q", tt.path, got, tt.format)
			}
		})
	}
}
```

---

## Resumen de la Iteracion 09

### Commits:
1. `feat(cli): add review command structure`
2. `feat(cli): add review flag validation`
3. `feat(cli): integrate review engine`
4. `feat(cli): add progress indicator`
5. `feat(cli): add file output support`
6. `test(cli): add review command integration tests`

### Archivos:
```
goreview/cmd/
├── review.go
├── progress.go
├── output.go
└── review_test.go
```

---

## Siguiente Iteracion

Continua con: **[10-COMANDO-COMMIT.md](10-COMANDO-COMMIT.md)**
