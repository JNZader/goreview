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
