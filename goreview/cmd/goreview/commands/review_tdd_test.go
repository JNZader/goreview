package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsTestFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		// Go tests
		{"internal/review/engine_test.go", true},
		{"internal/review/engine.go", false},

		// JavaScript/TypeScript tests
		{"src/utils.test.js", true},
		{"src/utils.spec.ts", true},
		{"src/utils.test.tsx", true},
		{"src/utils.js", false},

		// Test directories
		{"__tests__/utils.js", true},
		{"tests/helper.go", true},
		{"spec/auth_spec.rb", true},

		// Regular files
		{"main.go", false},
		{"README.md", false},
		{"package.json", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isTestFile(tt.path)
			if result != tt.expected {
				t.Errorf("isTestFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestGetTestPathVariants(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{
			"internal/review/engine.go",
			[]string{"internal/review/engine_test.go"},
		},
		{
			"src/utils.js",
			[]string{
				"src/utils.test.js",
				"src/utils.spec.js",
				"src/__tests__/utils.js",
				"src/utils.test.ts",
				"src/utils.spec.ts",
			},
		},
		{
			"src/component.tsx",
			[]string{
				"src/component.test.tsx",
				"src/component.spec.tsx",
				"src/__tests__/component.tsx",
			},
		},
		{
			"app/models/user.py",
			[]string{
				"app/models/test_user.py",
				"app/models/user_test.py",
				filepath.Join("app/models", "tests", "test_user.py"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getTestPathVariants(tt.path)

			// Check that expected paths are in the result
			for _, exp := range tt.expected {
				found := false
				for _, r := range result {
					// Normalize paths for comparison
					if filepath.Clean(r) == filepath.Clean(exp) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("getTestPathVariants(%q) missing expected path %q, got %v", tt.path, exp, result)
				}
			}
		})
	}
}

func TestHasCorrespondingTest(t *testing.T) {
	// Create a temp directory with test files
	tmpDir := t.TempDir()

	// Create source and test files
	srcFile := filepath.Join(tmpDir, "utils.go")
	testFile := filepath.Join(tmpDir, "utils_test.go")

	if err := os.WriteFile(srcFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(testFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test with test file
	if !hasCorrespondingTest(srcFile) {
		t.Errorf("hasCorrespondingTest(%q) = false, want true", srcFile)
	}

	// Create file without test
	noTestFile := filepath.Join(tmpDir, "standalone.go")
	if err := os.WriteFile(noTestFile, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	if hasCorrespondingTest(noTestFile) {
		t.Errorf("hasCorrespondingTest(%q) = true, want false", noTestFile)
	}
}

func TestGetExpectedTestPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"internal/review/engine.go", "internal/review/engine_test.go"},
		{"src/utils.js", "src/utils.test.js"},
		{"app/models/user.py", "app/models/test_user.py"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := getExpectedTestPath(tt.path)
			if filepath.Clean(result) != filepath.Clean(tt.expected) {
				t.Errorf("getExpectedTestPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}
