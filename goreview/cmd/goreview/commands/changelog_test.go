package commands

import (
	"testing"

	"github.com/JNZader/goreview/goreview/internal/git"
)

func TestParseConventionalCommitMsg(t *testing.T) {
	tests := []struct {
		name     string
		commit   git.Commit
		expected git.ConventionalCommit
	}{
		{
			name: "simple feat",
			commit: git.Commit{
				Hash:      "abc123",
				ShortHash: "abc",
				Subject:   "feat: add new feature",
				Author:    "Test Author",
				Date:      "2024-01-01",
			},
			expected: git.ConventionalCommit{
				Type:        "feat",
				Description: "add new feature",
				Hash:        "abc123",
				ShortHash:   "abc",
				Author:      "Test Author",
				Date:        "2024-01-01",
			},
		},
		{
			name: "feat with scope",
			commit: git.Commit{
				Hash:      "def456",
				ShortHash: "def",
				Subject:   "feat(api): add endpoint",
				Author:    "Test Author",
				Date:      "2024-01-02",
			},
			expected: git.ConventionalCommit{
				Type:        "feat",
				Scope:       "api",
				Description: "add endpoint",
				Hash:        "def456",
				ShortHash:   "def",
				Author:      "Test Author",
				Date:        "2024-01-02",
			},
		},
		{
			name: "breaking change with !",
			commit: git.Commit{
				Hash:      "ghi789",
				ShortHash: "ghi",
				Subject:   "feat(auth)!: change login flow",
				Author:    "Test Author",
				Date:      "2024-01-03",
			},
			expected: git.ConventionalCommit{
				Type:        "feat",
				Scope:       "auth",
				Breaking:    true,
				Description: "change login flow",
				Hash:        "ghi789",
				ShortHash:   "ghi",
				Author:      "Test Author",
				Date:        "2024-01-03",
			},
		},
		{
			name: "fix commit",
			commit: git.Commit{
				Hash:      "jkl012",
				ShortHash: "jkl",
				Subject:   "fix: resolve bug",
				Author:    "Test Author",
				Date:      "2024-01-04",
			},
			expected: git.ConventionalCommit{
				Type:        "fix",
				Description: "resolve bug",
				Hash:        "jkl012",
				ShortHash:   "jkl",
				Author:      "Test Author",
				Date:        "2024-01-04",
			},
		},
		{
			name: "non-conventional commit",
			commit: git.Commit{
				Hash:      "mno345",
				ShortHash: "mno",
				Subject:   "Update README",
				Author:    "Test Author",
				Date:      "2024-01-05",
			},
			expected: git.ConventionalCommit{
				Type:        "other",
				Description: "Update README",
				Hash:        "mno345",
				ShortHash:   "mno",
				Author:      "Test Author",
				Date:        "2024-01-05",
			},
		},
		{
			name: "breaking change in body",
			commit: git.Commit{
				Hash:      "pqr678",
				ShortHash: "pqr",
				Subject:   "feat: major update",
				Body:      "BREAKING CHANGE: API has changed",
				Author:    "Test Author",
				Date:      "2024-01-06",
			},
			expected: git.ConventionalCommit{
				Type:        "feat",
				Breaking:    true,
				Description: "major update",
				Body:        "BREAKING CHANGE: API has changed",
				Hash:        "pqr678",
				ShortHash:   "pqr",
				Author:      "Test Author",
				Date:        "2024-01-06",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseConventionalCommitMsg(tt.commit)

			if result.Type != tt.expected.Type {
				t.Errorf("Type: got %q, want %q", result.Type, tt.expected.Type)
			}
			if result.Scope != tt.expected.Scope {
				t.Errorf("Scope: got %q, want %q", result.Scope, tt.expected.Scope)
			}
			if result.Breaking != tt.expected.Breaking {
				t.Errorf("Breaking: got %v, want %v", result.Breaking, tt.expected.Breaking)
			}
			if result.Description != tt.expected.Description {
				t.Errorf("Description: got %q, want %q", result.Description, tt.expected.Description)
			}
			if result.Hash != tt.expected.Hash {
				t.Errorf("Hash: got %q, want %q", result.Hash, tt.expected.Hash)
			}
		})
	}
}

func TestGroupCommitsByType(t *testing.T) {
	commits := []git.Commit{
		{Hash: "1", ShortHash: "1", Subject: "feat: feature 1"},
		{Hash: "2", ShortHash: "2", Subject: "feat: feature 2"},
		{Hash: "3", ShortHash: "3", Subject: "fix: bug fix"},
		{Hash: "4", ShortHash: "4", Subject: "docs: update readme"},
		{Hash: "5", ShortHash: "5", Subject: "Random commit"},
	}

	grouped := groupCommitsByType(commits)

	if len(grouped["feat"]) != 2 {
		t.Errorf("feat count: got %d, want 2", len(grouped["feat"]))
	}
	if len(grouped["fix"]) != 1 {
		t.Errorf("fix count: got %d, want 1", len(grouped["fix"]))
	}
	if len(grouped["docs"]) != 1 {
		t.Errorf("docs count: got %d, want 1", len(grouped["docs"]))
	}
	if len(grouped["other"]) != 1 {
		t.Errorf("other count: got %d, want 1", len(grouped["other"]))
	}
}

func TestGenerateChangelog(t *testing.T) {
	grouped := map[string][]git.ConventionalCommit{
		"feat": {
			{Type: "feat", Description: "add feature", ShortHash: "abc"},
			{Type: "feat", Scope: "api", Description: "add endpoint", ShortHash: "def"},
		},
		"fix": {
			{Type: "fix", Description: "fix bug", ShortHash: "ghi"},
		},
	}

	opts := changelogOptions{
		Version:  "v1.0.0",
		NoDate:   true,
		NoLinks:  false,
		NoHeader: false,
	}

	changelog := generateChangelog(grouped, opts)

	// Check header
	if !contains(changelog, "## v1.0.0") {
		t.Error("changelog should contain version header")
	}

	// Check sections
	if !contains(changelog, "### Features") {
		t.Error("changelog should contain Features section")
	}
	if !contains(changelog, "### Bug Fixes") {
		t.Error("changelog should contain Bug Fixes section")
	}

	// Check commits
	if !contains(changelog, "add feature") {
		t.Error("changelog should contain feature description")
	}
	if !contains(changelog, "**api:**") {
		t.Error("changelog should contain scope")
	}
	if !contains(changelog, "(abc)") {
		t.Error("changelog should contain commit hash")
	}
}

func TestGenerateChangelogNoLinks(t *testing.T) {
	grouped := map[string][]git.ConventionalCommit{
		"feat": {
			{Type: "feat", Description: "add feature", ShortHash: "abc"},
		},
	}

	opts := changelogOptions{
		NoLinks: true,
		NoDate:  true,
	}

	changelog := generateChangelog(grouped, opts)

	if contains(changelog, "(abc)") {
		t.Error("changelog should not contain commit hash when NoLinks is true")
	}
}

func TestGenerateChangelogBreakingChanges(t *testing.T) {
	grouped := map[string][]git.ConventionalCommit{
		"feat": {
			{Type: "feat", Description: "normal feature", ShortHash: "abc"},
			{Type: "feat", Description: "breaking feature", ShortHash: "def", Breaking: true},
		},
	}

	opts := changelogOptions{
		NoDate: true,
	}

	changelog := generateChangelog(grouped, opts)

	if !contains(changelog, "### BREAKING CHANGES") {
		t.Error("changelog should contain BREAKING CHANGES section")
	}
	if !contains(changelog, "breaking feature") {
		t.Error("changelog should contain breaking change description")
	}
}

func TestCollectBreakingChanges(t *testing.T) {
	grouped := map[string][]git.ConventionalCommit{
		"feat": {
			{Type: "feat", Description: "normal", Breaking: false},
			{Type: "feat", Description: "breaking", Breaking: true},
		},
		"fix": {
			{Type: "fix", Description: "also breaking", Breaking: true},
		},
	}

	breaking := collectBreakingChanges(grouped)

	if len(breaking) != 2 {
		t.Errorf("breaking count: got %d, want 2", len(breaking))
	}
}

func TestFilterNonBreaking(t *testing.T) {
	commits := []git.ConventionalCommit{
		{Description: "normal 1", Breaking: false},
		{Description: "breaking", Breaking: true},
		{Description: "normal 2", Breaking: false},
	}

	nonBreaking := filterNonBreaking(commits)

	if len(nonBreaking) != 2 {
		t.Errorf("non-breaking count: got %d, want 2", len(nonBreaking))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
