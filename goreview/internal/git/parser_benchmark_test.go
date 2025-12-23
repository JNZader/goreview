// Package git provides diff parser benchmarks
package git

import (
	"fmt"
	"strings"
	"testing"
)

// generateDiff creates a synthetic diff for benchmarking
func generateDiff(files, hunksPerFile, linesPerHunk int) string {
	var sb strings.Builder

	for f := 0; f < files; f++ {
		sb.WriteString(fmt.Sprintf("diff --git a/file%d.go b/file%d.go\n", f, f))
		sb.WriteString("index abc123..def456 100644\n")
		sb.WriteString(fmt.Sprintf("--- a/file%d.go\n", f))
		sb.WriteString(fmt.Sprintf("+++ b/file%d.go\n", f))

		for h := 0; h < hunksPerFile; h++ {
			startLine := h*linesPerHunk + 1
			sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n",
				startLine, linesPerHunk, startLine, linesPerHunk))

			for l := 0; l < linesPerHunk; l++ {
				switch l % 3 {
				case 0:
					sb.WriteString(fmt.Sprintf("+func added%d() {}\n", l))
				case 1:
					sb.WriteString(fmt.Sprintf("-func removed%d() {}\n", l))
				default:
					sb.WriteString(fmt.Sprintf(" func unchanged%d() {}\n", l))
				}
			}
		}
	}

	return sb.String()
}

// BenchmarkParseDiff_Small measures parsing of small diff
// 1 file, 1 hunk, 10 lines
func BenchmarkParseDiff_Small(b *testing.B) {
	diff := generateDiff(1, 1, 10)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseDiff_Medium measures parsing of medium diff
// 5 files, 3 hunks, 20 lines
func BenchmarkParseDiff_Medium(b *testing.B) {
	diff := generateDiff(5, 3, 20)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseDiff_Large measures parsing of large diff
// 20 files, 10 hunks, 50 lines
func BenchmarkParseDiff_Large(b *testing.B) {
	diff := generateDiff(20, 10, 50)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseDiff_Allocs tracks memory allocations
func BenchmarkParseDiff_Allocs(b *testing.B) {
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

// BenchmarkDetectLanguage measures language detection
func BenchmarkDetectLanguage(b *testing.B) {
	paths := []string{
		"main.go",
		"app.ts",
		"script.py",
		"lib.rs",
		"unknown.xyz",
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			_ = detectLanguage(path)
		}
	}
}

// BenchmarkParseDiff_RealWorld simulates a real-world PR diff
func BenchmarkParseDiff_RealWorld(b *testing.B) {
	// Simulate a typical PR: 10 files, varied sizes
	var sb strings.Builder

	for f := 0; f < 10; f++ {
		ext := []string{".go", ".ts", ".py", ".yaml", ".md"}[f%5]
		sb.WriteString(fmt.Sprintf("diff --git a/src/file%d%s b/src/file%d%s\n", f, ext, f, ext))
		sb.WriteString("index abc123..def456 100644\n")
		sb.WriteString(fmt.Sprintf("--- a/src/file%d%s\n", f, ext))
		sb.WriteString(fmt.Sprintf("+++ b/src/file%d%s\n", f, ext))

		// Varied hunk count
		hunks := (f % 4) + 1
		for h := 0; h < hunks; h++ {
			lines := (h+1)*5 + 10
			sb.WriteString(fmt.Sprintf("@@ -%d,%d +%d,%d @@\n", h*20+1, lines, h*20+1, lines))

			for l := 0; l < lines; l++ {
				switch l % 4 {
				case 0:
					sb.WriteString("+// Added line\n")
				case 1:
					sb.WriteString("-// Removed line\n")
				default:
					sb.WriteString(" // Context line\n")
				}
			}
		}
	}

	diff := sb.String()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := ParseDiff(diff)
		if err != nil {
			b.Fatal(err)
		}
	}
}
