// Package cache provides cache benchmarks
package cache

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JNZader/goreview/goreview/internal/providers"
)

// BenchmarkLRUCache_Get measures cache read performance
func BenchmarkLRUCache_Get(b *testing.B) {
	c := NewLRUCache(1000, time.Hour)

	// Pre-populate cache
	for i := 0; i < 500; i++ {
		key := fmt.Sprintf("key-%d", i)
		c.Set(key, &providers.ReviewResponse{
			Summary: fmt.Sprintf("Review %d", i),
			Score:   i % 100,
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%500)
		c.Get(key)
	}
}

// BenchmarkLRUCache_Set measures cache write performance
func BenchmarkLRUCache_Set(b *testing.B) {
	c := NewLRUCache(1000, time.Hour)
	response := &providers.ReviewResponse{
		Summary: "Test review",
		Score:   85.0,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		c.Set(key, response)
	}
}

// BenchmarkLRUCache_Concurrent measures concurrent access
func BenchmarkLRUCache_Concurrent(b *testing.B) {
	c := NewLRUCache(1000, time.Hour)

	// Pre-populate
	for i := 0; i < 500; i++ {
		c.Set(fmt.Sprintf("key-%d", i), &providers.ReviewResponse{
			Summary: fmt.Sprintf("Review %d", i),
		})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key-%d", i%500)
			if i%2 == 0 {
				c.Get(key)
			} else {
				c.Set(key, &providers.ReviewResponse{Summary: "new"})
			}
			i++
		}
	})
}

// BenchmarkLRUCache_WithEviction measures performance under eviction
func BenchmarkLRUCache_WithEviction(b *testing.B) {
	// Small cache to force eviction
	c := NewLRUCache(100, time.Hour)
	response := &providers.ReviewResponse{
		Summary: "Test review",
		Score:   85.0,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		c.Set(key, response)
	}
}

// BenchmarkComputeKey measures cache key computation
func BenchmarkComputeKey(b *testing.B) {
	req := &providers.ReviewRequest{
		Diff:     "diff content here with some text",
		FilePath: "src/main.go",
		Language: "go",
		Rules:    []string{"SEC-001", "PERF-001"},
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ComputeKey(req)
	}
}

// BenchmarkFileCache_Save measures file persistence
func BenchmarkFileCache_Save(b *testing.B) {
	ctx := context.Background()
	dir := b.TempDir()

	fc, err := NewFileCache(dir, time.Hour)
	if err != nil {
		b.Fatal(err)
	}

	response := &providers.ReviewResponse{
		Summary: "Test review with some content to simulate real data",
		Score:   85.0,
		Issues: []providers.Issue{
			{Type: "security", Severity: "warning", Message: "Potential issue"},
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i)
		if err := fc.Set(key, response); err != nil {
			b.Fatal(err)
		}
	}

	// Cleanup will be handled by TempDir
	_ = ctx
}

// BenchmarkFileCache_Load measures file loading
func BenchmarkFileCache_Load(b *testing.B) {
	ctx := context.Background()
	dir := b.TempDir()

	fc, err := NewFileCache(dir, time.Hour)
	if err != nil {
		b.Fatal(err)
	}

	// Pre-save entries
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key-%d", i)
		fc.Set(key, &providers.ReviewResponse{
			Summary: "Test review",
			Score:   i,
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key-%d", i%100)
		fc.Get(key)
	}

	_ = ctx
}

// BenchmarkLRUCache_Stats measures stats collection
func BenchmarkLRUCache_Stats(b *testing.B) {
	c := NewLRUCache(1000, time.Hour)

	// Add some entries
	for i := 0; i < 500; i++ {
		c.Set(fmt.Sprintf("key-%d", i), &providers.ReviewResponse{Summary: "test"})
	}

	// Simulate some hits/misses
	for i := 0; i < 1000; i++ {
		c.Get(fmt.Sprintf("key-%d", i%600))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = c.Stats()
	}
}

// BenchmarkLRUCache_MixedWorkload simulates real-world usage
func BenchmarkLRUCache_MixedWorkload(b *testing.B) {
	c := NewLRUCache(500, time.Hour)

	// Pre-populate with 250 entries
	for i := 0; i < 250; i++ {
		c.Set(fmt.Sprintf("existing-%d", i), &providers.ReviewResponse{
			Summary: fmt.Sprintf("Review %d", i),
		})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		switch i % 10 {
		case 0, 1, 2, 3, 4, 5: // 60% reads (hits)
			c.Get(fmt.Sprintf("existing-%d", i%250))
		case 6, 7: // 20% reads (misses)
			c.Get(fmt.Sprintf("nonexistent-%d", i))
		case 8, 9: // 20% writes
			c.Set(fmt.Sprintf("new-%d", i), &providers.ReviewResponse{
				Summary: "New review",
			})
		}
	}
}

func init() {
	// Ensure temp directories work
	_ = os.TempDir()
}
