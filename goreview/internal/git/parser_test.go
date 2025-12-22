package git

import (
	"testing"
)

func TestParseDiff(t *testing.T) {
	diffText := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,6 @@
 package main

+import "fmt"
+
 func main() {
-    println("hello")
+    fmt.Println("hello")
 }
`

	diff, err := ParseDiff(diffText)
	if err != nil {
		t.Fatalf("ParseDiff() error = %v", err)
	}

	if len(diff.Files) != 1 {
		t.Errorf("len(Files) = %d, want 1", len(diff.Files))
	}

	file := diff.Files[0]
	if file.Path != "main.go" {
		t.Errorf("Path = %v, want main.go", file.Path)
	}

	if file.Language != "go" {
		t.Errorf("Language = %v, want go", file.Language)
	}

	if file.Additions != 3 {
		t.Errorf("Additions = %d, want 3", file.Additions)
	}

	if file.Deletions != 1 {
		t.Errorf("Deletions = %d, want 1", file.Deletions)
	}
}

func TestParseDiffNewFile(t *testing.T) {
	diffText := `diff --git a/new.go b/new.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/new.go
@@ -0,0 +1,3 @@
+package main
+
+func new() {}
`

	diff, err := ParseDiff(diffText)
	if err != nil {
		t.Fatalf("ParseDiff() error = %v", err)
	}

	if diff.Files[0].Status != FileAdded {
		t.Errorf("Status = %v, want added", diff.Files[0].Status)
	}
}

func TestParseDiffDeleted(t *testing.T) {
	diffText := `diff --git a/old.go b/old.go
deleted file mode 100644
index 1234567..0000000
--- a/old.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func old() {}
`

	diff, err := ParseDiff(diffText)
	if err != nil {
		t.Fatalf("ParseDiff() error = %v", err)
	}

	if diff.Files[0].Status != FileDeleted {
		t.Errorf("Status = %v, want deleted", diff.Files[0].Status)
	}
}

func TestDetectLanguage(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"main.go", "go"},
		{"script.py", "python"},
		{"app.ts", "typescript"},
		{"Component.tsx", "typescript"},
		{"style.css", "css"},
		{"unknown.xyz", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := detectLanguage(tt.path)
			if got != tt.want {
				t.Errorf("detectLanguage(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestEmptyDiff(t *testing.T) {
	diff, err := ParseDiff("")
	if err != nil {
		t.Fatalf("ParseDiff() error = %v", err)
	}

	if len(diff.Files) != 0 {
		t.Errorf("len(Files) = %d, want 0", len(diff.Files))
	}
}
