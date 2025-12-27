package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/export"
	"github.com/JNZader/goreview/goreview/internal/review"
)

var exportCmd = &cobra.Command{
	Use:   "export [format]",
	Short: "Export review results to various formats",
	Long: `Export code review results to different formats.

Supported formats:
  obsidian - Export to Obsidian vault with full metadata and wiki features

Examples:
  # Export from a JSON report
  goreview export obsidian --from report.json

  # Export with custom vault path
  goreview export obsidian --from report.json --vault ~/Documents/MyVault

  # Export from stdin
  cat report.json | goreview export obsidian

  # Export with custom tags
  goreview export obsidian --from report.json --tags sprint-42,backend`,
	RunE: runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Input source
	exportCmd.Flags().String("from", "", "Source file to export (JSON report)")

	// Obsidian options
	exportCmd.Flags().String("vault", "", "Obsidian vault path (overrides config)")
	exportCmd.Flags().String("folder", "", "Folder name within vault (default: GoReview)")
	exportCmd.Flags().String("project", "", "Project name (default: current directory name)")

	// Features
	exportCmd.Flags().Bool("no-tags", false, "Disable Obsidian tags")
	exportCmd.Flags().Bool("no-callouts", false, "Disable Obsidian callouts")
	exportCmd.Flags().Bool("no-links", false, "Disable wiki links")
	exportCmd.Flags().StringSlice("tags", nil, "Additional tags to add")

	// Template
	exportCmd.Flags().String("template", "", "Custom template file")
}

func runExport(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("export format required. Use: obsidian")
	}

	format := args[0]

	switch format {
	case "obsidian":
		return runObsidianExport(cmd)
	default:
		return fmt.Errorf("unknown export format: %s. Supported: obsidian", format)
	}
}

func runObsidianExport(cmd *cobra.Command) error {
	// Load config
	cfg, err := config.LoadDefault()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Apply flag overrides
	applyExportFlagOverrides(cmd, cfg)

	// Get review result
	result, err := loadReviewResult(cmd)
	if err != nil {
		return fmt.Errorf("loading review result: %w", err)
	}

	// Validate vault path
	if cfg.Export.Obsidian.VaultPath == "" {
		return fmt.Errorf("obsidian vault path not configured (use --vault or config)")
	}

	// Build metadata
	metadata := buildExportMetadataForExport(cmd, cfg)

	// Create exporter
	exporter, err := export.NewObsidianExporter(&cfg.Export.Obsidian)
	if err != nil {
		return fmt.Errorf("creating exporter: %w", err)
	}

	// Export
	if err := exporter.Export(result, metadata); err != nil {
		return fmt.Errorf("exporting: %w", err)
	}

	outputPath := exporter.GetOutputPath(metadata)
	fmt.Printf("Successfully exported to Obsidian: %s\n", outputPath)
	return nil
}

func applyExportFlagOverrides(cmd *cobra.Command, cfg *config.Config) {
	if vault, _ := cmd.Flags().GetString("vault"); vault != "" {
		cfg.Export.Obsidian.VaultPath = vault
	}
	if folder, _ := cmd.Flags().GetString("folder"); folder != "" {
		cfg.Export.Obsidian.FolderName = folder
	}
	if noTags, _ := cmd.Flags().GetBool("no-tags"); noTags {
		cfg.Export.Obsidian.IncludeTags = false
	}
	if noCallouts, _ := cmd.Flags().GetBool("no-callouts"); noCallouts {
		cfg.Export.Obsidian.IncludeCallouts = false
	}
	if noLinks, _ := cmd.Flags().GetBool("no-links"); noLinks {
		cfg.Export.Obsidian.IncludeLinks = false
	}
	if tags, _ := cmd.Flags().GetStringSlice("tags"); len(tags) > 0 {
		cfg.Export.Obsidian.CustomTags = append(cfg.Export.Obsidian.CustomTags, tags...)
	}
	if template, _ := cmd.Flags().GetString("template"); template != "" {
		cfg.Export.Obsidian.TemplateFile = template
	}
}

func loadReviewResult(cmd *cobra.Command) (*review.Result, error) {
	fromFile, _ := cmd.Flags().GetString("from")

	if fromFile != "" {
		// Load from JSON file
		data, err := os.ReadFile(fromFile) // #nosec G304 - user-provided file path
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}

		var result review.Result
		if err := json.Unmarshal(data, &result); err != nil {
			return nil, fmt.Errorf("parsing JSON: %w", err)
		}
		return &result, nil
	}

	// Try to read from stdin
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		var result review.Result
		if err := json.NewDecoder(os.Stdin).Decode(&result); err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		return &result, nil
	}

	return nil, fmt.Errorf("no input source. Use --from or pipe JSON to stdin")
}

func buildExportMetadataForExport(cmd *cobra.Command, cfg *config.Config) *export.Metadata {
	projectName, _ := cmd.Flags().GetString("project")

	// Try to get git info
	branch := "unknown"
	repoRoot := ""
	commitHash := ""
	commitShort := ""
	author := ""

	// Check if we're in a git repo
	if _, err := exec.Command("git", "rev-parse", "--git-dir").Output(); err == nil {
		// Get branch
		if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
			branch = strings.TrimSpace(string(out))
		}

		// Get repo root
		if out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output(); err == nil {
			repoRoot = strings.TrimSpace(string(out))
		}

		// Get commit hash
		if out, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
			commitHash = strings.TrimSpace(string(out))
			if len(commitHash) >= 7 {
				commitShort = commitHash[:7]
			}
		}

		// Get author
		if out, err := exec.Command("git", "log", "-1", "--format=%an").Output(); err == nil {
			author = strings.TrimSpace(string(out))
		}
	}

	// Determine project name
	if projectName == "" {
		if repoRoot != "" {
			projectName = filepath.Base(repoRoot)
		} else {
			cwd, _ := os.Getwd()
			projectName = filepath.Base(cwd)
		}
	}

	return &export.Metadata{
		ProjectName: projectName,
		Branch:      branch,
		CommitHash:  commitHash,
		CommitShort: commitShort,
		Author:      author,
		ReviewDate:  time.Now(),
		ReviewMode:  "export",
		BaseBranch:  cfg.Git.BaseBranch,
	}
}
