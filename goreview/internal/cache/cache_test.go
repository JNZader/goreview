package cache

import (
	"testing"
	"time"

	"github.com/JNZader/goreview/goreview/internal/providers"
)

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(2, time.Hour)

	// Test Set and Get
	resp := &providers.ReviewResponse{Summary: "test"}
	if err := cache.Set("key1", resp); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	got, found, err := cache.Get("key1")
	if err != nil || !found {
		t.Fatalf("Get() = %v, %v, want found", got, err)
	}
	if got.Summary != "test" {
		t.Errorf("Summary = %v, want test", got.Summary)
	}

	// Test miss
	_, found, err = cache.Get("nonexistent")
	if err != nil {
		t.Errorf("Get(nonexistent) error = %v", err)
	}
	if found {
		t.Error("Get(nonexistent) found, want miss")
	}
}

func TestLRUEviction(t *testing.T) {
	cache := NewLRUCache(2, time.Hour)

	_ = cache.Set("key1", &providers.ReviewResponse{Summary: "1"})
	_ = cache.Set("key2", &providers.ReviewResponse{Summary: "2"})
	_ = cache.Set("key3", &providers.ReviewResponse{Summary: "3"}) // Evicts key1

	_, found, _ := cache.Get("key1")
	if found {
		t.Error("key1 should be evicted")
	}

	_, found, _ = cache.Get("key2")
	if !found {
		t.Error("key2 should exist")
	}
}

func TestLRUExpiration(t *testing.T) {
	cache := NewLRUCache(10, 10*time.Millisecond)

	_ = cache.Set("key1", &providers.ReviewResponse{Summary: "test"})

	time.Sleep(20 * time.Millisecond)

	_, found, _ := cache.Get("key1")
	if found {
		t.Error("key1 should be expired")
	}
}

func TestLRUClear(t *testing.T) {
	cache := NewLRUCache(10, time.Hour)

	_ = cache.Set("key1", &providers.ReviewResponse{Summary: "1"})
	_ = cache.Set("key2", &providers.ReviewResponse{Summary: "2"})

	_ = cache.Clear()

	stats := cache.Stats()
	if stats.Entries != 0 {
		t.Errorf("Entries after Clear() = %d, want 0", stats.Entries)
	}
}

func TestLRUStats(t *testing.T) {
	cache := NewLRUCache(10, time.Hour)

	_ = cache.Set("key1", &providers.ReviewResponse{Summary: "test"})
	_, _, _ = cache.Get("key1")        // hit
	_, _, _ = cache.Get("key1")        // hit
	_, _, _ = cache.Get("nonexistent") // miss

	stats := cache.Stats()
	if stats.Hits != 2 {
		t.Errorf("Hits = %d, want 2", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Misses = %d, want 1", stats.Misses)
	}
}

func TestFileCache(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, time.Hour)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	resp := &providers.ReviewResponse{Summary: "test"}
	if setErr := cache.Set("key1", resp); setErr != nil {
		t.Fatalf("Set() error = %v", setErr)
	}

	got, found, getErr := cache.Get("key1")
	if getErr != nil || !found {
		t.Fatalf("Get() = %v, %v, want found", got, getErr)
	}
	if got.Summary != "test" {
		t.Errorf("Summary = %v, want test", got.Summary)
	}
}

func TestFileCacheExpiration(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	_ = cache.Set("key1", &providers.ReviewResponse{Summary: "test"})

	time.Sleep(20 * time.Millisecond)

	_, found, _ := cache.Get("key1")
	if found {
		t.Error("key1 should be expired")
	}
}

func TestFileCacheClear(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, time.Hour)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	_ = cache.Set("key1", &providers.ReviewResponse{Summary: "1"})
	_ = cache.Set("key2", &providers.ReviewResponse{Summary: "2"})

	_ = cache.Clear()

	stats := cache.Stats()
	if stats.Entries != 0 {
		t.Errorf("Entries after Clear() = %d, want 0", stats.Entries)
	}
}

func TestComputeKey(t *testing.T) {
	req1 := &providers.ReviewRequest{Diff: "diff1", Language: "go"}
	req2 := &providers.ReviewRequest{Diff: "diff1", Language: "go"}
	req3 := &providers.ReviewRequest{Diff: "diff2", Language: "go"}

	key1 := ComputeKey(req1)
	key2 := ComputeKey(req2)
	key3 := ComputeKey(req3)

	if key1 != key2 {
		t.Error("Same request should have same key")
	}

	if key1 == key3 {
		t.Error("Different requests should have different keys")
	}

	// Key should be 64 chars (SHA-256 hex)
	if len(key1) != 64 {
		t.Errorf("Key length = %d, want 64", len(key1))
	}
}
