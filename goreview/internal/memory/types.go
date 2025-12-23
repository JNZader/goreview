// Package memory provides a cognitive memory system for storing and retrieving
// review context, patterns, and learned associations.
package memory

import (
	"context"
	"time"
)

// Entry represents a memory entry with content and metadata.
type Entry struct {
	// ID is the unique identifier for this memory entry.
	ID string `json:"id"`

	// Content is the main content of the memory entry.
	Content string `json:"content"`

	// Type classifies the memory (e.g., "review", "pattern", "rule", "context").
	Type string `json:"type"`

	// Tags are labels for categorizing and filtering entries.
	Tags []string `json:"tags,omitempty"`

	// Metadata contains additional key-value pairs.
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Embedding is the vector representation for semantic search.
	Embedding []float32 `json:"embedding,omitempty"`

	// CreatedAt is when the entry was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the entry was last modified.
	UpdatedAt time.Time `json:"updated_at"`

	// AccessedAt is when the entry was last accessed.
	AccessedAt time.Time `json:"accessed_at"`

	// AccessCount tracks how many times this entry was accessed.
	AccessCount int64 `json:"access_count"`

	// Strength represents the association strength (for Hebbian learning).
	Strength float64 `json:"strength"`

	// TTL is the time-to-live for this entry (0 means no expiration).
	TTL time.Duration `json:"ttl,omitempty"`
}

// Query represents a search query for memory entries.
type Query struct {
	// ID searches for a specific entry by ID.
	ID string

	// Type filters by entry type.
	Type string

	// Tags filters entries that have all specified tags.
	Tags []string

	// Content performs text search on content.
	Content string

	// Embedding performs semantic similarity search.
	Embedding []float32

	// MinStrength filters entries with strength >= this value.
	MinStrength float64

	// Limit restricts the number of results.
	Limit int

	// Offset skips the first N results.
	Offset int

	// SortBy specifies the field to sort by.
	SortBy string

	// SortDesc sorts in descending order when true.
	SortDesc bool
}

// SearchResult represents a search result with relevance score.
type SearchResult struct {
	Entry *Entry  `json:"entry"`
	Score float64 `json:"score"`
}

// Stats contains memory system statistics.
type Stats struct {
	// TotalEntries is the total number of entries.
	TotalEntries int64 `json:"total_entries"`

	// TotalSize is the approximate size in bytes.
	TotalSize int64 `json:"total_size"`

	// ByType contains entry counts by type.
	ByType map[string]int64 `json:"by_type"`

	// Hits is the number of successful retrievals.
	Hits int64 `json:"hits"`

	// Misses is the number of failed retrievals.
	Misses int64 `json:"misses"`
}

// Memory defines the interface for the cognitive memory system.
type Memory interface {
	// Store saves an entry to memory.
	Store(ctx context.Context, entry *Entry) error

	// Get retrieves an entry by ID.
	Get(ctx context.Context, id string) (*Entry, error)

	// Search finds entries matching the query.
	Search(ctx context.Context, query *Query) ([]*SearchResult, error)

	// Update modifies an existing entry.
	Update(ctx context.Context, entry *Entry) error

	// Delete removes an entry by ID.
	Delete(ctx context.Context, id string) error

	// Clear removes all entries.
	Clear(ctx context.Context) error

	// Stats returns memory statistics.
	Stats(ctx context.Context) (*Stats, error)

	// Close releases resources.
	Close() error
}

// WorkingMemory provides fast, short-term storage for active context.
type WorkingMemory interface {
	Memory

	// Capacity returns the maximum number of entries.
	Capacity() int

	// Evict removes the least recently used entry.
	Evict(ctx context.Context) (*Entry, error)

	// Touch updates the access time for an entry.
	Touch(ctx context.Context, id string) error
}

// SessionMemory provides session-scoped storage.
type SessionMemory interface {
	Memory

	// SessionID returns the current session identifier.
	SessionID() string

	// NewSession starts a new session, optionally preserving some context.
	NewSession(ctx context.Context) (string, error)

	// LoadSession restores a previous session.
	LoadSession(ctx context.Context, sessionID string) error

	// ListSessions returns available session IDs.
	ListSessions(ctx context.Context) ([]string, error)
}

// LongTermMemory provides persistent storage with semantic capabilities.
type LongTermMemory interface {
	Memory

	// SemanticSearch finds entries similar to the given embedding.
	SemanticSearch(ctx context.Context, embedding []float32, limit int) ([]*SearchResult, error)

	// Consolidate moves important entries from working memory.
	Consolidate(ctx context.Context, entries []*Entry) error

	// GarbageCollect removes expired and weak entries.
	GarbageCollect(ctx context.Context) (int, error)
}

// Association represents a learned connection between concepts.
type Association struct {
	// SourceID is the ID of the source entry.
	SourceID string `json:"source_id"`

	// TargetID is the ID of the target entry.
	TargetID string `json:"target_id"`

	// Strength is the association strength (0.0 to 1.0).
	Strength float64 `json:"strength"`

	// CoActivations counts how many times both were accessed together.
	CoActivations int64 `json:"co_activations"`

	// CreatedAt is when the association was created.
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the association was last modified.
	UpdatedAt time.Time `json:"updated_at"`
}

// HebbianLearner manages association learning between memory entries.
type HebbianLearner interface {
	// Strengthen increases the association between two entries.
	Strengthen(ctx context.Context, sourceID, targetID string) error

	// Weaken decreases the association between two entries.
	Weaken(ctx context.Context, sourceID, targetID string) error

	// GetAssociations returns associations for an entry.
	GetAssociations(ctx context.Context, id string) ([]*Association, error)

	// Decay applies time-based decay to all associations.
	Decay(ctx context.Context) error

	// Prune removes associations below minimum strength.
	Prune(ctx context.Context, minStrength float64) (int, error)
}
