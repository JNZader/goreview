package memory

import (
	"context"
)

// NoopMemory is a no-operation memory implementation.
// It's used when the memory system is disabled.
type NoopMemory struct{}

// NewNoopMemory creates a new NoopMemory instance.
func NewNoopMemory() *NoopMemory {
	return &NoopMemory{}
}

// Compile-time interface checks.
var (
	_ Memory         = (*NoopMemory)(nil)
	_ WorkingMemory  = (*NoopMemory)(nil)
	_ SessionMemory  = (*NoopMemory)(nil)
	_ LongTermMemory = (*NoopMemory)(nil)
	_ HebbianLearner = (*NoopMemory)(nil)
)

// Store is a no-op.
func (n *NoopMemory) Store(_ context.Context, _ *Entry) error {
	return nil
}

// Get returns nil (entry not found).
func (n *NoopMemory) Get(_ context.Context, _ string) (*Entry, error) {
	return nil, nil
}

// Search returns empty results.
func (n *NoopMemory) Search(_ context.Context, _ *Query) ([]*SearchResult, error) {
	return nil, nil
}

// Update is a no-op.
func (n *NoopMemory) Update(_ context.Context, _ *Entry) error {
	return nil
}

// Delete is a no-op.
func (n *NoopMemory) Delete(_ context.Context, _ string) error {
	return nil
}

// Clear is a no-op.
func (n *NoopMemory) Clear(_ context.Context) error {
	return nil
}

// Stats returns empty stats.
func (n *NoopMemory) Stats(_ context.Context) (*Stats, error) {
	return &Stats{
		ByType: make(map[string]int64),
	}, nil
}

// Close is a no-op.
func (n *NoopMemory) Close() error {
	return nil
}

// WorkingMemory interface

// Capacity returns 0.
func (n *NoopMemory) Capacity() int {
	return 0
}

// Evict returns nil.
func (n *NoopMemory) Evict(_ context.Context) (*Entry, error) {
	return nil, nil
}

// Touch is a no-op.
func (n *NoopMemory) Touch(_ context.Context, _ string) error {
	return nil
}

// SessionMemory interface

// SessionID returns empty string.
func (n *NoopMemory) SessionID() string {
	return ""
}

// NewSession returns empty session ID.
func (n *NoopMemory) NewSession(_ context.Context) (string, error) {
	return "", nil
}

// LoadSession is a no-op.
func (n *NoopMemory) LoadSession(_ context.Context, _ string) error {
	return nil
}

// ListSessions returns empty list.
func (n *NoopMemory) ListSessions(_ context.Context) ([]string, error) {
	return nil, nil
}

// LongTermMemory interface

// SemanticSearch returns empty results.
func (n *NoopMemory) SemanticSearch(_ context.Context, _ []float32, _ int) ([]*SearchResult, error) {
	return nil, nil
}

// Consolidate is a no-op.
func (n *NoopMemory) Consolidate(_ context.Context, _ []*Entry) error {
	return nil
}

// GarbageCollect returns 0 (no entries collected).
func (n *NoopMemory) GarbageCollect(_ context.Context) (int, error) {
	return 0, nil
}

// HebbianLearner interface

// Strengthen is a no-op.
func (n *NoopMemory) Strengthen(_ context.Context, _, _ string) error {
	return nil
}

// Weaken is a no-op.
func (n *NoopMemory) Weaken(_ context.Context, _, _ string) error {
	return nil
}

// GetAssociations returns empty list.
func (n *NoopMemory) GetAssociations(_ context.Context, _ string) ([]*Association, error) {
	return nil, nil
}

// Decay is a no-op.
func (n *NoopMemory) Decay(_ context.Context) error {
	return nil
}

// Prune returns 0 (no associations pruned).
func (n *NoopMemory) Prune(_ context.Context, _ float64) (int, error) {
	return 0, nil
}
