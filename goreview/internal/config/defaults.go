package config

import (
	"os"
	"path/filepath"
	"time"
)

// DefaultConfig returns a Config with sensible default values.
// These defaults are designed to work out-of-the-box with Ollama.
func DefaultConfig() *Config {
	cacheDir := defaultCacheDir()

	return &Config{
		Provider: defaultProviderConfig(),
		Git:      defaultGitConfig(),
		Review:   defaultReviewConfig(),
		Output:   defaultOutputConfig(),
		Cache:    defaultCacheConfig(cacheDir),
		Rules:    RulesConfig{Preset: "standard"},
		Memory:   defaultMemoryConfig(cacheDir),
		Export:   defaultExportConfig(),
	}
}

// defaultCacheDir returns the default cache directory path.
func defaultCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".cache", "goreview")
}

// defaultProviderConfig returns the default provider configuration.
func defaultProviderConfig() ProviderConfig {
	return ProviderConfig{
		Name:         "ollama",
		Model:        "qwen2.5-coder:14b",
		BaseURL:      "http://localhost:11434",
		Timeout:      5 * time.Minute,
		MaxTokens:    4096,
		Temperature:  0.1,
		RateLimitRPS: 0,
	}
}

// defaultGitConfig returns the default git configuration.
func defaultGitConfig() GitConfig {
	return GitConfig{
		RepoPath:   ".",
		BaseBranch: "main",
		IgnorePatterns: []string{
			"*.md", "*.txt", "*.rst",
			"*.pb.go", "*_generated.go", "*.gen.go",
			"go.sum", "package-lock.json", "yarn.lock", "pnpm-lock.yaml",
			"dist/*", "build/*", "node_modules/*", "vendor/*",
			"*.json", "*.yaml", "*.yml", "*.toml",
		},
	}
}

// defaultReviewConfig returns the default review configuration.
func defaultReviewConfig() ReviewConfig {
	return ReviewConfig{
		Mode:           "staged",
		MinSeverity:    "warning",
		MaxIssues:      50,
		MaxConcurrency: 0,
		Personality:    "default",
	}
}

// defaultOutputConfig returns the default output configuration.
func defaultOutputConfig() OutputConfig {
	return OutputConfig{
		Format:      "markdown",
		IncludeCode: true,
		Color:       true,
		Verbose:     false,
		Quiet:       false,
	}
}

// defaultCacheConfig returns the default cache configuration.
func defaultCacheConfig(cacheDir string) CacheConfig {
	return CacheConfig{
		Enabled:    true,
		Dir:        cacheDir,
		TTL:        24 * time.Hour,
		MaxSizeMB:  100,
		MaxEntries: 1000,
	}
}

// defaultMemoryConfig returns the default memory configuration.
func defaultMemoryConfig(cacheDir string) MemoryConfig {
	return MemoryConfig{
		Enabled: false,
		Dir:     filepath.Join(cacheDir, "memory"),
		Working: WorkingMemoryConfig{
			Capacity: 100,
			TTL:      15 * time.Minute,
		},
		Session: SessionMemoryConfig{
			MaxSessions: 10,
			SessionTTL:  1 * time.Hour,
		},
		LongTerm: LongTermMemoryConfig{
			Enabled:    false,
			MaxSizeMB:  500,
			GCInterval: 5 * time.Minute,
		},
		Hebbian: HebbianConfig{
			Enabled:      false,
			LearningRate: 0.1,
			DecayRate:    0.01,
			MinStrength:  0.1,
		},
	}
}

// defaultExportConfig returns the default export configuration.
func defaultExportConfig() ExportConfig {
	return ExportConfig{
		Obsidian: ObsidianExportConfig{
			Enabled:               false,
			VaultPath:             "",
			FolderName:            "GoReview",
			IncludeTags:           true,
			IncludeCallouts:       true,
			IncludeLinks:          true,
			LinkToPreviousReviews: true,
			CustomTags:            []string{},
			TemplateFile:          "",
		},
	}
}

// DefaultIgnorePatterns returns the default file patterns to ignore.
// These are common patterns that typically don't need code review.
func DefaultIgnorePatterns() []string {
	return []string{
		// Documentation
		"*.md",
		"*.txt",
		"*.rst",
		"LICENSE",
		"CHANGELOG*",

		// Images
		"*.png",
		"*.jpg",
		"*.jpeg",
		"*.gif",
		"*.svg",
		"*.ico",

		// Generated code
		"*.pb.go",
		"*_generated.go",
		"*.gen.go",
		"*_mock.go",

		// Dependencies
		"go.sum",
		"package-lock.json",
		"yarn.lock",
		"pnpm-lock.yaml",

		// Build output
		"dist/*",
		"build/*",
		"out/*",
		".next/*",

		// Dependencies
		"node_modules/*",
		"vendor/*",

		// IDE/Editor
		".idea/*",
		".vscode/*",
		"*.swp",

		// Testing
		"coverage/*",
		"*.out",
		"*.html",

		// Misc
		".DS_Store",
		"Thumbs.db",
	}
}
