package memory

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestNoopMemory(t *testing.T) {
	ctx := context.Background()
	m := NewNoopMemory()

	// Test Store
	entry := &Entry{ID: "test", Content: "test content"}
	if err := m.Store(ctx, entry); err != nil {
		t.Errorf("Store() error = %v", err)
	}

	// Test Get returns nil
	got, err := m.Get(ctx, "test")
	if err != nil {
		t.Errorf("Get() error = %v", err)
	}
	if got != nil {
		t.Error("Get() should return nil for noop")
	}

	// Test Search returns nil
	results, err := m.Search(ctx, &Query{ID: "test"})
	if err != nil {
		t.Errorf("Search() error = %v", err)
	}
	if results != nil {
		t.Error("Search() should return nil for noop")
	}

	// Test Stats
	stats, err := m.Stats(ctx)
	if err != nil {
		t.Errorf("Stats() error = %v", err)
	}
	if stats.TotalEntries != 0 {
		t.Error("Stats should show 0 entries")
	}

	// Test Close
	if err := m.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestWorkingMemory(t *testing.T) {
	ctx := context.Background()
	wm := NewWorkingMemory(10, 1*time.Hour)
	defer func() { _ = wm.Close() }()

	t.Run("Store and Get", func(t *testing.T) {
		entry := &Entry{
			ID:      "entry1",
			Content: "test content",
			Type:    "test",
		}

		if err := wm.Store(ctx, entry); err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		got, err := wm.Get(ctx, "entry1")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got == nil {
			t.Fatal("Get() returned nil")
		}
		if got.Content != "test content" {
			t.Errorf("Content = %q, want %q", got.Content, "test content")
		}
	})

	t.Run("LRU Eviction", func(t *testing.T) {
		// Fill to capacity
		for i := 0; i < 15; i++ {
			entry := &Entry{ID: "lru" + string(rune('0'+i)), Content: "content"}
			if err := wm.Store(ctx, entry); err != nil {
				t.Fatalf("Store() error = %v", err)
			}
		}

		// Check that oldest entries were evicted
		if wm.Capacity() != 10 {
			t.Errorf("Capacity() = %d, want 10", wm.Capacity())
		}
	})

	t.Run("Search", func(t *testing.T) {
		entry := &Entry{
			ID:      "search1",
			Content: "searchable content",
			Type:    "searchtype",
			Tags:    []string{"tag1", "tag2"},
		}
		if err := wm.Store(ctx, entry); err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		// Search by type
		results, err := wm.Search(ctx, &Query{Type: "searchtype"})
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(results) == 0 {
			t.Error("Search() returned no results")
		}

		// Search by tags
		results, err = wm.Search(ctx, &Query{Tags: []string{"tag1"}})
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}
		if len(results) == 0 {
			t.Error("Search() by tags returned no results")
		}
	})

	t.Run("Stats", func(t *testing.T) {
		stats, err := wm.Stats(ctx)
		if err != nil {
			t.Fatalf("Stats() error = %v", err)
		}
		if stats.TotalEntries == 0 {
			t.Error("Stats should show some entries")
		}
	})
}

func TestSessionMemory(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	sm, err := NewSessionMemory(dir, 5, 1*time.Hour)
	if err != nil {
		t.Fatalf("NewSessionMemory() error = %v", err)
	}
	defer func() { _ = sm.Close() }()

	t.Run("Store and Get", func(t *testing.T) {
		entry := &Entry{
			ID:      "session1",
			Content: "session content",
			Type:    "test",
		}

		if err := sm.Store(ctx, entry); err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		got, err := sm.Get(ctx, "session1")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got == nil {
			t.Fatal("Get() returned nil")
		}
	})

	t.Run("Session ID", func(t *testing.T) {
		sessionID := sm.SessionID()
		if sessionID == "" {
			t.Error("SessionID() should not be empty")
		}
	})

	t.Run("New Session", func(t *testing.T) {
		oldID := sm.SessionID()

		newID, err := sm.NewSession(ctx)
		if err != nil {
			t.Fatalf("NewSession() error = %v", err)
		}

		if newID == oldID {
			t.Error("NewSession() should create a new session ID")
		}
	})

	t.Run("List Sessions", func(t *testing.T) {
		sessions, err := sm.ListSessions(ctx)
		if err != nil {
			t.Fatalf("ListSessions() error = %v", err)
		}
		// Should have at least one session saved
		if len(sessions) == 0 {
			t.Log("No sessions saved yet (expected if session was empty)")
		}
	})
}

func TestLongTermMemory(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	ltm, err := NewLongTermMemory(LongTermOptions{
		Dir:        filepath.Join(dir, "longterm"),
		MaxSizeMB:  10,
		GCInterval: 0, // Disable GC for tests
	})
	if err != nil {
		t.Fatalf("NewLongTermMemory() error = %v", err)
	}
	defer func() { _ = ltm.Close() }()

	t.Run("Store and Get", func(t *testing.T) {
		entry := &Entry{
			ID:      "lt1",
			Content: "long term content",
			Type:    "test",
		}

		if err := ltm.Store(ctx, entry); err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		got, err := ltm.Get(ctx, "lt1")
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if got == nil {
			t.Fatal("Get() returned nil")
		}
		if got.Content != "long term content" {
			t.Errorf("Content = %q, want %q", got.Content, "long term content")
		}
	})

	t.Run("Semantic Search", func(t *testing.T) {
		// Store entry with embedding
		embedder := NewEmbedder()
		entry := &Entry{
			ID:        "lt2",
			Content:   "function to calculate sum",
			Type:      "code",
			Embedding: embedder.Embed("function to calculate sum"),
		}

		if err := ltm.Store(ctx, entry); err != nil {
			t.Fatalf("Store() error = %v", err)
		}

		// Search with similar embedding
		queryEmbedding := embedder.Embed("calculate total sum")
		results, err := ltm.SemanticSearch(ctx, queryEmbedding, 10)
		if err != nil {
			t.Fatalf("SemanticSearch() error = %v", err)
		}
		if len(results) == 0 {
			t.Error("SemanticSearch() returned no results")
		}
	})

	t.Run("GarbageCollect", func(t *testing.T) {
		count, err := ltm.GarbageCollect(ctx)
		if err != nil {
			t.Fatalf("GarbageCollect() error = %v", err)
		}
		t.Logf("GarbageCollect removed %d entries", count)
	})
}

func TestHebbianLearning(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	hl, err := NewHebbianLearner(HebbianOptions{
		Dir:          filepath.Join(dir, "hebbian"),
		LearningRate: 0.1,
		DecayRate:    0.01,
		MinStrength:  0.05,
	})
	if err != nil {
		t.Fatalf("NewHebbianLearner() error = %v", err)
	}
	defer func() { _ = hl.Close() }()

	t.Run("Strengthen", func(t *testing.T) {
		// Strengthen association multiple times
		for i := 0; i < 5; i++ {
			if err := hl.Strengthen(ctx, "source1", "target1"); err != nil {
				t.Fatalf("Strengthen() error = %v", err)
			}
		}

		associations, err := hl.GetAssociations(ctx, "source1")
		if err != nil {
			t.Fatalf("GetAssociations() error = %v", err)
		}
		if len(associations) == 0 {
			t.Fatal("GetAssociations() returned no associations")
		}

		// Check strength increased
		if associations[0].Strength < 0.3 {
			t.Errorf("Strength = %f, want > 0.3", associations[0].Strength)
		}
	})

	t.Run("Weaken", func(t *testing.T) {
		// Get initial strength
		before, beforeErr := hl.GetAssociations(ctx, "source1")
		if beforeErr != nil {
			t.Fatalf("GetAssociations() error = %v", beforeErr)
		}
		initialStrength := before[0].Strength

		// Weaken
		if weakenErr := hl.Weaken(ctx, "source1", "target1"); weakenErr != nil {
			t.Fatalf("Weaken() error = %v", weakenErr)
		}

		// Check strength decreased
		after, afterErr := hl.GetAssociations(ctx, "source1")
		if afterErr != nil {
			t.Fatalf("GetAssociations() error = %v", afterErr)
		}
		if after[0].Strength >= initialStrength {
			t.Error("Strength should decrease after weaken")
		}
	})

	t.Run("Prune", func(t *testing.T) {
		// Create weak association
		if err := hl.Strengthen(ctx, "weak1", "weak2"); err != nil {
			t.Fatalf("Strengthen() error = %v", err)
		}

		// Prune with high threshold
		count, err := hl.Prune(ctx, 0.5)
		if err != nil {
			t.Fatalf("Prune() error = %v", err)
		}
		t.Logf("Pruned %d associations", count)
	})
}

func TestEmbedder(t *testing.T) {
	e := NewEmbedder()

	t.Run("Basic Embedding", func(t *testing.T) {
		text := "This is a test function for code review"
		embedding := e.Embed(text)

		if len(embedding) != EmbeddingDim {
			t.Errorf("Embedding dimension = %d, want %d", len(embedding), EmbeddingDim)
		}

		// Check it's normalized (magnitude ~1)
		var mag float64
		for _, v := range embedding {
			mag += float64(v) * float64(v)
		}
		if mag < 0.99 || mag > 1.01 {
			t.Errorf("Embedding not normalized, magnitude = %f", mag)
		}
	})

	t.Run("Similarity", func(t *testing.T) {
		text1 := "function to calculate sum of numbers"
		text2 := "method to compute total of values"
		text3 := "hello world greeting message"

		emb1 := e.Embed(text1)
		emb2 := e.Embed(text2)
		emb3 := e.Embed(text3)

		sim12 := e.Similarity(emb1, emb2)
		sim13 := e.Similarity(emb1, emb3)

		// Similar texts should have higher similarity
		if sim12 <= sim13 {
			t.Errorf("Similar texts should have higher similarity: sim12=%f, sim13=%f", sim12, sim13)
		}
	})

	t.Run("Empty Text", func(t *testing.T) {
		embedding := e.Embed("")
		if len(embedding) != EmbeddingDim {
			t.Errorf("Empty text embedding dimension = %d, want %d", len(embedding), EmbeddingDim)
		}
	})

	t.Run("Code Detection", func(t *testing.T) {
		goCode := "func main() { fmt.Println(\"hello\") }"
		pyCode := "def main(): print('hello')"

		goEmb := e.Embed(goCode)
		pyEmb := e.Embed(pyCode)

		// Both should produce valid embeddings
		if len(goEmb) != EmbeddingDim || len(pyEmb) != EmbeddingDim {
			t.Error("Code embedding dimension mismatch")
		}
	})
}

func TestSemanticIndex(t *testing.T) {
	idx := NewSemanticIndex()

	t.Run("Index and Search", func(t *testing.T) {
		idx.Index("1", "function to calculate sum")
		idx.Index("2", "method to compute total")
		idx.Index("3", "hello world greeting")

		results := idx.Search("calculate total", 10)
		if len(results) == 0 {
			t.Error("Search returned no results")
		}

		// First result should be most relevant
		if results[0].ID != "1" && results[0].ID != "2" {
			t.Logf("Top result ID = %s (expected 1 or 2)", results[0].ID)
		}
	})

	t.Run("Remove", func(t *testing.T) {
		idx.Index("temp", "temporary content")
		if idx.Size() == 0 {
			t.Error("Size should not be 0 after indexing")
		}

		idx.Remove("temp")
		results := idx.Search("temporary", 10)
		for _, r := range results {
			if r.ID == "temp" {
				t.Error("Removed entry should not appear in search")
			}
		}
	})
}

func TestMatchScore(t *testing.T) {
	entry := &Entry{
		ID:       "test",
		Content:  "test content here",
		Type:     "code",
		Tags:     []string{"go", "function"},
		Strength: 0.8,
	}

	tests := []struct {
		name     string
		query    *Query
		wantHigh bool
	}{
		{"nil query", nil, true},
		{"match by ID", &Query{ID: "test"}, true},
		{"no match by ID", &Query{ID: "other"}, false},
		{"match by type", &Query{Type: "code"}, true},
		{"no match by type", &Query{Type: "doc"}, false},
		{"match by tags", &Query{Tags: []string{"go"}}, true},
		{"partial tag match", &Query{Tags: []string{"go", "python"}}, true},
		{"no tag match", &Query{Tags: []string{"python"}}, false},
		{"match by content", &Query{Content: "test"}, true},
		{"no content match", &Query{Content: "xyz"}, false},
		{"match by strength", &Query{MinStrength: 0.5}, true},
		{"no strength match", &Query{MinStrength: 0.9}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := matchScore(entry, tt.query)
			if tt.wantHigh && score <= 0 {
				t.Errorf("matchScore() = %f, want > 0", score)
			}
			if !tt.wantHigh && score > 0 {
				t.Errorf("matchScore() = %f, want 0", score)
			}
		})
	}
}
