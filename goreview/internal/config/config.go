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

	// Memory configures the cognitive memory system
	Memory MemoryConfig `mapstructure:"memory" yaml:"memory"`

	// RAG configures Retrieval-Augmented Generation with external docs
	RAG RAGConfig `mapstructure:"rag" yaml:"rag"`

	// Export configures export behavior to external systems
	Export ExportConfig `mapstructure:"export" yaml:"export"`
}

// RAGConfig configures the RAG system for external documentation.
type RAGConfig struct {
	// Enabled enables/disables RAG
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// CacheDir is the directory for cached documentation
	CacheDir string `mapstructure:"cache_dir" yaml:"cache_dir"`

	// DefaultCacheTTL is the default cache duration
	DefaultCacheTTL string `mapstructure:"default_cache_ttl" yaml:"default_cache_ttl"`

	// MaxCacheSize is the maximum cache size in bytes
	MaxCacheSize int64 `mapstructure:"max_cache_size" yaml:"max_cache_size"`

	// AutoDetect enables automatic framework detection
	AutoDetect bool `mapstructure:"auto_detect" yaml:"auto_detect"`

	// Sources is the list of external documentation sources
	Sources []RAGSource `mapstructure:"sources" yaml:"sources"`
}

// RAGSource represents an external documentation source.
type RAGSource struct {
	URL      string `mapstructure:"url" yaml:"url"`
	Type     string `mapstructure:"type" yaml:"type"`
	Name     string `mapstructure:"name" yaml:"name"`
	Language string `mapstructure:"language,omitempty" yaml:"language,omitempty"`
	CacheTTL string `mapstructure:"cache_ttl,omitempty" yaml:"cache_ttl,omitempty"`
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
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

	// Personality is the reviewer personality style: "default", "senior", "strict", "friendly", "security-expert"
	Personality string `mapstructure:"personality" yaml:"personality"`

	// Modes specifies specialized review focus areas: "security", "perf", "clean", "docs", "tests"
	// Multiple modes can be combined with commas: "security,perf"
	Modes string `mapstructure:"modes" yaml:"modes"`

	// RootCauseTracing enables root cause analysis for each issue
	RootCauseTracing bool `mapstructure:"root_cause_tracing" yaml:"root_cause_tracing"`
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

	// InheritFrom specifies sources to inherit rules from (URLs or local paths)
	// Rules are merged with later sources taking precedence
	// Example: ["https://company.com/rules.yaml", "./team-rules.yaml"]
	InheritFrom []string `mapstructure:"inherit_from" yaml:"inherit_from"`

	// Override contains rule property overrides for this project
	// Example: {"SEC-001": {"severity": "critical"}}
	Override map[string]interface{} `mapstructure:"override" yaml:"override"`
}

// MemoryConfig configures the cognitive memory system.
type MemoryConfig struct {
	// Enabled enables the memory system
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// Dir is the directory for persistent memory storage
	Dir string `mapstructure:"dir" yaml:"dir"`

	// Working configures working memory (short-term, in-memory)
	Working WorkingMemoryConfig `mapstructure:"working" yaml:"working"`

	// Session configures session memory (per-session persistence)
	Session SessionMemoryConfig `mapstructure:"session" yaml:"session"`

	// LongTerm configures long-term memory (BadgerDB)
	LongTerm LongTermMemoryConfig `mapstructure:"long_term" yaml:"long_term"`

	// Hebbian configures Hebbian learning (association strengthening)
	Hebbian HebbianConfig `mapstructure:"hebbian" yaml:"hebbian"`
}

// WorkingMemoryConfig configures working memory.
type WorkingMemoryConfig struct {
	// Capacity is the maximum number of items in working memory
	Capacity int `mapstructure:"capacity" yaml:"capacity"`

	// TTL is how long items stay in working memory
	TTL time.Duration `mapstructure:"ttl" yaml:"ttl"`
}

// SessionMemoryConfig configures session memory.
type SessionMemoryConfig struct {
	// MaxSessions is the maximum number of concurrent sessions
	MaxSessions int `mapstructure:"max_sessions" yaml:"max_sessions"`

	// SessionTTL is how long sessions are kept
	SessionTTL time.Duration `mapstructure:"session_ttl" yaml:"session_ttl"`
}

// LongTermMemoryConfig configures long-term memory.
type LongTermMemoryConfig struct {
	// Enabled enables long-term memory persistence
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// MaxSizeMB is the maximum storage size in megabytes
	MaxSizeMB int `mapstructure:"max_size_mb" yaml:"max_size_mb"`

	// GCInterval is how often to run garbage collection
	GCInterval time.Duration `mapstructure:"gc_interval" yaml:"gc_interval"`
}

// HebbianConfig configures Hebbian learning.
type HebbianConfig struct {
	// Enabled enables Hebbian learning
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// LearningRate controls association strengthening speed
	LearningRate float64 `mapstructure:"learning_rate" yaml:"learning_rate"`

	// DecayRate controls how fast associations weaken
	DecayRate float64 `mapstructure:"decay_rate" yaml:"decay_rate"`

	// MinStrength is the minimum association strength before removal
	MinStrength float64 `mapstructure:"min_strength" yaml:"min_strength"`
}

// ExportConfig configures export behavior to external systems.
type ExportConfig struct {
	// Obsidian configures Obsidian vault export
	Obsidian ObsidianExportConfig `mapstructure:"obsidian" yaml:"obsidian"`
}

// ObsidianExportConfig configures Obsidian export settings.
type ObsidianExportConfig struct {
	// Enabled enables automatic Obsidian export after reviews
	Enabled bool `mapstructure:"enabled" yaml:"enabled"`

	// VaultPath is the path to the Obsidian vault
	VaultPath string `mapstructure:"vault_path" yaml:"vault_path"`

	// FolderName is the folder name within the vault (default: "GoReview")
	FolderName string `mapstructure:"folder_name" yaml:"folder_name"`

	// IncludeTags enables Obsidian tag generation (#security, #bug, etc.)
	IncludeTags bool `mapstructure:"include_tags" yaml:"include_tags"`

	// IncludeCallouts enables Obsidian callouts (> [!warning])
	IncludeCallouts bool `mapstructure:"include_callouts" yaml:"include_callouts"`

	// IncludeLinks enables wiki-style links [[related]]
	IncludeLinks bool `mapstructure:"include_links" yaml:"include_links"`

	// LinkToPreviousReviews links to previous reviews of the same project
	LinkToPreviousReviews bool `mapstructure:"link_to_previous" yaml:"link_to_previous"`

	// CustomTags are additional tags to add to all exports
	CustomTags []string `mapstructure:"custom_tags" yaml:"custom_tags"`

	// TemplateFile is an optional custom template file path
	TemplateFile string `mapstructure:"template_file" yaml:"template_file"`
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
