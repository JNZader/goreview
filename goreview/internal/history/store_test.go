package history

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewStore(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(StoreConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}
}

func TestStoreAndSearch(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(StoreConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Store a record
	record := &ReviewRecord{
		CommitHash:  "abc123",
		FilePath:    "src/auth/login.go",
		IssueType:   "security",
		Severity:    "critical",
		Message:     "Potential SQL injection vulnerability",
		Suggestion:  "Use parameterized queries",
		Line:        42,
		Author:      "john",
		Branch:      "main",
		CreatedAt:   time.Now(),
		Resolved:    false,
		ReviewRound: 1,
	}

	if err := store.Store(ctx, record); err != nil {
		t.Fatalf("Failed to store record: %v", err)
	}

	if record.ID == 0 {
		t.Error("Record ID was not set after insert")
	}

	// Search by text
	result, err := store.Search(ctx, SearchQuery{
		Text: "SQL injection",
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("Expected 1 result, got %d", result.TotalCount)
	}

	if len(result.Records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(result.Records))
	}

	if result.Records[0].FilePath != "src/auth/login.go" {
		t.Errorf("Expected file path src/auth/login.go, got %s", result.Records[0].FilePath)
	}
}

func TestSearchFilters(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(StoreConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Store multiple records
	records := []*ReviewRecord{
		{
			CommitHash:  "abc123",
			FilePath:    "src/auth/login.go",
			IssueType:   "security",
			Severity:    "critical",
			Message:     "Security issue in auth",
			Author:      "john",
			Branch:      "main",
			CreatedAt:   time.Now(),
			ReviewRound: 1,
		},
		{
			CommitHash:  "def456",
			FilePath:    "src/api/handler.go",
			IssueType:   "performance",
			Severity:    "warning",
			Message:     "Slow database query",
			Author:      "jane",
			Branch:      "feature",
			CreatedAt:   time.Now(),
			ReviewRound: 1,
		},
		{
			CommitHash:  "ghi789",
			FilePath:    "src/auth/logout.go",
			IssueType:   "bug",
			Severity:    "error",
			Message:     "Bug in logout flow",
			Author:      "john",
			Branch:      "main",
			CreatedAt:   time.Now(),
			Resolved:    true,
			ReviewRound: 2,
		},
	}

	if err := store.StoreBatch(ctx, records); err != nil {
		t.Fatalf("Failed to store batch: %v", err)
	}

	// Test severity filter
	t.Run("filter by severity", func(t *testing.T) {
		result, err := store.Search(ctx, SearchQuery{Severity: "critical"})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if result.TotalCount != 1 {
			t.Errorf("Expected 1 critical issue, got %d", result.TotalCount)
		}
	})

	// Test author filter
	t.Run("filter by author", func(t *testing.T) {
		result, err := store.Search(ctx, SearchQuery{Author: "john"})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if result.TotalCount != 2 {
			t.Errorf("Expected 2 issues by john, got %d", result.TotalCount)
		}
	})

	// Test file pattern filter
	t.Run("filter by file pattern", func(t *testing.T) {
		result, err := store.Search(ctx, SearchQuery{File: "src/auth/*"})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if result.TotalCount != 2 {
			t.Errorf("Expected 2 issues in auth dir, got %d", result.TotalCount)
		}
	})

	// Test resolved filter
	t.Run("filter by resolved", func(t *testing.T) {
		resolved := true
		result, err := store.Search(ctx, SearchQuery{Resolved: &resolved})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if result.TotalCount != 1 {
			t.Errorf("Expected 1 resolved issue, got %d", result.TotalCount)
		}
	})

	// Test type filter
	t.Run("filter by type", func(t *testing.T) {
		result, err := store.Search(ctx, SearchQuery{Type: "security"})
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}
		if result.TotalCount != 1 {
			t.Errorf("Expected 1 security issue, got %d", result.TotalCount)
		}
	})
}

func TestGetFileHistory(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(StoreConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Store records for a file
	records := []*ReviewRecord{
		{
			CommitHash:  "abc123",
			FilePath:    "src/auth/login.go",
			IssueType:   "security",
			Severity:    "critical",
			Message:     "Issue 1",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			ReviewRound: 1,
		},
		{
			CommitHash:  "def456",
			FilePath:    "src/auth/login.go",
			IssueType:   "bug",
			Severity:    "error",
			Message:     "Issue 2",
			CreatedAt:   time.Now(),
			Resolved:    true,
			ReviewRound: 2,
		},
	}

	if err := store.StoreBatch(ctx, records); err != nil {
		t.Fatalf("Failed to store batch: %v", err)
	}

	// Get file history
	hist, err := store.GetFileHistory(ctx, "src/auth/login.go")
	if err != nil {
		t.Fatalf("GetFileHistory failed: %v", err)
	}

	if hist.TotalIssues != 2 {
		t.Errorf("Expected 2 total issues, got %d", hist.TotalIssues)
	}

	if hist.Resolved != 1 {
		t.Errorf("Expected 1 resolved, got %d", hist.Resolved)
	}

	if hist.Pending != 1 {
		t.Errorf("Expected 1 pending, got %d", hist.Pending)
	}

	if hist.ReviewRounds != 2 {
		t.Errorf("Expected 2 review rounds, got %d", hist.ReviewRounds)
	}

	if hist.BySeverity["critical"] != 1 {
		t.Errorf("Expected 1 critical issue, got %d", hist.BySeverity["critical"])
	}

	if hist.BySeverity["error"] != 1 {
		t.Errorf("Expected 1 error issue, got %d", hist.BySeverity["error"])
	}
}

func TestGetStats(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(StoreConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Store some records
	records := []*ReviewRecord{
		{
			CommitHash:  "abc123",
			FilePath:    "file1.go",
			IssueType:   "security",
			Severity:    "critical",
			Message:     "Issue 1",
			CreatedAt:   time.Now(),
			ReviewRound: 1,
		},
		{
			CommitHash:  "def456",
			FilePath:    "file1.go",
			IssueType:   "bug",
			Severity:    "error",
			Message:     "Issue 2",
			CreatedAt:   time.Now(),
			Resolved:    true,
			ReviewRound: 1,
		},
		{
			CommitHash:  "ghi789",
			FilePath:    "file2.go",
			IssueType:   "performance",
			Severity:    "warning",
			Message:     "Issue 3",
			CreatedAt:   time.Now(),
			ReviewRound: 1,
		},
	}

	if err := store.StoreBatch(ctx, records); err != nil {
		t.Fatalf("Failed to store batch: %v", err)
	}

	stats, err := store.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.TotalIssues != 3 {
		t.Errorf("Expected 3 total issues, got %d", stats.TotalIssues)
	}

	if stats.ResolvedIssues != 1 {
		t.Errorf("Expected 1 resolved issue, got %d", stats.ResolvedIssues)
	}

	if stats.BySeverity["critical"] != 1 {
		t.Errorf("Expected 1 critical, got %d", stats.BySeverity["critical"])
	}

	if stats.ByFile["file1.go"] != 2 {
		t.Errorf("Expected 2 issues in file1.go, got %d", stats.ByFile["file1.go"])
	}
}

func TestMarkResolved(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	store, err := NewStore(StoreConfig{Path: dbPath})
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Store a record
	record := &ReviewRecord{
		CommitHash:  "abc123",
		FilePath:    "test.go",
		IssueType:   "bug",
		Severity:    "error",
		Message:     "Test issue",
		CreatedAt:   time.Now(),
		ReviewRound: 1,
	}

	if err := store.Store(ctx, record); err != nil {
		t.Fatalf("Failed to store record: %v", err)
	}

	// Mark as resolved
	if err := store.MarkResolved(ctx, record.ID); err != nil {
		t.Fatalf("MarkResolved failed: %v", err)
	}

	// Verify resolved
	resolved := true
	result, err := store.Search(ctx, SearchQuery{Resolved: &resolved})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if result.TotalCount != 1 {
		t.Errorf("Expected 1 resolved issue, got %d", result.TotalCount)
	}
}
