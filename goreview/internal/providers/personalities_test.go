package providers

import (
	"testing"
)

func TestValidPersonalities(t *testing.T) {
	personalities := ValidPersonalities()

	expected := []string{"default", "senior", "strict", "friendly", "security-expert"}

	if len(personalities) != len(expected) {
		t.Errorf("expected %d personalities, got %d", len(expected), len(personalities))
	}

	for _, exp := range expected {
		found := false
		for _, p := range personalities {
			if p == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected personality %q not found", exp)
		}
	}
}

func TestIsValidPersonality(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"default is valid", "default", true},
		{"senior is valid", "senior", true},
		{"strict is valid", "strict", true},
		{"friendly is valid", "friendly", true},
		{"security-expert is valid", "security-expert", true},
		{"invalid personality", "aggressive", false},
		{"empty string", "", false},
		{"random string", "random", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidPersonality(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidPersonality(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetPersonalityPrompt(t *testing.T) {
	tests := []struct {
		name        string
		personality string
		contains    string
	}{
		{
			name:        "default personality",
			personality: "default",
			contains:    "expert code reviewer",
		},
		{
			name:        "senior personality",
			personality: "senior",
			contains:    "senior developer mentoring",
		},
		{
			name:        "strict personality",
			personality: "strict",
			contains:    "strict, demanding",
		},
		{
			name:        "friendly personality",
			personality: "friendly",
			contains:    "friendly and encouraging",
		},
		{
			name:        "security-expert personality",
			personality: "security-expert",
			contains:    "security-focused",
		},
		{
			name:        "unknown falls back to default",
			personality: "unknown",
			contains:    "expert code reviewer",
		},
		{
			name:        "empty falls back to default",
			personality: "",
			contains:    "expert code reviewer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := GetPersonalityPrompt(tt.personality)
			if !containsString(prompt, tt.contains) {
				t.Errorf("GetPersonalityPrompt(%q) should contain %q, got: %s", tt.personality, tt.contains, prompt)
			}
		})
	}
}

func TestPersonalityPromptsNotEmpty(t *testing.T) {
	for name, prompt := range PersonalityPrompts {
		if prompt == "" {
			t.Errorf("personality %q has empty prompt", name)
		}
		if len(prompt) < 50 {
			t.Errorf("personality %q prompt is too short (%d chars)", name, len(prompt))
		}
	}
}

func TestSecurityExpertMentionsOWASP(t *testing.T) {
	prompt := GetPersonalityPrompt("security-expert")
	if !containsString(prompt, "OWASP") {
		t.Error("security-expert personality should mention OWASP")
	}
}

func TestSeniorExplainsWhy(t *testing.T) {
	prompt := GetPersonalityPrompt("senior")
	if !containsString(prompt, "why") {
		t.Error("senior personality should emphasize explaining the 'why'")
	}
}

func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
