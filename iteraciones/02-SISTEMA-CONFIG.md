# Iteracion 02: Sistema de Configuracion

## Objetivos

Al completar esta iteracion tendras:
- Estructuras de configuracion completas
- Loader con Viper (archivo + env + flags)
- Valores por defecto sensibles
- Validacion de configuracion
- Comando `config show` funcional
- Tests completos del sistema de config

## Prerequisitos

- Iteracion 01 completada
- CLI basico funcionando

## Tiempo Estimado: 6 horas

---

## Commit 2.1: Crear estructuras de configuracion

**Mensaje de commit:**
```
feat(config): add config struct definitions

- Define Config as main configuration struct
- Add ProviderConfig for AI provider settings
- Add GitConfig for git repository settings
- Add ReviewConfig for review behavior
- Add OutputConfig for output formatting
- Add CacheConfig for caching settings
```

**Archivos a crear:**

### 1. `goreview/internal/config/config.go`

```go
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
```

**Verificacion:**
```bash
cd goreview

# Verificar que compila
go build ./internal/config/...

# No deberia haber errores
```

**Explicacion didactica:**

**Estructuras anidadas con mapstructure tags:**

```go
type Config struct {
    Provider ProviderConfig `mapstructure:"provider" yaml:"provider"`
}
```

Los tags permiten:
- `mapstructure`: Viper usa esto para decodificar
- `yaml`: Para serializar/deserializar YAML

**Patron de validacion:**

Implementamos `Validate()` que verifica:
1. Campos requeridos
2. Valores validos (enums)
3. Dependencias entre campos

Retornamos un error estructurado (`ValidationError`) para mejor manejo.

---

## Commit 2.2: Agregar valores por defecto

**Mensaje de commit:**
```
feat(config): add default configuration values

- Set sensible defaults for all config options
- Default to Ollama with qwen2.5-coder model
- Default to staged review mode
- Configure reasonable cache settings
```

**Archivos a crear:**

### 1. `goreview/internal/config/defaults.go`

```go
package config

import (
	"os"
	"path/filepath"
	"time"
)

// DefaultConfig returns a Config with sensible default values.
// These defaults are designed to work out-of-the-box with Ollama.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
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
			MaxConcurrency: 0, // Auto-detect based on CPU
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
```

**Verificacion:**
```bash
cd goreview

# Verificar que compila
go build ./internal/config/...

# Crear un test rapido
go test -v ./internal/config/... -run TestDefault
# (Fallara porque no hay test aun, pero valida la sintaxis)
```

**Explicacion didactica:**

**Por que estos defaults:**

1. **Ollama como default**: Es gratuito y local, no requiere API key
2. **qwen2.5-coder:14b**: Buen balance calidad/velocidad para code review
3. **Temperature 0.1**: Queremos respuestas consistentes, no creativas
4. **Timeout 5min**: LLMs locales pueden ser lentos
5. **MaxConcurrency 0**: Auto-detectar CPUs es mas portable
6. **Cache 24h TTL**: Suficiente para desarrollo normal

**Patrones a ignorar:**

Los patterns se basan en que:
- Documentacion no es codigo
- Archivos generados no merecen review
- Lock files son auto-generados
- Build artifacts cambian constantemente

---

## Commit 2.3: Crear loader de configuracion

**Mensaje de commit:**
```
feat(config): add config loader with viper

- Load config from file, env vars, and defaults
- Support multiple config file locations
- Merge configurations in priority order
- Add environment variable prefix GOREVIEW_
```

**Archivos a crear:**

### 1. `goreview/internal/config/loader.go`

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
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
	v.AddConfigPath(".")                            // Current directory (highest priority)
	v.AddConfigPath("$HOME")                        // Home directory
	v.AddConfigPath("/etc/goreview")                // System config (lowest priority)

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
	if _, err := os.Stat(".goreview.yaml"); err == nil {
		if abs, err := filepath.Abs(".goreview.yaml"); err == nil {
			return abs
		}
	}

	// Check home directory
	if home, err := os.UserHomeDir(); err == nil {
		path := filepath.Join(home, ".goreview.yaml")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	// Check /etc
	if _, err := os.Stat("/etc/goreview/.goreview.yaml"); err == nil {
		return "/etc/goreview/.goreview.yaml"
	}

	return ""
}
```

**Verificacion:**
```bash
cd goreview

# Verificar que compila
go build ./internal/config/...

# Crear archivo de config para probar
cat > .goreview.yaml << 'EOF'
provider:
  name: ollama
  model: qwen2.5-coder:7b
  base_url: http://localhost:11434

review:
  mode: staged
  min_severity: info

output:
  format: markdown
  color: true
EOF

# Deberia compilar sin errores
go build ./...
```

**Explicacion didactica:**

**Viper configuration cascade:**

```
Flags > Env Vars > Config File > Defaults
```

Esto permite:
- Override en CI: `GOREVIEW_PROVIDER_NAME=openai`
- Override temporal: `--config custom.yaml`
- Defaults sensibles sin config file

**Environment variable naming:**

```go
v.SetEnvPrefix("GOREVIEW")
v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
```

Esto convierte:
- `provider.base_url` -> `GOREVIEW_PROVIDER_BASE_URL`

**Search paths:**

```go
v.AddConfigPath(".")        // Proyecto actual
v.AddConfigPath("$HOME")    // Configuracion del usuario
v.AddConfigPath("/etc/...")  // Configuracion del sistema
```

---

## Commit 2.4: Agregar tests del sistema de config

**Mensaje de commit:**
```
test(config): add configuration tests

- Test default config values
- Test config validation
- Test loader with various sources
- Test environment variable override
```

**Archivos a crear:**

### 1. `goreview/internal/config/config_test.go`

```go
package config

import (
	"os"
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
				if err == nil || !contains(err.Error(), tt.errMsg) {
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
	// Set environment variable
	os.Setenv("GOREVIEW_PROVIDER_NAME", "openai")
	os.Setenv("GOREVIEW_PROVIDER_API_KEY", "test-key")
	os.Setenv("GOREVIEW_PROVIDER_MODEL", "gpt-4")
	defer func() {
		os.Unsetenv("GOREVIEW_PROVIDER_NAME")
		os.Unsetenv("GOREVIEW_PROVIDER_API_KEY")
		os.Unsetenv("GOREVIEW_PROVIDER_MODEL")
	}()

	loader := NewLoader()
	cfg, err := loader.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Provider.Name != "openai" {
		t.Errorf("Provider.Name = %v, want openai", cfg.Provider.Name)
	}

	if cfg.Provider.APIKey != "test-key" {
		t.Errorf("Provider.APIKey = %v, want test-key", cfg.Provider.APIKey)
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

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

**Verificacion:**
```bash
cd goreview

# Ejecutar tests
go test -v ./internal/config/...

# Con coverage
go test -v -cover ./internal/config/...
```

---

## Commit 2.5: Agregar comando config show

**Mensaje de commit:**
```
feat(cli): add config show command

- Display current configuration
- Support JSON output format
- Show config file location if used
- Mask sensitive values (API keys)
```

**Archivos a crear:**

### 1. `goreview/cmd/goreview/commands/config.go`

```go
package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
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
		fmt.Printf("  api_key: %s\n", cfg.Provider.APIKey)
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
```

**Verificacion:**
```bash
cd goreview

# Recompilar
go build -o build/goreview ./cmd/goreview

# Probar comando
./build/goreview config show

# Probar con JSON
./build/goreview config show --json

# Probar con config file
./build/goreview config show --config .goreview.yaml
```

---

## Resumen de la Iteracion 02

### Commits realizados:
1. `feat(config): add config struct definitions`
2. `feat(config): add default configuration values`
3. `feat(config): add config loader with viper`
4. `test(config): add configuration tests`
5. `feat(cli): add config show command`

### Archivos creados:
```
goreview/
├── internal/config/
│   ├── config.go
│   ├── defaults.go
│   ├── loader.go
│   └── config_test.go
└── cmd/goreview/commands/
    └── config.go
```

### Verificacion final:
```bash
cd goreview

# Compilar
make build

# Tests
go test -v -cover ./internal/config/...

# Probar comandos
./build/goreview config show
./build/goreview config show --json
```

---

## Siguiente Iteracion

Continua con: **[03-GIT-INTEGRACION.md](03-GIT-INTEGRACION.md)**
