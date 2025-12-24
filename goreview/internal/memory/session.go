package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// SessionMem implements SessionMemory with file-based persistence.
type SessionMem struct {
	mu          sync.RWMutex
	sessionID   string
	entries     map[string]*Entry
	dir         string
	maxSessions int
	sessionTTL  time.Duration

	// Statistics
	hits   int64
	misses int64
}

// NewSessionMemory creates a new session memory instance.
func NewSessionMemory(dir string, maxSessions int, sessionTTL time.Duration) (*SessionMem, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("creating session directory: %w", err)
	}

	sm := &SessionMem{
		sessionID:   uuid.New().String(),
		entries:     make(map[string]*Entry),
		dir:         dir,
		maxSessions: maxSessions,
		sessionTTL:  sessionTTL,
	}

	// Clean old sessions
	if err := sm.cleanOldSessions(); err != nil {
		// Non-fatal, just log and continue
		_ = err
	}

	return sm, nil
}

// Compile-time interface check.
var _ SessionMemory = (*SessionMem)(nil)

// Store saves an entry to session memory.
func (s *SessionMem) Store(ctx context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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

	s.entries[entry.ID] = entry
	return nil
}

// Get retrieves an entry by ID.
func (s *SessionMem) Get(ctx context.Context, id string) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[id]
	if !exists {
		atomic.AddInt64(&s.misses, 1)
		return nil, nil
	}

	entry.AccessedAt = time.Now()
	entry.AccessCount++

	atomic.AddInt64(&s.hits, 1)
	return entry, nil
}

// Search finds entries matching the query.
func (s *SessionMem) Search(ctx context.Context, query *Query) ([]*SearchResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*SearchResult, 0, len(s.entries))

	for _, entry := range s.entries {
		score := matchScore(entry, query)
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
func (s *SessionMem) Update(ctx context.Context, entry *Entry) error {
	if entry == nil {
		return fmt.Errorf("entry cannot be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[entry.ID]; !exists {
		return fmt.Errorf("entry not found: %s", entry.ID)
	}

	entry.UpdatedAt = time.Now()
	s.entries[entry.ID] = entry

	return nil
}

// Delete removes an entry by ID.
func (s *SessionMem) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, id)
	return nil
}

// Clear removes all entries.
func (s *SessionMem) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries = make(map[string]*Entry)
	return nil
}

// Stats returns memory statistics.
func (s *SessionMem) Stats(ctx context.Context) (*Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	byType := make(map[string]int64)
	var totalSize int64

	for _, entry := range s.entries {
		byType[entry.Type]++
		totalSize += int64(len(entry.Content))
	}

	return &Stats{
		TotalEntries: int64(len(s.entries)),
		TotalSize:    totalSize,
		ByType:       byType,
		Hits:         atomic.LoadInt64(&s.hits),
		Misses:       atomic.LoadInt64(&s.misses),
	}, nil
}

// Close persists the session and releases resources.
func (s *SessionMem) Close() error {
	return s.saveSession()
}

// SessionID returns the current session identifier.
func (s *SessionMem) SessionID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessionID
}

// NewSession starts a new session.
func (s *SessionMem) NewSession(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Save current session
	if err := s.saveSessionLocked(); err != nil {
		return "", fmt.Errorf("saving current session: %w", err)
	}

	// Start new session
	s.sessionID = uuid.New().String()
	s.entries = make(map[string]*Entry)
	atomic.StoreInt64(&s.hits, 0)
	atomic.StoreInt64(&s.misses, 0)

	return s.sessionID, nil
}

// LoadSession restores a previous session.
func (s *SessionMem) LoadSession(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	sessionFile := filepath.Join(s.dir, sessionID+".json")
	data, err := os.ReadFile(sessionFile) //nolint:gosec // Path is constructed from trusted session directory and UUID
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session not found: %s", sessionID)
		}
		return fmt.Errorf("reading session: %w", err)
	}

	var session sessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("parsing session: %w", err)
	}

	// Restore session
	s.sessionID = sessionID
	s.entries = make(map[string]*Entry)
	for _, entry := range session.Entries {
		entryCopy := entry // Copy to avoid pointer issues
		s.entries[entry.ID] = &entryCopy
	}

	return nil
}

// ListSessions returns available session IDs.
func (s *SessionMem) ListSessions(ctx context.Context) ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading sessions: %w", err)
	}

	sessions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			sessionID := entry.Name()[:len(entry.Name())-5] // Remove .json
			sessions = append(sessions, sessionID)
		}
	}

	return sessions, nil
}

// Internal types

type sessionData struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Entries   []Entry   `json:"entries"`
}

// Internal helpers

func (s *SessionMem) saveSession() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveSessionLocked()
}

func (s *SessionMem) saveSessionLocked() error {
	if len(s.entries) == 0 {
		return nil
	}

	entries := make([]Entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, *e)
	}

	session := sessionData{
		ID:        s.sessionID,
		CreatedAt: time.Now(),
		Entries:   entries,
	}

	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}

	sessionFile := filepath.Join(s.dir, s.sessionID+".json")
	if err := os.WriteFile(sessionFile, data, 0600); err != nil {
		return fmt.Errorf("writing session: %w", err)
	}

	return nil
}

func (s *SessionMem) cleanOldSessions() error {
	sessions, err := s.ListSessions(context.Background())
	if err != nil {
		return err
	}

	if len(sessions) <= s.maxSessions {
		return nil
	}

	// Get session info with modification times
	type sessionInfo struct {
		id      string
		modTime time.Time
	}

	infos := make([]sessionInfo, 0, len(sessions))
	for _, sessionID := range sessions {
		file := filepath.Join(s.dir, sessionID+".json")
		stat, err := os.Stat(file)
		if err != nil {
			continue
		}
		infos = append(infos, sessionInfo{id: sessionID, modTime: stat.ModTime()})
	}

	// Sort by modification time (oldest first)
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].modTime.Before(infos[j].modTime)
	})

	// Remove old sessions
	toRemove := len(infos) - s.maxSessions
	for i := 0; i < toRemove; i++ {
		file := filepath.Join(s.dir, infos[i].id+".json")
		_ = os.Remove(file)
	}

	return nil
}

// matchResult holds intermediate matching state.
type matchResult struct {
	score   float64
	matches int
	done    bool // early return indicator
}

// matchByID checks if entry matches by ID.
func matchByID(entry *Entry, query *Query) (float64, bool) {
	if query.ID == "" {
		return 0, false
	}
	if entry.ID == query.ID {
		return 1.0, true
	}
	return 0, true
}

// matchByType checks if entry matches by type.
func matchByType(entry *Entry, query *Query, r *matchResult) bool {
	if query.Type == "" {
		return true
	}
	if entry.Type == query.Type {
		r.score += 1.0
		r.matches++
		return true
	}
	return false
}

// matchByTags checks if entry matches by tags.
func matchByTags(entry *Entry, query *Query, r *matchResult) bool {
	if len(query.Tags) == 0 {
		return true
	}
	entryTags := make(map[string]bool)
	for _, t := range entry.Tags {
		entryTags[t] = true
	}
	tagMatches := 0
	for _, t := range query.Tags {
		if entryTags[t] {
			tagMatches++
		}
	}
	if tagMatches == 0 {
		return false
	}
	r.score += float64(tagMatches) / float64(len(query.Tags))
	r.matches++
	return true
}

// matchByContent checks if entry matches by content.
func matchByContent(entry *Entry, query *Query, r *matchResult) bool {
	if query.Content == "" {
		return true
	}
	if contains(toLowerASCII(entry.Content), toLowerASCII(query.Content)) {
		r.score += 1.0
		r.matches++
		return true
	}
	return false
}

// matchByStrength checks if entry matches by strength threshold.
func matchByStrength(entry *Entry, query *Query, r *matchResult) bool {
	if query.MinStrength <= 0 {
		return true
	}
	if entry.Strength < query.MinStrength {
		return false
	}
	r.score += entry.Strength
	r.matches++
	return true
}

// toLowerASCII converts ASCII letters to lowercase.
func toLowerASCII(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// matchScore calculates how well an entry matches a query.
// Shared between SessionMem and WorkingMem.
func matchScore(entry *Entry, query *Query) float64 {
	if query == nil {
		return 1.0
	}

	// Check ID match first (exact match)
	if score, done := matchByID(entry, query); done {
		return score
	}

	r := &matchResult{}

	// Check each criterion
	if !matchByType(entry, query, r) {
		return 0
	}
	if !matchByTags(entry, query, r) {
		return 0
	}
	if !matchByContent(entry, query, r) {
		return 0
	}
	if !matchByStrength(entry, query, r) {
		return 0
	}

	if r.matches == 0 {
		return 1.0 // No filters, everything matches
	}
	return r.score / float64(r.matches)
}

// contains checks if s contains substr.
func contains(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
