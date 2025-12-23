# Iteracion 08: Sistema de Reglas

## Objetivos

- Definir estructura de reglas
- Loader de archivos YAML
- Filtrado por lenguaje y path
- Soporte para presets
- Archivos de reglas base

## Tiempo Estimado: 6 horas

---

## Commit 8.1: Crear tipos de reglas

**Mensaje de commit:**
```
feat(rules): add rule struct definitions

- Define Rule struct with all fields
- Add Severity and Category enums
- Support language-specific rules
```

### `goreview/internal/rules/types.go`

```go
package rules

// Rule defines a code review rule.
type Rule struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Category    Category `yaml:"category" json:"category"`
	Severity    Severity `yaml:"severity" json:"severity"`
	Languages   []string `yaml:"languages" json:"languages"`
	Patterns    []string `yaml:"patterns" json:"patterns"` // File patterns
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	Message     string   `yaml:"message" json:"message"`
	Suggestion  string   `yaml:"suggestion" json:"suggestion"`
}

// Category categorizes rules.
type Category string

const (
	CategorySecurity     Category = "security"
	CategoryPerformance  Category = "performance"
	CategoryBestPractice Category = "best_practice"
	CategoryStyle        Category = "style"
	CategoryBug          Category = "bug"
	CategoryMaintenance  Category = "maintenance"
)

// Severity indicates rule importance.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// RuleSet contains a collection of rules.
type RuleSet struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Rules       []Rule `yaml:"rules" json:"rules"`
}

// Preset defines a collection of enabled rules.
type Preset struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Includes    []string `yaml:"includes" json:"includes"` // Rule IDs
	Excludes    []string `yaml:"excludes" json:"excludes"` // Rule IDs
}
```

---

## Commit 8.2: Implementar loader de reglas

**Mensaje de commit:**
```
feat(rules): add yaml rule loader

- Load rules from YAML files
- Support directory scanning
- Merge multiple rule files
- Handle embedded rules
```

### `goreview/internal/rules/loader.go`

```go
package rules

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

//go:embed defaults/*.yaml
var embeddedRules embed.FS

// Loader handles loading rules from files.
type Loader struct {
	rulesDir string
}

// NewLoader creates a new rule loader.
func NewLoader(rulesDir string) *Loader {
	return &Loader{rulesDir: rulesDir}
}

// Load loads all rules from configured sources.
func (l *Loader) Load() ([]Rule, error) {
	var allRules []Rule

	// Load embedded default rules
	embedded, err := l.loadEmbedded()
	if err != nil {
		return nil, fmt.Errorf("loading embedded rules: %w", err)
	}
	allRules = append(allRules, embedded...)

	// Load custom rules if directory exists
	if l.rulesDir != "" {
		custom, err := l.loadFromDir(l.rulesDir)
		if err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("loading custom rules: %w", err)
		}
		allRules = append(allRules, custom...)
	}

	return allRules, nil
}

func (l *Loader) loadEmbedded() ([]Rule, error) {
	var allRules []Rule

	entries, err := embeddedRules.ReadDir("defaults")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		data, err := embeddedRules.ReadFile("defaults/" + entry.Name())
		if err != nil {
			return nil, err
		}

		rules, err := parseRulesYAML(data)
		if err != nil {
			return nil, fmt.Errorf("parsing %s: %w", entry.Name(), err)
		}

		allRules = append(allRules, rules...)
	}

	return allRules, nil
}

func (l *Loader) loadFromDir(dir string) ([]Rule, error) {
	var allRules []Rule

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".yaml" && ext != ".yml" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		rules, err := parseRulesYAML(data)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		allRules = append(allRules, rules...)
		return nil
	})

	return allRules, err
}

func parseRulesYAML(data []byte) ([]Rule, error) {
	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return nil, err
	}
	return ruleSet.Rules, nil
}

// LoadPreset loads a preset by name.
func (l *Loader) LoadPreset(name string) (*Preset, error) {
	// Define built-in presets
	presets := map[string]*Preset{
		"standard": {
			Name:        "standard",
			Description: "Standard rules for most projects",
			Includes:    []string{}, // Empty means all
			Excludes:    []string{},
		},
		"strict": {
			Name:        "strict",
			Description: "Strict rules for high-quality code",
			Includes:    []string{},
			Excludes:    []string{},
		},
		"minimal": {
			Name:        "minimal",
			Description: "Only critical security rules",
			Includes:    []string{"SEC-001", "SEC-002", "SEC-003"},
			Excludes:    []string{},
		},
	}

	preset, ok := presets[name]
	if !ok {
		return nil, fmt.Errorf("unknown preset: %s", name)
	}
	return preset, nil
}
```

---

## Commit 8.3: Implementar filtrado de reglas

**Mensaje de commit:**
```
feat(rules): add rule filtering by language

- Filter rules by programming language
- Filter by file path patterns
- Apply preset includes/excludes
```

### `goreview/internal/rules/filter.go`

```go
package rules

import (
	"path/filepath"
	"strings"
)

// Filter filters rules based on criteria.
func Filter(rules []Rule, language, filePath string) []Rule {
	var filtered []Rule

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// Check language
		if len(rule.Languages) > 0 && !containsString(rule.Languages, language) {
			continue
		}

		// Check file patterns
		if len(rule.Patterns) > 0 && !matchesAnyPattern(rule.Patterns, filePath) {
			continue
		}

		filtered = append(filtered, rule)
	}

	return filtered
}

// ApplyPreset applies a preset to filter rules.
func ApplyPreset(rules []Rule, preset *Preset) []Rule {
	if preset == nil {
		return rules
	}

	var filtered []Rule

	for _, rule := range rules {
		// Check excludes
		if containsString(preset.Excludes, rule.ID) {
			continue
		}

		// Check includes (empty means all)
		if len(preset.Includes) > 0 && !containsString(preset.Includes, rule.ID) {
			continue
		}

		filtered = append(filtered, rule)
	}

	return filtered
}

// GetRulesByCategory returns rules for a specific category.
func GetRulesByCategory(rules []Rule, category Category) []Rule {
	var filtered []Rule
	for _, rule := range rules {
		if rule.Category == category {
			filtered = append(filtered, rule)
		}
	}
	return filtered
}

// GetRuleBySeverity returns rules at or above severity.
func GetRulesBySeverity(rules []Rule, minSeverity Severity) []Rule {
	severityOrder := map[Severity]int{
		SeverityInfo:     0,
		SeverityWarning:  1,
		SeverityError:    2,
		SeverityCritical: 3,
	}

	minOrder := severityOrder[minSeverity]
	var filtered []Rule

	for _, rule := range rules {
		if severityOrder[rule.Severity] >= minOrder {
			filtered = append(filtered, rule)
		}
	}

	return filtered
}

func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}

func matchesAnyPattern(patterns []string, path string) bool {
	for _, pattern := range patterns {
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if matched {
			return true
		}
	}
	return false
}
```

---

## Commit 8.4: Agregar reglas base

**Mensaje de commit:**
```
chore(rules): add base rules yaml

- Add security rules (hardcoded secrets, SQL injection)
- Add performance rules (N+1 queries, memory leaks)
- Add best practice rules (error handling, logging)
```

### `goreview/internal/rules/defaults/base.yaml`

```yaml
name: Base Rules
description: Base code review rules

rules:
  - id: SEC-001
    name: Hardcoded Secrets
    description: Detect hardcoded passwords, API keys, tokens
    category: security
    severity: critical
    languages: [go, python, javascript, typescript]
    enabled: true
    message: Hardcoded secret detected. Use environment variables.
    suggestion: Move secrets to environment variables or secret manager.

  - id: SEC-002
    name: SQL Injection
    description: Detect potential SQL injection vulnerabilities
    category: security
    severity: critical
    languages: [go, python, javascript, typescript]
    enabled: true
    message: Potential SQL injection. Use parameterized queries.
    suggestion: Use prepared statements with placeholders.

  - id: SEC-003
    name: Command Injection
    description: Detect potential command injection
    category: security
    severity: critical
    languages: [go, python, javascript, typescript]
    enabled: true
    message: Potential command injection. Sanitize user input.
    suggestion: Avoid shell execution or properly escape arguments.

  - id: PERF-001
    name: N+1 Query
    description: Detect N+1 query patterns
    category: performance
    severity: warning
    languages: [go, python, javascript, typescript]
    enabled: true
    message: Potential N+1 query detected.
    suggestion: Use batch loading or eager loading.

  - id: PERF-002
    name: Unbounded Collection
    description: Detect unbounded collections in memory
    category: performance
    severity: warning
    languages: [go, javascript, typescript]
    enabled: true
    message: Unbounded collection may cause memory issues.
    suggestion: Add pagination or limits.

  - id: BP-001
    name: Error Not Handled
    description: Detect unhandled errors
    category: best_practice
    severity: error
    languages: [go]
    enabled: true
    message: Error is not checked.
    suggestion: Always check and handle errors.

  - id: BP-002
    name: Missing Logging
    description: Detect catch blocks without logging
    category: best_practice
    severity: warning
    languages: [javascript, typescript, python]
    enabled: true
    message: Exception caught but not logged.
    suggestion: Add logging in catch blocks.

  - id: BP-003
    name: Magic Numbers
    description: Detect unexplained numeric literals
    category: best_practice
    severity: info
    languages: [go, python, javascript, typescript]
    enabled: true
    message: Magic number detected.
    suggestion: Use named constants instead.
```

---

## Commit 8.5: Agregar tests de reglas

**Mensaje de commit:**
```
test(rules): add loader and filter tests

- Test YAML loading
- Test language filtering
- Test preset application
- Test severity filtering
```

### `goreview/internal/rules/rules_test.go`

```go
package rules

import (
	"testing"
)

func TestFilter(t *testing.T) {
	rules := []Rule{
		{ID: "R1", Languages: []string{"go"}, Enabled: true},
		{ID: "R2", Languages: []string{"python"}, Enabled: true},
		{ID: "R3", Languages: []string{"go", "python"}, Enabled: true},
		{ID: "R4", Languages: []string{}, Enabled: true}, // All languages
		{ID: "R5", Languages: []string{"go"}, Enabled: false},
	}

	filtered := Filter(rules, "go", "main.go")

	if len(filtered) != 3 {
		t.Errorf("len(filtered) = %d, want 3", len(filtered))
	}

	// Should have R1, R3, R4 (not R2-python, not R5-disabled)
	ids := make(map[string]bool)
	for _, r := range filtered {
		ids[r.ID] = true
	}

	if !ids["R1"] || !ids["R3"] || !ids["R4"] {
		t.Error("Missing expected rules")
	}
	if ids["R2"] || ids["R5"] {
		t.Error("Unexpected rules included")
	}
}

func TestApplyPreset(t *testing.T) {
	rules := []Rule{
		{ID: "SEC-001", Enabled: true},
		{ID: "SEC-002", Enabled: true},
		{ID: "PERF-001", Enabled: true},
	}

	preset := &Preset{
		Includes: []string{"SEC-001", "SEC-002"},
		Excludes: []string{},
	}

	filtered := ApplyPreset(rules, preset)

	if len(filtered) != 2 {
		t.Errorf("len(filtered) = %d, want 2", len(filtered))
	}
}

func TestGetRulesBySeverity(t *testing.T) {
	rules := []Rule{
		{ID: "R1", Severity: SeverityInfo},
		{ID: "R2", Severity: SeverityWarning},
		{ID: "R3", Severity: SeverityError},
		{ID: "R4", Severity: SeverityCritical},
	}

	// Get warning and above
	filtered := GetRulesBySeverity(rules, SeverityWarning)

	if len(filtered) != 3 {
		t.Errorf("len(filtered) = %d, want 3", len(filtered))
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"go", "Python", "JAVASCRIPT"}

	if !containsString(slice, "go") {
		t.Error("Should find 'go'")
	}
	if !containsString(slice, "python") { // Case insensitive
		t.Error("Should find 'python'")
	}
	if containsString(slice, "rust") {
		t.Error("Should not find 'rust'")
	}
}
```

---

## Resumen de la Iteracion 08

### Commits:
1. `feat(rules): add rule struct definitions`
2. `feat(rules): add yaml rule loader`
3. `feat(rules): add rule filtering by language`
4. `chore(rules): add base rules yaml`
5. `test(rules): add loader and filter tests`

### Archivos:
```
goreview/internal/rules/
├── types.go
├── loader.go
├── filter.go
├── rules_test.go
└── defaults/
    └── base.yaml
```

---

## Siguiente Iteracion

Continua con: **[09-COMANDO-REVIEW.md](09-COMANDO-REVIEW.md)**
