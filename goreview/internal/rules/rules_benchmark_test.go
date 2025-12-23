// Package rules provides rule engine benchmarks
package rules

import (
	"fmt"
	"testing"
)

// createTestRules generates rules for benchmarking
func createTestRules(count int) []Rule {
	rules := make([]Rule, count)
	categories := []Category{CategorySecurity, CategoryPerformance, CategoryBestPractice, CategoryStyle, CategoryBug}
	severities := []Severity{SeverityInfo, SeverityWarning, SeverityError, SeverityCritical}

	for i := 0; i < count; i++ {
		rules[i] = Rule{
			ID:          fmt.Sprintf("RULE-%03d", i),
			Name:        fmt.Sprintf("Test Rule %d", i),
			Description: fmt.Sprintf("Description for rule %d with some text", i),
			Category:    categories[i%len(categories)],
			Severity:    severities[i%len(severities)],
			Languages:   []string{"go", "python", "javascript"}[0 : (i%3)+1],
			Patterns:    []string{fmt.Sprintf("*.%s", []string{"go", "py", "js"}[i%3])},
			Enabled:     i%5 != 0, // 80% enabled
			Message:     fmt.Sprintf("Issue message for rule %d", i),
			Suggestion:  fmt.Sprintf("Fix suggestion for rule %d", i),
		}
	}

	return rules
}

// BenchmarkFilter_Small measures filtering with few rules
func BenchmarkFilter_Small(b *testing.B) {
	rules := createTestRules(10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Filter(rules, "go", "main.go")
	}
}

// BenchmarkFilter_Medium measures filtering with moderate rules
func BenchmarkFilter_Medium(b *testing.B) {
	rules := createTestRules(50)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Filter(rules, "python", "script.py")
	}
}

// BenchmarkFilter_Large measures filtering with many rules
func BenchmarkFilter_Large(b *testing.B) {
	rules := createTestRules(200)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = Filter(rules, "javascript", "app.js")
	}
}

// BenchmarkFilter_Allocs tracks memory allocations
func BenchmarkFilter_Allocs(b *testing.B) {
	rules := createTestRules(50)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = Filter(rules, "go", "handler.go")
	}
}

// BenchmarkApplyPreset measures preset application
func BenchmarkApplyPreset(b *testing.B) {
	rules := createTestRules(100)
	preset := &Preset{
		Name:     "test",
		Includes: []string{"RULE-001", "RULE-010", "RULE-020", "RULE-030", "RULE-040"},
		Excludes: []string{"RULE-050"},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = ApplyPreset(rules, preset)
	}
}

// BenchmarkGetRulesByCategory measures category filtering
func BenchmarkGetRulesByCategory(b *testing.B) {
	rules := createTestRules(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetRulesByCategory(rules, CategorySecurity)
	}
}

// BenchmarkGetRulesBySeverity measures severity filtering
func BenchmarkGetRulesBySeverity(b *testing.B) {
	rules := createTestRules(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetRulesBySeverity(rules, SeverityWarning)
	}
}

// BenchmarkContainsString measures string search
func BenchmarkContainsString(b *testing.B) {
	slice := make([]string, 100)
	for i := 0; i < 100; i++ {
		slice[i] = fmt.Sprintf("item-%d", i)
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Search for various items
		_ = containsString(slice, "item-50")
		_ = containsString(slice, "item-99")
		_ = containsString(slice, "nonexistent")
	}
}

// BenchmarkMatchesAnyPattern measures pattern matching
func BenchmarkMatchesAnyPattern(b *testing.B) {
	patterns := []string{"*.go", "*.ts", "*.py", "*_test.go", "internal/*"}
	paths := []string{
		"main.go",
		"app.ts",
		"script.py",
		"handler_test.go",
		"internal/cache.go",
		"noMatch.txt",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = matchesAnyPattern(patterns, path)
		}
	}
}

// BenchmarkComplexFilter simulates real-world filtering
func BenchmarkComplexFilter(b *testing.B) {
	rules := createTestRules(100)
	preset := &Preset{
		Name:     "standard",
		Includes: []string{}, // Include all
		Excludes: []string{"RULE-010", "RULE-020", "RULE-030"},
	}

	// Pre-apply preset
	filteredByPreset := ApplyPreset(rules, preset)

	languages := []string{"go", "python", "javascript", "typescript", "rust"}
	files := []string{"main.go", "app.py", "index.js", "lib.ts", "core.rs"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lang := languages[i%len(languages)]
		file := files[i%len(files)]
		filtered := Filter(filteredByPreset, lang, file)

		// Apply severity filter
		_ = GetRulesBySeverity(filtered, SeverityWarning)
	}
}

// BenchmarkLoaderPreset measures preset loading
func BenchmarkLoaderPreset(b *testing.B) {
	loader := NewLoader("")
	presets := []string{"standard", "strict", "minimal"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		preset := presets[i%len(presets)]
		_, _ = loader.LoadPreset(preset)
	}
}
