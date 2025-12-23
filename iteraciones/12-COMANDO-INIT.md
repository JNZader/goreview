# Iteracion 12: Comando Init

## Objetivos

- Implementar comando `goreview init`
- Wizard de configuracion inicial
- Deteccion automatica de lenguaje
- Generacion de archivo de configuracion

## Tiempo Estimado: 3 horas

---

## Commit 12.1: Crear estructura del comando init

**Mensaje de commit:**
```
feat(cli): add init command structure

- Add init.go command file
- Define initialization flags
- Support interactive and non-interactive modes
```

### `goreview/cmd/init.go`

```go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize goreview configuration",
	Long: `Initialize goreview configuration in your project.

This command creates a .goreview.yaml configuration file with
sensible defaults based on your project structure.

Examples:
  # Interactive initialization
  goreview init

  # Non-interactive with defaults
  goreview init --yes

  # Specify provider
  goreview init --provider ollama --model codellama`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Mode flags
	initCmd.Flags().BoolP("yes", "y", false, "Accept all defaults (non-interactive)")
	initCmd.Flags().Bool("force", false, "Overwrite existing configuration")

	// Provider flags
	initCmd.Flags().String("provider", "", "AI provider (ollama, openai)")
	initCmd.Flags().String("model", "", "Model to use")
	initCmd.Flags().String("api-key", "", "API key for provider")

	// Project flags
	initCmd.Flags().String("preset", "standard", "Rule preset (minimal, standard, strict)")
	initCmd.Flags().StringSlice("exclude", nil, "Patterns to exclude")
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("Init command - implementation follows")
	return nil
}
```

---

## Commit 12.2: Implementar deteccion de proyecto

**Mensaje de commit:**
```
feat(cli): add project detection

- Detect programming languages
- Identify project type (go mod, npm, etc.)
- Suggest appropriate defaults
```

### `goreview/cmd/init_detect.go`

```go
package cmd

import (
	"os"
	"path/filepath"
	"strings"
)

// ProjectInfo contains detected project information.
type ProjectInfo struct {
	Languages   []string
	ProjectType string
	HasGit      bool
	HasCI       bool
	Frameworks  []string
}

// DetectProject analyzes the current directory for project info.
func DetectProject(dir string) (*ProjectInfo, error) {
	info := &ProjectInfo{}

	// Check for git
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		info.HasGit = true
	}

	// Check for CI
	ciFiles := []string{
		".github/workflows",
		".gitlab-ci.yml",
		".circleci",
		"Jenkinsfile",
	}
	for _, ci := range ciFiles {
		if _, err := os.Stat(filepath.Join(dir, ci)); err == nil {
			info.HasCI = true
			break
		}
	}

	// Detect languages and project types
	info.detectLanguages(dir)
	info.detectFrameworks(dir)

	return info, nil
}

func (p *ProjectInfo) detectLanguages(dir string) {
	languageFiles := map[string]string{
		"go.mod":         "go",
		"package.json":   "javascript",
		"requirements.txt": "python",
		"Cargo.toml":     "rust",
		"pom.xml":        "java",
		"build.gradle":   "java",
		"Gemfile":        "ruby",
		"composer.json":  "php",
	}

	for file, lang := range languageFiles {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			p.Languages = append(p.Languages, lang)
			p.ProjectType = file
		}
	}

	// If no specific files found, scan for source files
	if len(p.Languages) == 0 {
		p.scanForLanguages(dir)
	}
}

func (p *ProjectInfo) scanForLanguages(dir string) {
	extToLang := map[string]string{
		".go":   "go",
		".js":   "javascript",
		".ts":   "typescript",
		".py":   "python",
		".rs":   "rust",
		".java": "java",
		".rb":   "ruby",
		".php":  "php",
	}

	seen := make(map[string]bool)

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden and vendor directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := filepath.Ext(path)
		if lang, ok := extToLang[ext]; ok {
			if !seen[lang] {
				p.Languages = append(p.Languages, lang)
				seen[lang] = true
			}
		}

		return nil
	})
}

func (p *ProjectInfo) detectFrameworks(dir string) {
	// Check package.json for JS frameworks
	if data, err := os.ReadFile(filepath.Join(dir, "package.json")); err == nil {
		content := string(data)
		frameworks := []string{"react", "vue", "angular", "express", "next", "nest"}
		for _, fw := range frameworks {
			if strings.Contains(content, "\""+fw) {
				p.Frameworks = append(p.Frameworks, fw)
			}
		}
	}

	// Check go.mod for Go frameworks
	if data, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil {
		content := string(data)
		if strings.Contains(content, "gin-gonic") {
			p.Frameworks = append(p.Frameworks, "gin")
		}
		if strings.Contains(content, "echo") {
			p.Frameworks = append(p.Frameworks, "echo")
		}
		if strings.Contains(content, "fiber") {
			p.Frameworks = append(p.Frameworks, "fiber")
		}
	}
}

// SuggestDefaults returns suggested configuration based on project info.
func (p *ProjectInfo) SuggestDefaults() map[string]interface{} {
	defaults := map[string]interface{}{
		"provider":       "ollama",
		"model":          "codellama",
		"preset":         "standard",
		"exclude":        []string{},
	}

	// Language-specific excludes
	for _, lang := range p.Languages {
		switch lang {
		case "javascript", "typescript":
			defaults["exclude"] = append(defaults["exclude"].([]string),
				"node_modules/**", "dist/**", "build/**", "*.min.js")
		case "go":
			defaults["exclude"] = append(defaults["exclude"].([]string),
				"vendor/**", "*_test.go")
		case "python":
			defaults["exclude"] = append(defaults["exclude"].([]string),
				"__pycache__/**", "venv/**", ".venv/**", "*.pyc")
		}
	}

	return defaults
}

// PrimaryLanguage returns the main language of the project.
func (p *ProjectInfo) PrimaryLanguage() string {
	if len(p.Languages) > 0 {
		return p.Languages[0]
	}
	return "unknown"
}
```

---

## Commit 12.3: Implementar wizard interactivo

**Mensaje de commit:**
```
feat(cli): add interactive init wizard

- Step-by-step configuration
- Show detected settings
- Confirm before saving
```

### `goreview/cmd/init_wizard.go`

```go
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// InitWizard handles interactive initialization.
type InitWizard struct {
	reader *bufio.Reader
	info   *ProjectInfo
}

// NewInitWizard creates a new initialization wizard.
func NewInitWizard(info *ProjectInfo) *InitWizard {
	return &InitWizard{
		reader: bufio.NewReader(os.Stdin),
		info:   info,
	}
}

// Run executes the interactive wizard.
func (w *InitWizard) Run() (map[string]interface{}, error) {
	config := make(map[string]interface{})

	fmt.Println("\n┌─────────────────────────────────────┐")
	fmt.Println("│     GoReview Configuration Wizard   │")
	fmt.Println("└─────────────────────────────────────┘\n")

	// Show detected info
	w.showDetectedInfo()

	// Provider selection
	provider, err := w.selectProvider()
	if err != nil {
		return nil, err
	}
	config["provider"] = provider

	// Model selection
	model, err := w.selectModel(provider)
	if err != nil {
		return nil, err
	}
	config["model"] = model

	// API Key (if needed)
	if provider == "openai" {
		apiKey, err := w.promptAPIKey()
		if err != nil {
			return nil, err
		}
		config["api_key"] = apiKey
	}

	// Preset selection
	preset, err := w.selectPreset()
	if err != nil {
		return nil, err
	}
	config["preset"] = preset

	// Exclude patterns
	excludes := w.info.SuggestDefaults()["exclude"].([]string)
	config["exclude"] = excludes

	// Confirmation
	w.showSummary(config)
	if !w.confirm("Create configuration?") {
		return nil, fmt.Errorf("initialization cancelled")
	}

	return config, nil
}

func (w *InitWizard) showDetectedInfo() {
	fmt.Println("Detected project information:")
	fmt.Println("─────────────────────────────")

	if len(w.info.Languages) > 0 {
		fmt.Printf("  Languages:    %s\n", strings.Join(w.info.Languages, ", "))
	}
	if w.info.ProjectType != "" {
		fmt.Printf("  Project type: %s\n", w.info.ProjectType)
	}
	if len(w.info.Frameworks) > 0 {
		fmt.Printf("  Frameworks:   %s\n", strings.Join(w.info.Frameworks, ", "))
	}
	fmt.Printf("  Git repo:     %v\n", w.info.HasGit)
	fmt.Printf("  CI detected:  %v\n", w.info.HasCI)
	fmt.Println()
}

func (w *InitWizard) selectProvider() (string, error) {
	fmt.Println("Select AI provider:")
	fmt.Println("  [1] Ollama (local, free)")
	fmt.Println("  [2] OpenAI (cloud, requires API key)")
	fmt.Print("\nChoice [1]: ")

	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "", "1":
		return "ollama", nil
	case "2":
		return "openai", nil
	default:
		return "ollama", nil
	}
}

func (w *InitWizard) selectModel(provider string) (string, error) {
	var options []string
	var defaultModel string

	switch provider {
	case "ollama":
		options = []string{"qwen2.5-coder:14b", "codellama", "deepseek-coder", "mistral"}
		defaultModel = "qwen2.5-coder:14b"
	case "openai":
		options = []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"}
		defaultModel = "gpt-4"
	}

	fmt.Println("\nSelect model:")
	for i, opt := range options {
		def := ""
		if opt == defaultModel {
			def = " (recommended)"
		}
		fmt.Printf("  [%d] %s%s\n", i+1, opt, def)
	}
	fmt.Printf("\nChoice [1]: ")

	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultModel, nil
	}

	idx := 0
	fmt.Sscanf(input, "%d", &idx)
	if idx > 0 && idx <= len(options) {
		return options[idx-1], nil
	}

	return defaultModel, nil
}

func (w *InitWizard) promptAPIKey() (string, error) {
	fmt.Print("\nEnter OpenAI API key: ")
	input, _ := w.reader.ReadString('\n')
	return strings.TrimSpace(input), nil
}

func (w *InitWizard) selectPreset() (string, error) {
	fmt.Println("\nSelect rule preset:")
	fmt.Println("  [1] minimal  - Only critical security rules")
	fmt.Println("  [2] standard - Recommended for most projects")
	fmt.Println("  [3] strict   - Maximum code quality checks")
	fmt.Print("\nChoice [2]: ")

	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		return "minimal", nil
	case "", "2":
		return "standard", nil
	case "3":
		return "strict", nil
	default:
		return "standard", nil
	}
}

func (w *InitWizard) showSummary(config map[string]interface{}) {
	fmt.Println("\n┌─────────────────────────────────────┐")
	fmt.Println("│         Configuration Summary       │")
	fmt.Println("├─────────────────────────────────────┤")
	fmt.Printf("│  Provider: %-24s │\n", config["provider"])
	fmt.Printf("│  Model:    %-24s │\n", config["model"])
	fmt.Printf("│  Preset:   %-24s │\n", config["preset"])
	fmt.Println("└─────────────────────────────────────┘")
}

func (w *InitWizard) confirm(message string) bool {
	fmt.Printf("\n%s [Y/n]: ", message)
	input, _ := w.reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes"
}
```

---

## Commit 12.4: Implementar generacion de configuracion

**Mensaje de commit:**
```
feat(cli): generate configuration file

- Create .goreview.yaml
- Support YAML and JSON formats
- Handle existing configuration
```

### Actualizar `goreview/cmd/init.go`:

```go
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const configFileName = ".goreview.yaml"

func runInit(cmd *cobra.Command, args []string) error {
	// Check for existing config
	if _, err := os.Stat(configFileName); err == nil {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			return fmt.Errorf("configuration file already exists. Use --force to overwrite")
		}
	}

	// Detect project
	cwd, _ := os.Getwd()
	info, err := DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("detecting project: %w", err)
	}

	var config map[string]interface{}

	// Interactive or non-interactive mode
	yes, _ := cmd.Flags().GetBool("yes")
	if yes {
		config = buildConfigFromFlags(cmd, info)
	} else {
		wizard := NewInitWizard(info)
		config, err = wizard.Run()
		if err != nil {
			return err
		}
	}

	// Generate YAML
	yamlConfig := buildYAMLConfig(config, info)

	// Write configuration file
	data, err := yaml.Marshal(yamlConfig)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(configFileName, data, 0644); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	fmt.Printf("\nConfiguration saved to %s\n", configFileName)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review the configuration file")

	if config["provider"] == "ollama" {
		fmt.Println("  2. Ensure Ollama is running: ollama serve")
		fmt.Printf("  3. Pull the model: ollama pull %s\n", config["model"])
	} else {
		fmt.Println("  2. Set OPENAI_API_KEY environment variable")
	}

	fmt.Println("\nRun 'goreview review --staged' to review staged changes")

	return nil
}

func buildConfigFromFlags(cmd *cobra.Command, info *ProjectInfo) map[string]interface{} {
	config := info.SuggestDefaults()

	if provider, _ := cmd.Flags().GetString("provider"); provider != "" {
		config["provider"] = provider
	}
	if model, _ := cmd.Flags().GetString("model"); model != "" {
		config["model"] = model
	}
	if preset, _ := cmd.Flags().GetString("preset"); preset != "" {
		config["preset"] = preset
	}
	if excludes, _ := cmd.Flags().GetStringSlice("exclude"); len(excludes) > 0 {
		config["exclude"] = excludes
	}

	return config
}

func buildYAMLConfig(config map[string]interface{}, info *ProjectInfo) map[string]interface{} {
	yamlConfig := map[string]interface{}{
		"version": "1.0",
		"provider": map[string]interface{}{
			"name":  config["provider"],
			"model": config["model"],
		},
		"review": map[string]interface{}{
			"max_concurrency": 5,
			"timeout":         "5m",
		},
		"git": map[string]interface{}{
			"base_branch":     "main",
			"ignore_patterns": config["exclude"],
		},
		"rules": map[string]interface{}{
			"preset": config["preset"],
		},
		"cache": map[string]interface{}{
			"enabled":     true,
			"ttl":         "24h",
			"max_entries": 100,
		},
	}

	// Add Ollama-specific config
	if config["provider"] == "ollama" {
		yamlConfig["provider"].(map[string]interface{})["base_url"] = "http://localhost:11434"
	}

	// Add API key placeholder for OpenAI
	if config["provider"] == "openai" {
		yamlConfig["provider"].(map[string]interface{})["api_key"] = "${OPENAI_API_KEY}"
	}

	return yamlConfig
}
```

---

## Commit 12.5: Tests del comando init

**Mensaje de commit:**
```
test(cli): add init command tests

- Test project detection
- Test configuration generation
- Test wizard flow
```

### `goreview/cmd/init_test.go`

```go
package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProject(t *testing.T) {
	// Create temp directory with test files
	dir := t.TempDir()

	// Create go.mod
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test"), 0644)
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0644)

	// Create .git directory
	os.Mkdir(filepath.Join(dir, ".git"), 0755)

	info, err := DetectProject(dir)
	if err != nil {
		t.Fatalf("DetectProject() error = %v", err)
	}

	if !info.HasGit {
		t.Error("HasGit should be true")
	}

	found := false
	for _, lang := range info.Languages {
		if lang == "go" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Should detect Go language")
	}
}

func TestProjectInfoSuggestDefaults(t *testing.T) {
	tests := []struct {
		name      string
		languages []string
		wantExcl  []string
	}{
		{
			name:      "go project",
			languages: []string{"go"},
			wantExcl:  []string{"vendor/**"},
		},
		{
			name:      "js project",
			languages: []string{"javascript"},
			wantExcl:  []string{"node_modules/**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &ProjectInfo{Languages: tt.languages}
			defaults := info.SuggestDefaults()

			excludes := defaults["exclude"].([]string)
			found := false
			for _, excl := range excludes {
				for _, want := range tt.wantExcl {
					if excl == want {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("Expected excludes %v, got %v", tt.wantExcl, excludes)
			}
		})
	}
}

func TestBuildYAMLConfig(t *testing.T) {
	config := map[string]interface{}{
		"provider": "ollama",
		"model":    "codellama",
		"preset":   "standard",
		"exclude":  []string{"vendor/**"},
	}

	info := &ProjectInfo{Languages: []string{"go"}}

	yamlConfig := buildYAMLConfig(config, info)

	// Check structure
	if yamlConfig["version"] != "1.0" {
		t.Error("Version should be 1.0")
	}

	provider := yamlConfig["provider"].(map[string]interface{})
	if provider["name"] != "ollama" {
		t.Error("Provider name should be ollama")
	}
	if provider["model"] != "codellama" {
		t.Error("Model should be codellama")
	}

	rules := yamlConfig["rules"].(map[string]interface{})
	if rules["preset"] != "standard" {
		t.Error("Preset should be standard")
	}
}

func TestPrimaryLanguage(t *testing.T) {
	tests := []struct {
		languages []string
		want      string
	}{
		{[]string{"go", "javascript"}, "go"},
		{[]string{"python"}, "python"},
		{[]string{}, "unknown"},
	}

	for _, tt := range tests {
		info := &ProjectInfo{Languages: tt.languages}
		got := info.PrimaryLanguage()
		if got != tt.want {
			t.Errorf("PrimaryLanguage() = %q, want %q", got, tt.want)
		}
	}
}
```

---

## Resumen de la Iteracion 12

### Commits:
1. `feat(cli): add init command structure`
2. `feat(cli): add project detection`
3. `feat(cli): add interactive init wizard`
4. `feat(cli): generate configuration file`
5. `test(cli): add init command tests`

### Archivos:
```
goreview/cmd/
├── init.go
├── init_detect.go
├── init_wizard.go
└── init_test.go
```

---

## Siguiente Iteracion

Continua con: **[13-GITHUB-APP-SETUP.md](13-GITHUB-APP-SETUP.md)**
