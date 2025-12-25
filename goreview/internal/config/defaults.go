package config

import (
	"os"
	"path/filepath"
	"time"
)

// DefaultConfig returns a Config with sensible default values.
// These defaults are designed to work out-of-the-box with Ollama.
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	cacheDir := filepath.Join(homeDir, ".cache", "goreview")

	return &Config{
		Provider: ProviderConfig{
			Name:         "ollama",
			Model:        "qwen2.5-coder:14b",
			BaseURL:      "http://localhost:11434",
			Timeout:      5 * time.Minute,
			MaxTokens:    4096,
			Temperature:  0.1, // Low temperature for consistent reviews
			RateLimitRPS: 0,   // No rate limit by default
		},

		Git: GitConfig{
			RepoPath:   ".",
			BaseBranch: "main",
			IgnorePatterns: []string{
				// Documentation
				"*.md",
				"*.txt",
				"*.rst",

				// Generated files
				"*.pb.go",
				"*_generated.go",
				"*.gen.go",

				// Lock files
				"go.sum",
				"package-lock.json",
				"yarn.lock",
				"pnpm-lock.yaml",

				// Build artifacts
				"dist/*",
				"build/*",
				"node_modules/*",
				"vendor/*",

				// Config files (usually not code)
				"*.json",
				"*.yaml",
				"*.yml",
				"*.toml",
			},
		},

		Review: ReviewConfig{
			Mode:           "staged",
			MinSeverity:    "warning",
			MaxIssues:      50,
			MaxConcurrency: 0,         // Auto-detect based on CPU
			Personality:    "default", // Balanced reviewer style
		},

		Output: OutputConfig{
			Format:      "markdown",
			IncludeCode: true,
			Color:       true,
			Verbose:     false,
			Quiet:       false,
		},

		Cache: CacheConfig{
			Enabled:    true,
			Dir:        cacheDir,
			TTL:        24 * time.Hour,
			MaxSizeMB:  100,
			MaxEntries: 1000,
		},

		Rules: RulesConfig{
			Preset: "standard",
		},

		Memory: MemoryConfig{
			Enabled: false, // Disabled by default
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
