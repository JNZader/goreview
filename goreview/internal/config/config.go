// Package config handles all configuration management for goreview.
//
// Configuration is loaded from multiple sources in order of precedence:
// 1. Command-line flags (highest priority)
// 2. Environment variables (GOREVIEW_*)
// 3. Configuration file (.goreview.yaml)
// 4. Default values (lowest priority)
package config

import (
	"time"
)

// Config is the main configuration structure for goreview.
// It contains all settings needed to run the application.
type Config struct {
	// Provider configures the AI provider (Ollama, OpenAI, etc.)
	Provider ProviderConfig `mapstructure:"provider" yaml:"provider"`

	// Git configures git-related settings
	Git GitConfig `mapstructure:"git" yaml:"git"`

	// Review configures review behavior
	Review ReviewConfig `mapstructure:"review" yaml:"review"`

	// Output configures output formatting
	Output OutputConfig `mapstructure:"output" yaml:"output"`

	// Cache configures caching behavior
	Cache CacheConfig `mapstructure:"cache" yaml:"cache"`

	// Rules configures the rule system
	Rules RulesConfig `mapstructure:"rules" yaml:"rules"`
}

// ProviderConfig configures the AI provider.
type ProviderConfig struct {
	// Name is the provider name: "ollama", "openai"
	Name string `mapstructure:"name" yaml:"name"`

	// Model is the model to use (e.g., "qwen2.5-coder:14b", "gpt-4")
	Model string `mapstructure:"model" yaml:"model"`

	// BaseURL is the API base URL
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`

	// APIKey is the API key (for OpenAI, etc.)
	// This should be set via environment variable, not config file
	APIKey string `mapstructure:"api_key" yaml:"api_key"`

	// Timeout is the request timeout
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`

	// MaxTokens is the maximum tokens in response
	MaxTokens int `mapstructure:"max_tokens" yaml:"max_tokens"`

	// Temperature controls randomness (0.0 = deterministic, 1.0 = creative)
	Temperature float64 `mapstructure:"temperature" yaml:"temperature"`

	// RateLimitRPS is requests per second limit (0 = unlimited)
	RateLimitRPS int `mapstructure:"rate_limit_rps" yaml:"rate_limit_rps"`
}

// GitConfig configures git-related settings.
type GitConfig struct {
	// RepoPath is the path to the git repository (default: current directory)
	RepoPath string `mapstructure:"repo_path" yaml:"repo_path"`

	// BaseBranch is the base branch for comparisons (default: main)
	BaseBranch string `mapstructure:"base_branch" yaml:"base_branch"`

	// IgnorePatterns are file patterns to ignore during review
	IgnorePatterns []string `mapstructure:"ignore_patterns" yaml:"ignore_patterns"`
}

// ReviewConfig configures review behavior.
type ReviewConfig struct {
	// Mode is the review mode: "staged", "commit", "branch", "files"
	Mode string `mapstructure:"mode" yaml:"mode"`

	// Commit is the commit SHA to review (for mode=commit)
	Commit string `mapstructure:"commit" yaml:"commit"`

	// Files is the list of files to review (for mode=files)
	Files []string `mapstructure:"files" yaml:"files"`

	// MinSeverity is the minimum severity to report: "info", "warning", "error", "critical"
	MinSeverity string `mapstructure:"min_severity" yaml:"min_severity"`

	// MaxIssues is the maximum number of issues to report (0 = unlimited)
	MaxIssues int `mapstructure:"max_issues" yaml:"max_issues"`

	// MaxConcurrency is the maximum parallel file reviews (0 = auto)
	MaxConcurrency int `mapstructure:"max_concurrency" yaml:"max_concurrency"`

	// Context is additional context to include in prompts
	Context string `mapstructure:"context" yaml:"context"`
}

// OutputConfig configures output formatting.
type OutputConfig struct {
	// Format is the output format: "markdown", "json", "sarif"
	Format string `mapstructure:"format" yaml:"format"`

	// File is the output file path (empty = stdout)
	File string `mapstructure:"file" yaml:"file"`

	// IncludeCode includes code snippets in output
	IncludeCode bool `mapstructure:"include_code" yaml:"include_code"`

	// Color enables colored output (for terminal)
	Color bool `mapstructure:"color" yaml:"color"`

	// Verbose enables verbose output
	Verbose bool `mapstructure:"verbose" yaml:"verbose"`

	// Quiet suppresses all output except errors
	Quiet bool `mapstructure:"quiet" yaml:"quiet"`
}

// CacheConfig configures caching behavior.
type CacheConfig struct {
	// Enabled enables caching
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// Dir is the cache directory
	Dir string `mapstructure:"dir" yaml:"dir"`

	// TTL is the cache entry time-to-live
	TTL time.Duration `mapstructure:"ttl" yaml:"ttl"`

	// MaxSizeMB is the maximum cache size in megabytes
	MaxSizeMB int `mapstructure:"max_size_mb" yaml:"max_size_mb"`

	// MaxEntries is the maximum number of cache entries (for LRU)
	MaxEntries int `mapstructure:"max_entries" yaml:"max_entries"`
}

// RulesConfig configures the rule system.
type RulesConfig struct {
	// Preset is the rule preset to use: "standard", "strict", "minimal"
	Preset string `mapstructure:"preset" yaml:"preset"`

	// RulesDir is the directory containing custom rules
	RulesDir string `mapstructure:"rules_dir" yaml:"rules_dir"`

	// Enabled is the list of enabled rule IDs (empty = all)
	Enabled []string `mapstructure:"enabled" yaml:"enabled"`

	// Disabled is the list of disabled rule IDs
	Disabled []string `mapstructure:"disabled" yaml:"disabled"`
}

// Validate validates the configuration and returns an error if invalid.
func (c *Config) Validate() error {
	// Provider validation
	if c.Provider.Name == "" {
		return &ValidationError{Field: "provider.name", Message: "provider name is required"}
	}

	if c.Provider.Model == "" {
		return &ValidationError{Field: "provider.model", Message: "model is required"}
	}

	if c.Provider.Name == "openai" && c.Provider.APIKey == "" {
		return &ValidationError{Field: "provider.api_key", Message: "API key is required for OpenAI"}
	}

	// Review validation
	validModes := map[string]bool{"staged": true, "commit": true, "branch": true, "files": true}
	if !validModes[c.Review.Mode] {
		return &ValidationError{Field: "review.mode", Message: "invalid mode, must be one of: staged, commit, branch, files"}
	}

	// Output validation
	validFormats := map[string]bool{"markdown": true, "json": true, "sarif": true}
	if !validFormats[c.Output.Format] {
		return &ValidationError{Field: "output.format", Message: "invalid format, must be one of: markdown, json, sarif"}
	}

	// Cache validation
	if c.Cache.Enabled && c.Cache.Dir == "" {
		return &ValidationError{Field: "cache.dir", Message: "cache directory is required when cache is enabled"}
	}

	return nil
}

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return "config validation error: " + e.Field + ": " + e.Message
}
