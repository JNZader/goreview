package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync/atomic"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/google/uuid"
)

// LongTermMem implements LongTermMemory using BadgerDB for persistence.
type LongTermMem struct {
	db         *badger.DB
	gcInterval time.Duration
	gcStop     chan struct{}

	// Statistics
	hits   int64
	misses int64
}

// LongTermOptions configures long-term memory.
type LongTermOptions struct {
	Dir        string
	MaxSizeMB  int
	GCInterval time.Duration
}

// NewLongTermMemory creates a new long-term memory instance.
func NewLongTermMemory(opts LongTermOptions) (*LongTermMem, error) {
	badgerOpts := badger.DefaultOptions(opts.Dir)
	badgerOpts.Logger = nil // Disable BadgerDB logging

	// Set value log size limits
	if opts.MaxSizeMB > 0 {
		badgerOpts.ValueLogFileSize = int64(opts.MaxSizeMB) * 1024 * 1024 / 10
	}

	db, err := badger.Open(badgerOpts)
	if err != nil {
		return nil, fmt.Errorf("opening badger db: %w", err)
	}

	ltm := &LongTermMem{
		db:         db,
		gcInterval: opts.GCInterval,
		gcStop:     make(chan struct{}),
	}

	// Start background GC
	if opts.GCInterval > 0 {
		go ltm.runGC()
	}

	return ltm, nil
}

// Compile-time interface check.
var _ LongTermMemory = (*LongTermMem)(nil)

// Store saves an entry to long-term memory.
func (l *LongTermMem) Store(ctx context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	// Generate ID if not provided
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}

	now := time.Now()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = now
	}
	entry.UpdatedAt = now
	entry.AccessedAt = now

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling entry: %w", err)
	}

	return l.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(entry.ID), data)
	})
}

// Get retrieves an entry by ID.
func (l *LongTermMem) Get(ctx context.Context, id string) (*Entry, error) {
	var entry Entry

	err := l.db.View(func(txn *badger.Txn) error {
		item, getErr := txn.Get([]byte(id))
		if getErr != nil {
			return getErr
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &entry)
		})
	})

	if err == badger.ErrKeyNotFound {
		atomic.AddInt64(&l.misses, 1)
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("getting entry: %w", err)
	}

	// Update access stats
	entry.AccessedAt = time.Now()
	entry.AccessCount++

	// Update in background (non-blocking)
	go func() {
		_ = l.Update(context.Background(), &entry)
	}()

	atomic.AddInt64(&l.hits, 1)
	return &entry, nil
}

// Search finds entries matching the query.
func (l *LongTermMem) Search(ctx context.Context, query *Query) ([]*SearchResult, error) {
	results := make([]*SearchResult, 0)

	err := l.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100

		it := txn.NewIterator(opts)
		defer it.Close() //nolint:errcheck

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			var entry Entry
			valErr := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &entry)
			})
			if valErr != nil {
				continue
			}

			// Check TTL
			if entry.TTL > 0 && time.Since(entry.CreatedAt) > entry.TTL {
				continue
			}

			score := matchScore(&entry, query)
			if score > 0 {
				entryCopy := entry
				results = append(results, &SearchResult{
					Entry: &entryCopy,
					Score: score,
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("searching: %w", err)
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit and offset
	if query != nil {
		if query.Offset > 0 && query.Offset < len(results) {
			results = results[query.Offset:]
		}
		if query.Limit > 0 && query.Limit < len(results) {
			results = results[:query.Limit]
		}
	}

	return results, nil
}

// Update modifies an existing entry.
func (l *LongTermMem) Update(ctx context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	entry.UpdatedAt = time.Now()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshaling entry: %w", err)
	}

	return l.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(entry.ID), data)
	})
}

// Delete removes an entry by ID.
func (l *LongTermMem) Delete(ctx context.Context, id string) error {
	return l.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(id))
	})
}

// Clear removes all entries.
func (l *LongTermMem) Clear(ctx context.Context) error {
	return l.db.DropAll()
}

// Stats returns memory statistics.
func (l *LongTermMem) Stats(ctx context.Context) (*Stats, error) {
	var totalEntries int64
	byType := make(map[string]int64)

	err := l.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		opts.PrefetchSize = 100

		it := txn.NewIterator(opts)
		defer it.Close() //nolint:errcheck

		for it.Rewind(); it.Valid(); it.Next() {
			totalEntries++

			var entry Entry
			valErr := it.Item().Value(func(val []byte) error {
				return json.Unmarshal(val, &entry)
			})
			if valErr == nil {
				byType[entry.Type]++
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("getting stats: %w", err)
	}

	// Get LSM size
	lsmSize, vlogSize := l.db.Size()

	return &Stats{
		TotalEntries: totalEntries,
		TotalSize:    lsmSize + vlogSize,
		ByType:       byType,
		Hits:         atomic.LoadInt64(&l.hits),
		Misses:       atomic.LoadInt64(&l.misses),
	}, nil
}

// Close releases resources.
func (l *LongTermMem) Close() error {
	// Stop GC
	close(l.gcStop)
	return l.db.Close()
}

// SemanticSearch finds entries similar to the given embedding.
func (l *LongTermMem) SemanticSearch(ctx context.Context, embedding []float32, limit int) ([]*SearchResult, error) {
	if len(embedding) == 0 {
		return nil, nil
	}

	results := make([]*SearchResult, 0)

	err := l.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 100

		it := txn.NewIterator(opts)
		defer it.Close() //nolint:errcheck

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			var entry Entry
			valErr := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &entry)
			})
			if valErr != nil {
				continue
			}

			// Skip entries without embeddings
			if len(entry.Embedding) == 0 {
				continue
			}

			// Calculate cosine similarity
			similarity := cosineSimilarity(embedding, entry.Embedding)
			if similarity > 0 {
				entryCopy := entry
				results = append(results, &SearchResult{
					Entry: &entryCopy,
					Score: similarity,
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("semantic search: %w", err)
	}

	// Sort by similarity (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if limit > 0 && limit < len(results) {
		results = results[:limit]
	}

	return results, nil
}

// Consolidate moves important entries from working memory.
func (l *LongTermMem) Consolidate(ctx context.Context, entries []*Entry) error {
	if len(entries) == 0 {
		return nil
	}

	return l.db.Update(func(txn *badger.Txn) error {
		for _, entry := range entries {
			if entry == nil {
				continue
			}

			// Only consolidate entries with sufficient strength
			if entry.Strength < 0.5 {
				continue
			}

			data, err := json.Marshal(entry)
			if err != nil {
				continue
			}

			if err := txn.Set([]byte(entry.ID), data); err != nil {
				return err
			}
		}
		return nil
	})
}

// GarbageCollect removes expired and weak entries.
func (l *LongTermMem) GarbageCollect(ctx context.Context) (int, error) {
	keysToDelete := make([][]byte, 0)

	// Find entries to delete
	err := l.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close() //nolint:errcheck

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			var entry Entry
			valErr := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &entry)
			})
			if valErr != nil {
				continue
			}

			shouldDelete := false

			// Check TTL expiration
			if entry.TTL > 0 && time.Since(entry.CreatedAt) > entry.TTL {
				shouldDelete = true
			}

			// Check weak strength (not accessed recently)
			if entry.Strength < 0.1 && time.Since(entry.AccessedAt) > 7*24*time.Hour {
				shouldDelete = true
			}

			if shouldDelete {
				keysToDelete = append(keysToDelete, item.KeyCopy(nil))
			}
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("scanning for gc: %w", err)
	}

	// Delete entries
	if len(keysToDelete) > 0 {
		err = l.db.Update(func(txn *badger.Txn) error {
			for _, key := range keysToDelete {
				if delErr := txn.Delete(key); delErr != nil {
					return delErr
				}
			}
			return nil
		})
		if err != nil {
			return 0, fmt.Errorf("deleting entries: %w", err)
		}
	}

	// Run BadgerDB GC
	_ = l.db.RunValueLogGC(0.5)

	return len(keysToDelete), nil
}

// runGC runs periodic garbage collection.
func (l *LongTermMem) runGC() {
	ticker := time.NewTicker(l.gcInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_, _ = l.GarbageCollect(context.Background())
		case <-l.gcStop:
			return
		}
	}
}

// cosineSimilarity calculates the cosine similarity between two vectors.
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
