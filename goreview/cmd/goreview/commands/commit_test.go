package commands

import (
	"testing"
)

func TestParseConventionalCommit(t *testing.T) {
	tests := []struct {
		input        string
		wantType     string
		wantScope    string
		wantDesc     string
		wantBreaking bool
	}{
		{
			input:     "feat(auth): add login endpoint",
			wantType:  "feat",
			wantScope: "auth",
			wantDesc:  "add login endpoint",
		},
		{
			input:     "fix: resolve memory leak",
			wantType:  "fix",
			wantScope: "",
			wantDesc:  "resolve memory leak",
		},
		{
			input:        "feat(api)!: change response format",
			wantType:     "feat",
			wantScope:    "api",
			wantDesc:     "change response format",
			wantBreaking: true,
		},
		{
			input:     "chore: update dependencies",
			wantType:  "chore",
			wantScope: "",
			wantDesc:  "update dependencies",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parts := parseConventionalCommit(tt.input)

			if parts.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", parts.Type, tt.wantType)
			}
			if parts.Scope != tt.wantScope {
				t.Errorf("Scope = %q, want %q", parts.Scope, tt.wantScope)
			}
			if parts.Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", parts.Description, tt.wantDesc)
			}
			if parts.Breaking != tt.wantBreaking {
				t.Errorf("Breaking = %v, want %v", parts.Breaking, tt.wantBreaking)
			}
		})
	}
}

func TestConventionalPartsString(t *testing.T) {
	tests := []struct {
		parts *conventionalParts
		want  string
	}{
		{
			parts: &conventionalParts{Type: "feat", Description: "add feature"},
			want:  "feat: add feature",
		},
		{
			parts: &conventionalParts{Type: "fix", Scope: "api", Description: "fix bug"},
			want:  "fix(api): fix bug",
		},
		{
			parts: &conventionalParts{Type: "feat", Scope: "core", Breaking: true, Description: "breaking change"},
			want:  "feat(core)!: breaking change",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.parts.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildFullMessage(t *testing.T) {
	tests := []struct {
		subject string
		body    string
		footer  string
		want    string
	}{
		{
			subject: "feat: add feature",
			want:    "feat: add feature",
		},
		{
			subject: "feat: add feature",
			body:    "This is the body",
			want:    "feat: add feature\n\nThis is the body",
		},
		{
			subject: "fix: fix bug",
			body:    "Detailed description",
			footer:  "Closes #123",
			want:    "fix: fix bug\n\nDetailed description\n\nCloses #123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.subject, func(t *testing.T) {
			got := buildFullMessage(tt.subject, tt.body, tt.footer)
			if got != tt.want {
				t.Errorf("buildFullMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	tests := []struct {
		s    string
		max  int
		want string
	}{
		{"short", 10, "short"},
		{"this is a long string", 10, "this is..."},
		{"exactly10!", 10, "exactly10!"},
	}

	for _, tt := range tests {
		t.Run(tt.s, func(t *testing.T) {
			got := truncate(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}
