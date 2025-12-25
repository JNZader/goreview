package rules

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// InheritConfig configures rule inheritance.
type InheritConfig struct {
	// InheritFrom specifies sources to inherit rules from (URLs or local paths)
	InheritFrom []string `yaml:"inherit_from" mapstructure:"inherit_from"`

	// Override contains rule overrides for this level
	Override map[string]interface{} `yaml:"override" mapstructure:"override"`

	// Disable lists rule IDs to disable at this level
	Disable []string `yaml:"disable" mapstructure:"disable"`

	// Enable lists rule IDs to enable at this level
	Enable []string `yaml:"enable" mapstructure:"enable"`
}

// HierarchicalLoader loads and merges rules from multiple sources.
type HierarchicalLoader struct {
	baseLoader *Loader
	httpClient *http.Client
	cache      map[string][]Rule
}

// NewHierarchicalLoader creates a new hierarchical rule loader.
func NewHierarchicalLoader(rulesDir string) *HierarchicalLoader {
	return &HierarchicalLoader{
		baseLoader: NewLoader(rulesDir),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache: make(map[string][]Rule),
	}
}

// LoadWithInheritance loads rules with inheritance from parent sources.
func (hl *HierarchicalLoader) LoadWithInheritance(ctx context.Context, config InheritConfig) ([]Rule, error) {
	// Start with base rules
	baseRules, err := hl.baseLoader.Load()
	if err != nil {
		return nil, fmt.Errorf("loading base rules: %w", err)
	}

	// Create a map for easy rule lookup
	rulesMap := make(map[string]Rule)
	for _, r := range baseRules {
		rulesMap[r.ID] = r
	}

	// Load and merge rules from each parent source
	for _, source := range config.InheritFrom {
		parentRules, err := hl.loadFromSource(ctx, source)
		if err != nil {
			// Log warning but continue with other sources
			fmt.Fprintf(os.Stderr, "Warning: failed to load rules from %s: %v\n", source, err)
			continue
		}

		// Merge parent rules (parent rules take precedence for same ID)
		for _, r := range parentRules {
			rulesMap[r.ID] = r
		}
	}

	// Apply overrides
	rulesMap = hl.applyOverrides(rulesMap, config.Override)

	// Apply enable/disable
	rulesMap = hl.applyEnableDisable(rulesMap, config.Enable, config.Disable)

	// Convert back to slice
	result := make([]Rule, 0, len(rulesMap))
	for _, r := range rulesMap {
		result = append(result, r)
	}

	return result, nil
}

// loadFromSource loads rules from a URL or local file.
func (hl *HierarchicalLoader) loadFromSource(ctx context.Context, source string) ([]Rule, error) {
	// Check cache first
	if cached, ok := hl.cache[source]; ok {
		return cached, nil
	}

	var data []byte
	var err error

	if isURL(source) {
		data, err = hl.fetchFromURL(ctx, source)
	} else {
		data, err = hl.loadFromFile(source)
	}

	if err != nil {
		return nil, err
	}

	rules, err := hl.parseRulesData(data)
	if err != nil {
		return nil, err
	}

	// Cache the result
	hl.cache[source] = rules

	return rules, nil
}

// fetchFromURL fetches rules from a URL.
func (hl *HierarchicalLoader) fetchFromURL(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "GoReview/1.0")
	req.Header.Set("Accept", "application/yaml, text/yaml, application/x-yaml")

	resp, err := hl.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
}

// loadFromFile loads rules from a local file.
func (hl *HierarchicalLoader) loadFromFile(path string) ([]byte, error) {
	// Handle relative paths
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(cwd, path)
	}

	return os.ReadFile(path) //nolint:gosec // Path comes from config
}

// parseRulesData parses YAML rules data.
func (hl *HierarchicalLoader) parseRulesData(data []byte) ([]Rule, error) {
	var ruleSet RuleSet
	if err := yaml.Unmarshal(data, &ruleSet); err != nil {
		return nil, err
	}
	return ruleSet.Rules, nil
}

// applyOverrides applies rule property overrides.
func (hl *HierarchicalLoader) applyOverrides(rulesMap map[string]Rule, overrides map[string]interface{}) map[string]Rule {
	for key, value := range overrides {
		// Handle specific override patterns
		switch key {
		case "max_complexity":
			// This is a config override, not a rule override
			continue
		default:
			// Check if it's a rule ID
			if rule, ok := rulesMap[key]; ok {
				if overrideMap, isMap := value.(map[string]interface{}); isMap {
					rule = applyRuleOverride(rule, overrideMap)
					rulesMap[key] = rule
				}
			}
		}
	}
	return rulesMap
}

// applyRuleOverride applies overrides to a single rule.
func applyRuleOverride(rule Rule, overrides map[string]interface{}) Rule {
	if severity, ok := overrides["severity"].(string); ok {
		rule.Severity = Severity(severity)
	}
	if enabled, ok := overrides["enabled"].(bool); ok {
		rule.Enabled = enabled
	}
	if message, ok := overrides["message"].(string); ok {
		rule.Message = message
	}
	if suggestion, ok := overrides["suggestion"].(string); ok {
		rule.Suggestion = suggestion
	}
	if languages, ok := overrides["languages"].([]interface{}); ok {
		rule.Languages = make([]string, len(languages))
		for i, l := range languages {
			rule.Languages[i] = fmt.Sprintf("%v", l)
		}
	}
	if patterns, ok := overrides["patterns"].([]interface{}); ok {
		rule.Patterns = make([]string, len(patterns))
		for i, p := range patterns {
			rule.Patterns[i] = fmt.Sprintf("%v", p)
		}
	}
	return rule
}

// applyEnableDisable applies enable/disable lists.
func (hl *HierarchicalLoader) applyEnableDisable(rulesMap map[string]Rule, enable, disable []string) map[string]Rule {
	// Create lookup sets
	enableSet := make(map[string]bool)
	disableSet := make(map[string]bool)

	for _, id := range enable {
		enableSet[id] = true
	}
	for _, id := range disable {
		disableSet[id] = true
	}

	// Apply enable/disable
	for id, rule := range rulesMap {
		if enableSet[id] {
			rule.Enabled = true
			rulesMap[id] = rule
		}
		if disableSet[id] {
			rule.Enabled = false
			rulesMap[id] = rule
		}
	}

	return rulesMap
}

// isURL checks if a source string is a URL.
func isURL(source string) bool {
	return strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://")
}

// MergeRuleSets merges multiple rule sets with later sets taking precedence.
func MergeRuleSets(sets ...[]Rule) []Rule {
	rulesMap := make(map[string]Rule)

	for _, set := range sets {
		for _, rule := range set {
			rulesMap[rule.ID] = rule
		}
	}

	result := make([]Rule, 0, len(rulesMap))
	for _, rule := range rulesMap {
		result = append(result, rule)
	}

	return result
}

// InheritanceChain represents a chain of rule sources.
type InheritanceChain struct {
	Sources []InheritanceSource `yaml:"sources" json:"sources"`
}

// InheritanceSource represents a single source in the chain.
type InheritanceSource struct {
	Name     string `yaml:"name" json:"name"`
	Source   string `yaml:"source" json:"source"`
	Priority int    `yaml:"priority" json:"priority"` // Higher = more priority
	Type     string `yaml:"type" json:"type"`         // "url", "file", "embedded"
}

// LoadChain loads rules from an inheritance chain.
func (hl *HierarchicalLoader) LoadChain(ctx context.Context, chain InheritanceChain) ([]Rule, error) {
	// Sort by priority (lower first, so higher priority sources override)
	sources := chain.Sources
	for i := 0; i < len(sources)-1; i++ {
		for j := i + 1; j < len(sources); j++ {
			if sources[i].Priority > sources[j].Priority {
				sources[i], sources[j] = sources[j], sources[i]
			}
		}
	}

	rulesMap := make(map[string]Rule)

	for _, source := range sources {
		var rules []Rule
		var err error

		switch source.Type {
		case "embedded":
			rules, err = hl.baseLoader.loadEmbedded()
		case "url":
			rules, err = hl.loadFromSource(ctx, source.Source)
		case "file":
			rules, err = hl.loadFromSource(ctx, source.Source)
		default:
			rules, err = hl.loadFromSource(ctx, source.Source)
		}

		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to load rules from %s: %v\n", source.Name, err)
			continue
		}

		for _, r := range rules {
			rulesMap[r.ID] = r
		}
	}

	result := make([]Rule, 0, len(rulesMap))
	for _, r := range rulesMap {
		result = append(result, r)
	}

	return result, nil
}

// ValidateInheritConfig validates an inheritance configuration.
func ValidateInheritConfig(config InheritConfig) error {
	for _, source := range config.InheritFrom {
		if source == "" {
			return fmt.Errorf("empty source in inherit_from")
		}
		if isURL(source) && !strings.HasPrefix(source, "https://") {
			return fmt.Errorf("insecure URL (must use HTTPS): %s", source)
		}
	}
	return nil
}
