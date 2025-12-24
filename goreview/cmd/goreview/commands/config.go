package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/JNZader/goreview/goreview/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  `View and manage goreview configuration.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long: `Display the current configuration, including values from
config file, environment variables, and defaults.

Examples:
  # Show config in YAML format
  goreview config show

  # Show config as JSON
  goreview config show --json`,

	RunE: runConfigShow,
}

var (
	configShowJSON bool
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)

	configShowCmd.Flags().BoolVar(&configShowJSON, "json", false, "output as JSON")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	loader := config.NewLoader()

	// Use config file from flag if provided
	if cfgFile != "" {
		loader.SetConfigFile(cfgFile)
	}

	cfg, err := loader.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Mask sensitive values
	maskedCfg := maskSensitiveConfig(cfg)

	// Show config file location
	if !isQuiet() {
		if configFile := loader.ConfigFileUsed(); configFile != "" {
			fmt.Printf("# Config file: %s\n\n", configFile)
		} else {
			fmt.Println("# No config file found, using defaults")
			fmt.Println()
		}
	}

	if configShowJSON {
		return outputConfigJSON(maskedCfg)
	}

	return outputConfigYAML(maskedCfg)
}

// maskSensitiveConfig creates a copy with sensitive values masked
func maskSensitiveConfig(cfg *config.Config) *config.Config {
	masked := *cfg // Shallow copy

	// Mask API key
	if masked.Provider.APIKey != "" {
		masked.Provider.APIKey = "***REDACTED***"
	}

	return &masked
}

func outputConfigJSON(cfg *config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputConfigYAML(cfg *config.Config) error {
	fmt.Println("provider:")
	fmt.Printf("  name: %s\n", cfg.Provider.Name)
	fmt.Printf("  model: %s\n", cfg.Provider.Model)
	fmt.Printf("  base_url: %s\n", cfg.Provider.BaseURL)
	if cfg.Provider.APIKey != "" {
		// API key is already masked by maskSensitiveConfig before this function is called
		// The value here is "***REDACTED***", not the actual key
		fmt.Printf("  api_key: %s\n", cfg.Provider.APIKey) //nolint:gosec // Value is pre-masked
	}
	fmt.Printf("  timeout: %s\n", cfg.Provider.Timeout)
	fmt.Printf("  max_tokens: %d\n", cfg.Provider.MaxTokens)
	fmt.Printf("  temperature: %.2f\n", cfg.Provider.Temperature)

	fmt.Println("\ngit:")
	fmt.Printf("  repo_path: %s\n", cfg.Git.RepoPath)
	fmt.Printf("  base_branch: %s\n", cfg.Git.BaseBranch)
	if len(cfg.Git.IgnorePatterns) > 0 {
		fmt.Println("  ignore_patterns:")
		for _, p := range cfg.Git.IgnorePatterns[:min(5, len(cfg.Git.IgnorePatterns))] {
			fmt.Printf("    - %s\n", p)
		}
		if len(cfg.Git.IgnorePatterns) > 5 {
			fmt.Printf("    # ... and %d more\n", len(cfg.Git.IgnorePatterns)-5)
		}
	}

	fmt.Println("\nreview:")
	fmt.Printf("  mode: %s\n", cfg.Review.Mode)
	fmt.Printf("  min_severity: %s\n", cfg.Review.MinSeverity)
	fmt.Printf("  max_issues: %d\n", cfg.Review.MaxIssues)
	fmt.Printf("  max_concurrency: %d\n", cfg.Review.MaxConcurrency)

	fmt.Println("\noutput:")
	fmt.Printf("  format: %s\n", cfg.Output.Format)
	fmt.Printf("  include_code: %v\n", cfg.Output.IncludeCode)
	fmt.Printf("  color: %v\n", cfg.Output.Color)

	fmt.Println("\ncache:")
	fmt.Printf("  enabled: %v\n", cfg.Cache.Enabled)
	fmt.Printf("  dir: %s\n", cfg.Cache.Dir)
	fmt.Printf("  ttl: %s\n", cfg.Cache.TTL)
	fmt.Printf("  max_size_mb: %d\n", cfg.Cache.MaxSizeMB)

	fmt.Println("\nrules:")
	fmt.Printf("  preset: %s\n", cfg.Rules.Preset)

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
