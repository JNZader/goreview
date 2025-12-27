package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config file constants (SonarQube S1192)
const (
	configFileName = ".goreview.yaml"
)

// Loader handles configuration loading from multiple sources.
type Loader struct {
	v          *viper.Viper
	configFile string
}

// NewLoader creates a new configuration loader.
func NewLoader() *Loader {
	v := viper.New()

	// Set config name and type
	v.SetConfigName(".goreview")
	v.SetConfigType("yaml")

	// Add search paths in order of priority
	v.AddConfigPath(".")             // Current directory (highest priority)
	v.AddConfigPath("$HOME")         // Home directory
	v.AddConfigPath("/etc/goreview") // System config (lowest priority)

	// Environment variable support
	v.SetEnvPrefix("GOREVIEW")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return &Loader{v: v}
}

// SetConfigFile sets a specific config file to use.
func (l *Loader) SetConfigFile(path string) {
	l.configFile = path
	l.v.SetConfigFile(path)
}

// Load loads the configuration from all sources.
// Priority (highest to lowest):
// 1. Explicit config file (if set via SetConfigFile)
// 2. Environment variables (GOREVIEW_*)
// 3. Config file from search paths (.goreview.yaml)
// 4. Default values
func (l *Loader) Load() (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Set defaults in viper
	l.setDefaults(cfg)

	// Try to read config file
	if err := l.v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Config file found but error reading it
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found - that's ok, we'll use defaults
	}

	// Unmarshal into config struct
	if err := l.v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate the final config
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// setDefaults sets all default values in viper.
func (l *Loader) setDefaults(cfg *Config) {
	// Provider defaults
	l.v.SetDefault("provider.name", cfg.Provider.Name)
	l.v.SetDefault("provider.model", cfg.Provider.Model)
	l.v.SetDefault("provider.base_url", cfg.Provider.BaseURL)
	l.v.SetDefault("provider.timeout", cfg.Provider.Timeout)
	l.v.SetDefault("provider.max_tokens", cfg.Provider.MaxTokens)
	l.v.SetDefault("provider.temperature", cfg.Provider.Temperature)
	l.v.SetDefault("provider.rate_limit_rps", cfg.Provider.RateLimitRPS)

	// Git defaults
	l.v.SetDefault("git.repo_path", cfg.Git.RepoPath)
	l.v.SetDefault("git.base_branch", cfg.Git.BaseBranch)
	l.v.SetDefault("git.ignore_patterns", cfg.Git.IgnorePatterns)

	// Review defaults
	l.v.SetDefault("review.mode", cfg.Review.Mode)
	l.v.SetDefault("review.min_severity", cfg.Review.MinSeverity)
	l.v.SetDefault("review.max_issues", cfg.Review.MaxIssues)
	l.v.SetDefault("review.max_concurrency", cfg.Review.MaxConcurrency)

	// Output defaults
	l.v.SetDefault("output.format", cfg.Output.Format)
	l.v.SetDefault("output.include_code", cfg.Output.IncludeCode)
	l.v.SetDefault("output.color", cfg.Output.Color)
	l.v.SetDefault("output.verbose", cfg.Output.Verbose)
	l.v.SetDefault("output.quiet", cfg.Output.Quiet)

	// Cache defaults
	l.v.SetDefault("cache.enabled", cfg.Cache.Enabled)
	l.v.SetDefault("cache.dir", cfg.Cache.Dir)
	l.v.SetDefault("cache.ttl", cfg.Cache.TTL)
	l.v.SetDefault("cache.max_size_mb", cfg.Cache.MaxSizeMB)
	l.v.SetDefault("cache.max_entries", cfg.Cache.MaxEntries)

	// Rules defaults
	l.v.SetDefault("rules.preset", cfg.Rules.Preset)

	// Export defaults
	l.v.SetDefault("export.obsidian.enabled", cfg.Export.Obsidian.Enabled)
	l.v.SetDefault("export.obsidian.vault_path", cfg.Export.Obsidian.VaultPath)
	l.v.SetDefault("export.obsidian.folder_name", cfg.Export.Obsidian.FolderName)
	l.v.SetDefault("export.obsidian.include_tags", cfg.Export.Obsidian.IncludeTags)
	l.v.SetDefault("export.obsidian.include_callouts", cfg.Export.Obsidian.IncludeCallouts)
	l.v.SetDefault("export.obsidian.include_links", cfg.Export.Obsidian.IncludeLinks)
	l.v.SetDefault("export.obsidian.link_to_previous", cfg.Export.Obsidian.LinkToPreviousReviews)
	l.v.SetDefault("export.obsidian.custom_tags", cfg.Export.Obsidian.CustomTags)
	l.v.SetDefault("export.obsidian.template_file", cfg.Export.Obsidian.TemplateFile)
}

// ConfigFileUsed returns the path of the config file used, if any.
func (l *Loader) ConfigFileUsed() string {
	return l.v.ConfigFileUsed()
}

// GetViper returns the underlying viper instance for advanced usage.
func (l *Loader) GetViper() *viper.Viper {
	return l.v
}

// LoadFromFile loads configuration from a specific file.
func LoadFromFile(path string) (*Config, error) {
	loader := NewLoader()
	loader.SetConfigFile(path)
	return loader.Load()
}

// LoadDefault loads configuration with default search paths.
func LoadDefault() (*Config, error) {
	loader := NewLoader()
	return loader.Load()
}

// MustLoad loads configuration and panics on error.
// Use only in main() or init() functions.
func MustLoad() *Config {
	cfg, err := LoadDefault()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}
	return cfg
}

// FindConfigFile searches for a config file and returns its path.
// Returns empty string if no config file is found.
func FindConfigFile() string {
	// Check current directory
	if _, err := os.Stat(configFileName); err == nil {
		if abs, err := filepath.Abs(configFileName); err == nil {
			return abs
		}
	}

	// Check home directory
	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, configFileName)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check /etc
	etcPath := "/etc/goreview/" + configFileName
	if _, err := os.Stat(etcPath); err == nil {
		return etcPath
	}

	return ""
}
