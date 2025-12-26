package providers

import (
	"strings"
	"testing"
)

func TestValidModes(t *testing.T) {
	modes := ValidModes()

	expected := []string{"default", "security", "perf", "clean", "docs", "tests"}
	if len(modes) != len(expected) {
		t.Errorf("expected %d modes, got %d", len(expected), len(modes))
	}

	for _, e := range expected {
		found := false
		for _, m := range modes {
			if m == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected mode %q not found", e)
		}
	}
}

func TestIsValidMode(t *testing.T) {
	tests := []struct {
		name     string
		mode     string
		expected bool
	}{
		{"valid default", "default", true},
		{"valid security", "security", true},
		{"valid perf", "perf", true},
		{"valid clean", "clean", true},
		{"valid docs", "docs", true},
		{"valid tests", "tests", true},
		{"invalid mode", "invalid", false},
		{"empty mode", "", false},
		{"case sensitive", "Security", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidMode(tt.mode)
			if result != tt.expected {
				t.Errorf("IsValidMode(%q) = %v, want %v", tt.mode, result, tt.expected)
			}
		})
	}
}

func TestGetModePrompt(t *testing.T) {
	tests := []struct {
		mode            string
		expectedContain string
	}{
		{"security", "SECURITY REVIEW MODE"},
		{"perf", "PERFORMANCE REVIEW MODE"},
		{"clean", "CLEAN CODE REVIEW MODE"},
		{"docs", "DOCUMENTATION REVIEW MODE"},
		{"tests", "TEST COVERAGE REVIEW MODE"},
		{"default", "Focus on:"},
		{"invalid", "Focus on:"}, // Should return default
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			prompt := GetModePrompt(tt.mode)
			if !strings.Contains(prompt, tt.expectedContain) {
				t.Errorf("GetModePrompt(%q) should contain %q", tt.mode, tt.expectedContain)
			}
		})
	}
}

func TestParseModes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []ReviewMode
	}{
		{"empty string", "", []ReviewMode{ModeDefault}},
		{"default", "default", []ReviewMode{ModeDefault}},
		{"single valid", "security", []ReviewMode{ModeSecurity}},
		{"multiple valid", "security,perf", []ReviewMode{ModeSecurity, ModePerformance}},
		{"with spaces", "security, perf", []ReviewMode{ModeSecurity, ModePerformance}},
		{"with invalid", "security,invalid,perf", []ReviewMode{ModeSecurity, ModePerformance}},
		{"all invalid", "invalid,unknown", []ReviewMode{ModeDefault}},
		{"three modes", "security,perf,clean", []ReviewMode{ModeSecurity, ModePerformance, ModeClean}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseModes(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("ParseModes(%q) returned %d modes, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, m := range result {
				if m != tt.expected[i] {
					t.Errorf("ParseModes(%q)[%d] = %q, want %q", tt.input, i, m, tt.expected[i])
				}
			}
		})
	}
}

func TestCombineModePrompts(t *testing.T) {
	tests := []struct {
		name            string
		modes           []ReviewMode
		expectedContain []string
	}{
		{
			"empty returns default",
			[]ReviewMode{},
			[]string{"Focus on:"},
		},
		{
			"single default",
			[]ReviewMode{ModeDefault},
			[]string{"Focus on:"},
		},
		{
			"single security",
			[]ReviewMode{ModeSecurity},
			[]string{"SECURITY REVIEW MODE"},
		},
		{
			"multiple modes",
			[]ReviewMode{ModeSecurity, ModePerformance},
			[]string{"MULTI-MODE REVIEW", "SECURITY REVIEW MODE", "PERFORMANCE REVIEW MODE"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CombineModePrompts(tt.modes)
			for _, expected := range tt.expectedContain {
				if !strings.Contains(result, expected) {
					t.Errorf("CombineModePrompts should contain %q", expected)
				}
			}
		})
	}
}

func TestModePromptsContainKeyContent(t *testing.T) {
	// Verify each mode prompt contains expected keywords
	tests := []struct {
		mode     ReviewMode
		keywords []string
	}{
		{ModeSecurity, []string{"injection", "authentication", "critical", "security"}},
		{ModePerformance, []string{"n+1", "memory", "algorithm", "caching"}},
		{ModeClean, []string{"solid", "dry", "naming", "code smell"}},
		{ModeDocs, []string{"documentation", "jsdoc", "godoc", "docstring"}},
		{ModeTests, []string{"test coverage", "edge case", "mocking", "assertion"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.mode), func(t *testing.T) {
			prompt := ModePrompts[tt.mode]
			promptLower := strings.ToLower(prompt)
			for _, kw := range tt.keywords {
				if !strings.Contains(promptLower, strings.ToLower(kw)) {
					t.Errorf("Mode %q prompt should contain keyword %q", tt.mode, kw)
				}
			}
		})
	}
}
