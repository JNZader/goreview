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
