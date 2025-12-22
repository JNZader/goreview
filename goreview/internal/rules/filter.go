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

// GetRulesBySeverity returns rules at or above severity.
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
