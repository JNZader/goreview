package config

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Check provider defaults
	if cfg.Provider.Name != "ollama" {
		t.Errorf("Provider.Name = %v, want ollama", cfg.Provider.Name)
	}

	if cfg.Provider.Model != "qwen2.5-coder:14b" {
		t.Errorf("Provider.Model = %v, want qwen2.5-coder:14b", cfg.Provider.Model)
	}

	if cfg.Provider.Timeout != 5*time.Minute {
		t.Errorf("Provider.Timeout = %v, want 5m", cfg.Provider.Timeout)
	}

	// Check review defaults
	if cfg.Review.Mode != "staged" {
		t.Errorf("Review.Mode = %v, want staged", cfg.Review.Mode)
	}

	// Check output defaults
	if cfg.Output.Format != "markdown" {
		t.Errorf("Output.Format = %v, want markdown", cfg.Output.Format)
	}

	// Check cache defaults
	if !cfg.Cache.Enabled {
		t.Error("Cache.Enabled = false, want true")
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid default config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "missing provider name",
			modify: func(c *Config) {
				c.Provider.Name = ""
			},
			wantErr: true,
			errMsg:  "provider.name",
		},
		{
			name: "missing model",
			modify: func(c *Config) {
				c.Provider.Model = ""
			},
			wantErr: true,
			errMsg:  "provider.model",
		},
		{
			name: "openai without api key",
			modify: func(c *Config) {
				c.Provider.Name = "openai"
				c.Provider.APIKey = ""
			},
			wantErr: true,
			errMsg:  "api_key",
		},
		{
			name: "openai with api key",
			modify: func(c *Config) {
				c.Provider.Name = "openai"
				c.Provider.APIKey = "sk-test-key"
			},
			wantErr: false,
		},
		{
			name: "invalid review mode",
			modify: func(c *Config) {
				c.Review.Mode = "invalid"
			},
			wantErr: true,
			errMsg:  "review.mode",
		},
		{
			name: "invalid output format",
			modify: func(c *Config) {
				c.Output.Format = "invalid"
			},
			wantErr: true,
			errMsg:  "output.format",
		},
		{
			name: "cache enabled without dir",
			modify: func(c *Config) {
				c.Cache.Enabled = true
				c.Cache.Dir = ""
			},
			wantErr: true,
			errMsg:  "cache.dir",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(cfg)

			err := cfg.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errMsg != "" {
				if err == nil || !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestLoaderDefaults(t *testing.T) {
	// Remove any existing config file to test defaults
	loader := NewLoader()

	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Provider.Name != "ollama" {
		t.Errorf("Provider.Name = %v, want ollama", cfg.Provider.Name)
	}
}

func TestLoaderEnvOverride(t *testing.T) {
	// Set environment variables
	// Note: Viper with AutomaticEnv binds GOREVIEW_PROVIDER_MODEL to provider.model
	_ = os.Setenv("GOREVIEW_PROVIDER_MODEL", "codellama:7b")
	_ = os.Setenv("GOREVIEW_REVIEW_MODE", "commit")
	defer func() {
		_ = os.Unsetenv("GOREVIEW_PROVIDER_MODEL")
		_ = os.Unsetenv("GOREVIEW_REVIEW_MODE")
	}()

	loader := NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify env vars override defaults
	if cfg.Provider.Model != "codellama:7b" {
		t.Errorf("Provider.Model = %v, want codellama:7b", cfg.Provider.Model)
	}

	if cfg.Review.Mode != "commit" {
		t.Errorf("Review.Mode = %v, want commit", cfg.Review.Mode)
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test.field",
		Message: "test message",
	}

	want := "config validation error: test.field: test message"
	if err.Error() != want {
		t.Errorf("Error() = %v, want %v", err.Error(), want)
	}
}

func TestDefaultIgnorePatterns(t *testing.T) {
	patterns := DefaultIgnorePatterns()

	// Check some expected patterns exist
	expectedPatterns := []string{"*.md", "go.sum", "node_modules/*"}

	for _, expected := range expectedPatterns {
		found := false
		for _, p := range patterns {
			if p == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("DefaultIgnorePatterns() missing %q", expected)
		}
	}
}
