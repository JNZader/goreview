package git

import (
	"testing"
)

func TestParseDiffOptimized_MatchesOriginal(t *testing.T) {
	testCases := []struct {
		name string
		diff string
	}{
		{"empty", ""},
		{"small", generateDiff(1, 1, 10)},
		{"medium", generateDiff(5, 3, 20)},
		{"large", generateDiff(10, 5, 30)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			original, err1 := ParseDiff(tc.diff)
			optimized, err2 := ParseDiffOptimized(tc.diff)

			if err1 != nil || err2 != nil {
				t.Fatalf("errors: original=%v, optimized=%v", err1, err2)
			}

			if len(original.Files) != len(optimized.Files) {
				t.Fatalf("file count mismatch: original=%d, optimized=%d",
					len(original.Files), len(optimized.Files))
			}

			for i := range original.Files {
				origFile := &original.Files[i]
				optFile := &optimized.Files[i]

				if origFile.Path != optFile.Path {
					t.Errorf("file %d path mismatch: original=%q, optimized=%q",
						i, origFile.Path, optFile.Path)
				}
				if origFile.OldPath != optFile.OldPath {
					t.Errorf("file %d old path mismatch: original=%q, optimized=%q",
						i, origFile.OldPath, optFile.OldPath)
				}
				if origFile.Status != optFile.Status {
					t.Errorf("file %d status mismatch: original=%v, optimized=%v",
						i, origFile.Status, optFile.Status)
				}
				if origFile.Language != optFile.Language {
					t.Errorf("file %d language mismatch: original=%q, optimized=%q",
						i, origFile.Language, optFile.Language)
				}
				if origFile.Additions != optFile.Additions {
					t.Errorf("file %d additions mismatch: original=%d, optimized=%d",
						i, origFile.Additions, optFile.Additions)
				}
				if origFile.Deletions != optFile.Deletions {
					t.Errorf("file %d deletions mismatch: original=%d, optimized=%d",
						i, origFile.Deletions, optFile.Deletions)
				}
				if len(origFile.Hunks) != len(optFile.Hunks) {
					t.Errorf("file %d hunk count mismatch: original=%d, optimized=%d",
						i, len(origFile.Hunks), len(optFile.Hunks))
				}

				for j := range origFile.Hunks {
					if j >= len(optFile.Hunks) {
						break
					}
					origHunk := &origFile.Hunks[j]
					optHunk := &optFile.Hunks[j]

					if origHunk.OldStart != optHunk.OldStart {
						t.Errorf("file %d hunk %d OldStart mismatch", i, j)
					}
					if origHunk.OldLines != optHunk.OldLines {
						t.Errorf("file %d hunk %d OldLines mismatch", i, j)
					}
					if origHunk.NewStart != optHunk.NewStart {
						t.Errorf("file %d hunk %d NewStart mismatch", i, j)
					}
					if origHunk.NewLines != optHunk.NewLines {
						t.Errorf("file %d hunk %d NewLines mismatch", i, j)
					}
					if len(origHunk.Lines) != len(optHunk.Lines) {
						t.Errorf("file %d hunk %d line count mismatch: original=%d, optimized=%d",
							i, j, len(origHunk.Lines), len(optHunk.Lines))
					}
				}
			}
		})
	}
}

func TestParseDiffGitLine(t *testing.T) {
	tests := []struct {
		line    string
		oldPath string
		newPath string
	}{
		{"diff --git a/file.go b/file.go", "file.go", "file.go"},
		{"diff --git a/old.txt b/new.txt", "old.txt", "new.txt"},
		{"diff --git a/src/main.go b/src/main.go", "src/main.go", "src/main.go"},
	}

	for _, tc := range tests {
		oldPath, newPath := parseDiffGitLine(tc.line)
		if oldPath != tc.oldPath {
			t.Errorf("parseDiffGitLine(%q): oldPath = %q, want %q", tc.line, oldPath, tc.oldPath)
		}
		if newPath != tc.newPath {
			t.Errorf("parseDiffGitLine(%q): newPath = %q, want %q", tc.line, newPath, tc.newPath)
		}
	}
}

func TestParseHunkHeaderOptimized(t *testing.T) {
	tests := []struct {
		line     string
		oldStart int
		oldLines int
		newStart int
		newLines int
	}{
		{"@@ -1,10 +1,12 @@", 1, 10, 1, 12},
		{"@@ -5 +5 @@", 5, 1, 5, 1},
		{"@@ -100,50 +100,55 @@ func example()", 100, 50, 100, 55},
	}

	for _, tc := range tests {
		hunk := parseHunkHeaderOptimized(tc.line)
		if hunk.OldStart != tc.oldStart {
			t.Errorf("parseHunkHeaderOptimized(%q): OldStart = %d, want %d", tc.line, hunk.OldStart, tc.oldStart)
		}
		if hunk.OldLines != tc.oldLines {
			t.Errorf("parseHunkHeaderOptimized(%q): OldLines = %d, want %d", tc.line, hunk.OldLines, tc.oldLines)
		}
		if hunk.NewStart != tc.newStart {
			t.Errorf("parseHunkHeaderOptimized(%q): NewStart = %d, want %d", tc.line, hunk.NewStart, tc.newStart)
		}
		if hunk.NewLines != tc.newLines {
			t.Errorf("parseHunkHeaderOptimized(%q): NewLines = %d, want %d", tc.line, hunk.NewLines, tc.newLines)
		}
	}
}

func TestDetectLanguageOptimized(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "go"},
		{"app.py", "python"},
		{"index.ts", "typescript"},
		{"script.js", "javascript"},
		{"Main.java", "java"},
		{"lib.rs", "rust"},
		{"config.yaml", "yaml"},
		{"data.json", "json"},
		{"README.md", "markdown"},
		{"unknown.xyz", "unknown"},
		{"noext", "unknown"},
	}

	for _, tc := range tests {
		got := detectLanguageOptimized(tc.path)
		if got != tc.want {
			t.Errorf("detectLanguageOptimized(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}

// Benchmark comparison
func BenchmarkParseDiff_Original_Medium(b *testing.B) {
	diff := generateDiff(5, 3, 20)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDiff_Optimized_Medium(b *testing.B) {
	diff := generateDiff(5, 3, 20)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiffOptimized(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDiff_Original_Large(b *testing.B) {
	diff := generateDiff(20, 10, 50)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseDiff_Optimized_Large(b *testing.B) {
	diff := generateDiff(20, 10, 50)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiffOptimized(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDetectLanguage_Original(b *testing.B) {
	paths := []string{"main.go", "app.ts", "script.py", "lib.rs", "unknown.xyz"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = detectLanguage(path)
		}
	}
}

func BenchmarkDetectLanguage_Optimized(b *testing.B) {
	paths := []string{"main.go", "app.ts", "script.py", "lib.rs", "unknown.xyz"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = detectLanguageOptimized(path)
		}
	}
}
