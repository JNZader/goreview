package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// WorkingMem implements WorkingMemory with LRU eviction and TTL support.
type WorkingMem struct {
	mu       sync.RWMutex
	entries  map[string]*Entry
	order    []string // LRU order (oldest first)
	capacity int
	ttl      time.Duration

	// Statistics
	hits   int64
	misses int64
}

// NewWorkingMemory creates a new working memory instance.
func NewWorkingMemory(capacity int, ttl time.Duration) *WorkingMem {
	if capacity <= 0 {
		capacity = 100
	}
	return &WorkingMem{
		entries:  make(map[string]*Entry),
		order:    make([]string, 0, capacity),
		capacity: capacity,
		ttl:      ttl,
	}
}

// Compile-time interface check.
var _ WorkingMemory = (*WorkingMem)(nil)

// Store saves an entry to working memory.
func (w *WorkingMem) Store(ctx context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	now := time.Now()
	entry.CreatedAt = now
	entry.UpdatedAt = now
	entry.AccessedAt = now

	// Set TTL if not specified
	if entry.TTL == 0 && w.ttl > 0 {
		entry.TTL = w.ttl
	}

	// Check if entry already exists
	if _, exists := w.entries[entry.ID]; exists {
		// Update existing entry
		w.entries[entry.ID] = entry
		w.touch(entry.ID)
		return nil
	}

	// Evict if at capacity
	for len(w.entries) >= w.capacity {
		w.evictOldest()
	}

	// Store new entry
	w.entries[entry.ID] = entry
	w.order = append(w.order, entry.ID)

	return nil
}

// Get retrieves an entry by ID.
func (w *WorkingMem) Get(ctx context.Context, id string) (*Entry, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry, exists := w.entries[id]
	if !exists {
		atomic.AddInt64(&w.misses, 1)
		return nil, nil
	}

	// Check TTL
	if w.isExpired(entry) {
		w.deleteEntry(id)
		atomic.AddInt64(&w.misses, 1)
		return nil, nil
	}

	// Update access time and move to end of LRU
	entry.AccessedAt = time.Now()
	entry.AccessCount++
	w.touch(id)

	atomic.AddInt64(&w.hits, 1)
	return entry, nil
}

// Search finds entries matching the query.
func (w *WorkingMem) Search(ctx context.Context, query *Query) ([]*SearchResult, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	var results []*SearchResult

	for _, entry := range w.entries {
		// Skip expired entries
		if w.isExpired(entry) {
			continue
		}

		score := w.matchScore(entry, query)
		if score > 0 {
			results = append(results, &SearchResult{
				Entry: entry,
				Score: score,
			})
		}
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit and offset
	if query.Offset > 0 && query.Offset < len(results) {
		results = results[query.Offset:]
	}
	if query.Limit > 0 && query.Limit < len(results) {
		results = results[:query.Limit]
	}

	return results, nil
}

// Update modifies an existing entry.
func (w *WorkingMem) Update(ctx context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if _, exists := w.entries[entry.ID]; !exists {
		return fmt.Errorf("entry not found: %s", entry.ID)
	}

	entry.UpdatedAt = time.Now()
	w.entries[entry.ID] = entry
	w.touch(entry.ID)

	return nil
}

// Delete removes an entry by ID.
func (w *WorkingMem) Delete(ctx context.Context, id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.deleteEntry(id)
	return nil
}

// Clear removes all entries.
func (w *WorkingMem) Clear(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entries = make(map[string]*Entry)
	w.order = make([]string, 0, w.capacity)
	return nil
}

// Stats returns memory statistics.
func (w *WorkingMem) Stats(ctx context.Context) (*Stats, error) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	byType := make(map[string]int64)
	var totalSize int64

	for _, entry := range w.entries {
		if !w.isExpired(entry) {
			byType[entry.Type]++
			totalSize += int64(len(entry.Content))
		}
	}

	return &Stats{
		TotalEntries: int64(len(w.entries)),
		TotalSize:    totalSize,
		ByType:       byType,
		Hits:         atomic.LoadInt64(&w.hits),
		Misses:       atomic.LoadInt64(&w.misses),
	}, nil
}

// Close releases resources.
func (w *WorkingMem) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.entries = nil
	w.order = nil
	return nil
}

// Capacity returns the maximum number of entries.
func (w *WorkingMem) Capacity() int {
	return w.capacity
}

// Evict removes the least recently used entry.
func (w *WorkingMem) Evict(ctx context.Context) (*Entry, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.evictOldest(), nil
}

// Touch updates the access time for an entry.
func (w *WorkingMem) Touch(ctx context.Context, id string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	entry, exists := w.entries[id]
	if !exists {
		return fmt.Errorf("entry not found: %s", id)
	}

	entry.AccessedAt = time.Now()
	entry.AccessCount++
	w.touch(id)

	return nil
}

// Internal helpers

// touch moves an entry to the end of the LRU order.
func (w *WorkingMem) touch(id string) {
	// Find and remove from current position
	for i, eid := range w.order {
		if eid == id {
			w.order = append(w.order[:i], w.order[i+1:]...)
			break
		}
	}
	// Add to end
	w.order = append(w.order, id)
}

// evictOldest removes the oldest entry.
func (w *WorkingMem) evictOldest() *Entry {
	if len(w.order) == 0 {
		return nil
	}

	oldestID := w.order[0]
	entry := w.entries[oldestID]

	delete(w.entries, oldestID)
	w.order = w.order[1:]

	return entry
}

// deleteEntry removes an entry by ID.
func (w *WorkingMem) deleteEntry(id string) {
	delete(w.entries, id)

	// Remove from order
	for i, eid := range w.order {
		if eid == id {
			w.order = append(w.order[:i], w.order[i+1:]...)
			break
		}
	}
}

// isExpired checks if an entry has expired.
func (w *WorkingMem) isExpired(entry *Entry) bool {
	if entry.TTL == 0 {
		return false
	}
	return time.Since(entry.CreatedAt) > entry.TTL
}

// matchScore calculates how well an entry matches a query.
// Delegates to the shared matchScore function.
func (w *WorkingMem) matchScore(entry *Entry, query *Query) float64 {
	return matchScore(entry, query)
}

// CleanExpired removes all expired entries.
func (w *WorkingMem) CleanExpired(ctx context.Context) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	count := 0
	for id, entry := range w.entries {
		if w.isExpired(entry) {
			w.deleteEntry(id)
			count++
		}
	}
	return count
}

// GetAll returns all non-expired entries.
func (w *WorkingMem) GetAll(ctx context.Context) []*Entry {
	w.mu.RLock()
	defer w.mu.RUnlock()

	entries := make([]*Entry, 0, len(w.entries))
	for _, entry := range w.entries {
		if !w.isExpired(entry) {
			entries = append(entries, entry)
		}
	}
	return entries
}
