package memory

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/JNZader/goreview/goreview/internal/config"
)

// Store provides a unified interface to the cognitive memory system.
// It orchestrates working, session, and long-term memory tiers.
type Store struct {
	mu sync.RWMutex

	working  *WorkingMem
	session  *SessionMem
	longTerm *LongTermMem
	hebbian  *HebbianLearnerImpl
	embedder *Embedder
	index    *SemanticIndex

	cfg config.MemoryConfig
}

// NewStore creates a new memory store from configuration.
func NewStore(cfg config.MemoryConfig) (*Store, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	store := &Store{
		cfg:      cfg,
		embedder: NewEmbedder(),
		index:    NewSemanticIndex(),
	}

	// Initialize working memory
	store.working = NewWorkingMemory(cfg.Working.Capacity, cfg.Working.TTL)

	// Initialize session memory
	var err error
	store.session, err = NewSessionMemory(
		filepath.Join(cfg.Dir, "sessions"),
		cfg.Session.MaxSessions,
		cfg.Session.SessionTTL,
	)
	if err != nil {
		return nil, fmt.Errorf("creating session memory: %w", err)
	}

	// Initialize long-term memory if enabled
	if cfg.LongTerm.Enabled {
		store.longTerm, err = NewLongTermMemory(LongTermOptions{
			Dir:        filepath.Join(cfg.Dir, "longterm"),
			MaxSizeMB:  cfg.LongTerm.MaxSizeMB,
			GCInterval: cfg.LongTerm.GCInterval,
		})
		if err != nil {
			return nil, fmt.Errorf("creating long-term memory: %w", err)
		}
	}

	// Initialize Hebbian learning if enabled
	if cfg.Hebbian.Enabled {
		store.hebbian, err = NewHebbianLearner(HebbianOptions{
			Dir:          filepath.Join(cfg.Dir, "hebbian"),
			LearningRate: cfg.Hebbian.LearningRate,
			DecayRate:    cfg.Hebbian.DecayRate,
			MinStrength:  cfg.Hebbian.MinStrength,
		})
		if err != nil {
			return nil, fmt.Errorf("creating hebbian learner: %w", err)
		}
	}

	return store, nil
}

// Store saves an entry to the appropriate memory tier.
func (s *Store) Store(ctx context.Context, entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate embedding if not provided
	if len(entry.Embedding) == 0 && entry.Content != "" {
		entry.Embedding = s.embedder.Embed(entry.Content)
	}

	// Store in working memory first
	if err := s.working.Store(ctx, entry); err != nil {
		return fmt.Errorf("storing in working memory: %w", err)
	}

	// Also store in session memory for persistence
	if s.session != nil {
		if err := s.session.Store(ctx, entry); err != nil {
			return fmt.Errorf("storing in session memory: %w", err)
		}
	}

	// Index for semantic search
	s.index.Index(entry.ID, entry.Content)

	return nil
}

// Get retrieves an entry by ID, checking all memory tiers.
func (s *Store) Get(ctx context.Context, id string) (*Entry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try working memory first (fastest)
	if entry, err := s.working.Get(ctx, id); err != nil {
		return nil, err
	} else if entry != nil {
		return entry, nil
	}

	// Try session memory
	if s.session != nil {
		if entry, err := s.session.Get(ctx, id); err != nil {
			return nil, err
		} else if entry != nil {
			// Promote to working memory
			_ = s.working.Store(ctx, entry)
			return entry, nil
		}
	}

	// Try long-term memory
	if s.longTerm != nil {
		if entry, err := s.longTerm.Get(ctx, id); err != nil {
			return nil, err
		} else if entry != nil {
			// Promote to working memory
			_ = s.working.Store(ctx, entry)
			return entry, nil
		}
	}

	return nil, nil
}

// Search finds entries matching the query across all tiers.
func (s *Store) Search(ctx context.Context, query *Query) ([]*SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Aggregate results from all tiers
	seen := make(map[string]bool)
	results := make([]*SearchResult, 0)

	// Search working memory
	workingResults, err := s.working.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	for _, r := range workingResults {
		if !seen[r.Entry.ID] {
			seen[r.Entry.ID] = true
			results = append(results, r)
		}
	}

	// Search session memory
	if s.session != nil {
		sessionResults, err := s.session.Search(ctx, query)
		if err != nil {
			return nil, err
		}
		for _, r := range sessionResults {
			if !seen[r.Entry.ID] {
				seen[r.Entry.ID] = true
				results = append(results, r)
			}
		}
	}

	// Search long-term memory
	if s.longTerm != nil {
		longTermResults, err := s.longTerm.Search(ctx, query)
		if err != nil {
			return nil, err
		}
		for _, r := range longTermResults {
			if !seen[r.Entry.ID] {
				seen[r.Entry.ID] = true
				results = append(results, r)
			}
		}
	}

	// Apply limit
	if query != nil && query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results, nil
}

// SemanticSearch finds semantically similar entries.
func (s *Store) SemanticSearch(ctx context.Context, query string, limit int) ([]*SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Generate query embedding
	queryEmbedding := s.embedder.Embed(query)

	// Search in-memory index first
	indexResults := s.index.SearchByEmbedding(queryEmbedding, limit)

	results := make([]*SearchResult, 0, len(indexResults))
	for _, ir := range indexResults {
		// Get full entry
		entry, err := s.Get(ctx, ir.ID)
		if err != nil || entry == nil {
			continue
		}
		results = append(results, &SearchResult{
			Entry: entry,
			Score: ir.Similarity,
		})
	}

	// Also search long-term memory if available
	if s.longTerm != nil {
		ltResults, err := s.longTerm.SemanticSearch(ctx, queryEmbedding, limit)
		if err == nil {
			// Merge results
			seen := make(map[string]bool)
			for _, r := range results {
				seen[r.Entry.ID] = true
			}
			for _, r := range ltResults {
				if !seen[r.Entry.ID] {
					results = append(results, r)
				}
			}
		}
	}

	return results, nil
}

// Associate strengthens the association between two entries.
func (s *Store) Associate(ctx context.Context, sourceID, targetID string) error {
	if s.hebbian == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.hebbian.Strengthen(ctx, sourceID, targetID)
}

// GetAssociations returns entries associated with the given entry.
func (s *Store) GetAssociations(ctx context.Context, id string) ([]*Entry, error) {
	if s.hebbian == nil {
		return nil, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	associations, err := s.hebbian.GetAssociations(ctx, id)
	if err != nil {
		return nil, err
	}

	entries := make([]*Entry, 0, len(associations))
	for _, assoc := range associations {
		// Get the associated entry
		targetID := assoc.TargetID
		if targetID == id {
			targetID = assoc.SourceID // Reverse association
		}

		entry, err := s.Get(ctx, targetID)
		if err != nil || entry == nil {
			continue
		}

		// Set strength from association
		entry.Strength = assoc.Strength
		entries = append(entries, entry)
	}

	return entries, nil
}

// Consolidate moves important entries from working to long-term memory.
func (s *Store) Consolidate(ctx context.Context) error {
	if s.longTerm == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get all entries from working memory
	entries := s.working.GetAll(ctx)

	// Filter for important entries (high access count or strength)
	important := make([]*Entry, 0)
	for _, entry := range entries {
		if entry.AccessCount >= 3 || entry.Strength >= 0.5 {
			important = append(important, entry)
		}
	}

	return s.longTerm.Consolidate(ctx, important)
}

// NewSession starts a new session.
func (s *Store) NewSession(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear working memory
	if err := s.working.Clear(ctx); err != nil {
		return "", err
	}

	// Start new session
	if s.session != nil {
		return s.session.NewSession(ctx)
	}

	return "", nil
}

// SessionID returns the current session ID.
func (s *Store) SessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.session != nil {
		return s.session.SessionID()
	}
	return ""
}

// Stats returns combined statistics from all memory tiers.
func (s *Store) Stats(ctx context.Context) (*StoreStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := &StoreStats{}

	// Working memory stats
	if s.working != nil {
		workingStats, err := s.working.Stats(ctx)
		if err == nil {
			stats.WorkingEntries = workingStats.TotalEntries
			stats.WorkingHits = workingStats.Hits
			stats.WorkingMisses = workingStats.Misses
		}
	}

	// Session memory stats
	if s.session != nil {
		sessionStats, err := s.session.Stats(ctx)
		if err == nil {
			stats.SessionEntries = sessionStats.TotalEntries
		}
	}

	// Long-term memory stats
	if s.longTerm != nil {
		ltStats, err := s.longTerm.Stats(ctx)
		if err == nil {
			stats.LongTermEntries = ltStats.TotalEntries
			stats.LongTermSize = ltStats.TotalSize
		}
	}

	// Hebbian stats
	if s.hebbian != nil {
		total, avgStrength, err := s.hebbian.Stats(ctx)
		if err == nil {
			stats.Associations = total
			stats.AvgAssociationStrength = avgStrength
		}
	}

	// Semantic index stats
	stats.IndexedEntries = int64(s.index.Size())

	return stats, nil
}

// Close releases all resources.
func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var lastErr error

	if s.working != nil {
		if err := s.working.Close(); err != nil {
			lastErr = err
		}
	}

	if s.session != nil {
		if err := s.session.Close(); err != nil {
			lastErr = err
		}
	}

	if s.longTerm != nil {
		if err := s.longTerm.Close(); err != nil {
			lastErr = err
		}
	}

	if s.hebbian != nil {
		if err := s.hebbian.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// RunMaintenance performs maintenance tasks.
func (s *Store) RunMaintenance(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clean expired entries from working memory
	s.working.CleanExpired(ctx)

	// Run decay on Hebbian associations
	if s.hebbian != nil {
		if err := s.hebbian.Decay(ctx); err != nil {
			return fmt.Errorf("hebbian decay: %w", err)
		}
	}

	// Garbage collect long-term memory
	if s.longTerm != nil {
		if _, err := s.longTerm.GarbageCollect(ctx); err != nil {
			return fmt.Errorf("long-term gc: %w", err)
		}
	}

	return nil
}

// StoreStats contains combined memory statistics.
type StoreStats struct {
	// Working memory
	WorkingEntries int64 `json:"working_entries"`
	WorkingHits    int64 `json:"working_hits"`
	WorkingMisses  int64 `json:"working_misses"`

	// Session memory
	SessionEntries int64 `json:"session_entries"`

	// Long-term memory
	LongTermEntries int64 `json:"longterm_entries"`
	LongTermSize    int64 `json:"longterm_size"`

	// Associations
	Associations           int64   `json:"associations"`
	AvgAssociationStrength float64 `json:"avg_association_strength"`

	// Semantic index
	IndexedEntries int64 `json:"indexed_entries"`
}

// DefaultStoreConfig returns a default memory configuration.
func DefaultStoreConfig() config.MemoryConfig {
	return config.MemoryConfig{
		Enabled: true,
		Dir:     ".goreview/memory",
		Working: config.WorkingMemoryConfig{
			Capacity: 100,
			TTL:      15 * time.Minute,
		},
		Session: config.SessionMemoryConfig{
			MaxSessions: 10,
			SessionTTL:  1 * time.Hour,
		},
		LongTerm: config.LongTermMemoryConfig{
			Enabled:    false,
			MaxSizeMB:  500,
			GCInterval: 5 * time.Minute,
		},
		Hebbian: config.HebbianConfig{
			Enabled:      false,
			LearningRate: 0.1,
			DecayRate:    0.01,
			MinStrength:  0.1,
		},
	}
}
