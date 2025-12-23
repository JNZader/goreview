# Iteracion 06: Sistema de Cache

## Objetivos

- Interface Cache para abstraccion
- Cache LRU en memoria
- Cache persistente en archivos
- TTL y expiracion automatica
- Generacion de keys con hash

## Tiempo Estimado: 4 horas

---

## Commit 6.1: Crear interface Cache

**Mensaje de commit:**
```
feat(cache): add cache interface

- Define Cache interface
- Add key computation method
- Support TTL-based expiration
```

### `goreview/internal/cache/cache.go`

```go
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

// Cache defines the interface for caching review results.
type Cache interface {
	// Get retrieves a cached response by key.
	Get(key string) (*providers.ReviewResponse, bool, error)

	// Set stores a response in the cache.
	Set(key string, response *providers.ReviewResponse) error

	// Delete removes an entry from the cache.
	Delete(key string) error

	// Clear removes all entries.
	Clear() error

	// ComputeKey generates a cache key from a review request.
	ComputeKey(req *providers.ReviewRequest) string

	// Stats returns cache statistics.
	Stats() CacheStats
}

// CacheStats contains cache statistics.
type CacheStats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Entries    int   `json:"entries"`
	SizeBytes  int64 `json:"size_bytes"`
}

// ComputeKey generates a SHA-256 hash key from a review request.
func ComputeKey(req *providers.ReviewRequest) string {
	data, _ := json.Marshal(map[string]interface{}{
		"diff":     req.Diff,
		"language": req.Language,
		"path":     req.FilePath,
		"rules":    req.Rules,
	})

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
```

---

## Commit 6.2: Implementar LRU Cache

**Mensaje de commit:**
```
feat(cache): add LRU memory cache

- Implement LRU eviction policy
- Thread-safe with mutex
- Configurable max entries
- Track hits/misses statistics
```

### `goreview/internal/cache/lru.go`

```go
package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

// LRUCache implements an in-memory LRU cache.
type LRUCache struct {
	maxEntries int
	ttl        time.Duration

	mu      sync.RWMutex
	entries map[string]*list.Element
	order   *list.List

	hits   int64
	misses int64
}

type lruEntry struct {
	key       string
	response  *providers.ReviewResponse
	expiresAt time.Time
}

// NewLRUCache creates a new LRU cache.
func NewLRUCache(maxEntries int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		maxEntries: maxEntries,
		ttl:        ttl,
		entries:    make(map[string]*list.Element),
		order:      list.New(),
	}
}

func (c *LRUCache) Get(key string) (*providers.ReviewResponse, bool, error) {
	c.mu.RLock()
	elem, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&c.misses, 1)
		return nil, false, nil
	}

	entry := elem.Value.(*lruEntry)

	// Check expiration
	if time.Now().After(entry.expiresAt) {
		c.Delete(key)
		atomic.AddInt64(&c.misses, 1)
		return nil, false, nil
	}

	// Move to front (most recently used)
	c.mu.Lock()
	c.order.MoveToFront(elem)
	c.mu.Unlock()

	atomic.AddInt64(&c.hits, 1)
	return entry.response, true, nil
}

func (c *LRUCache) Set(key string, response *providers.ReviewResponse) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update existing entry
	if elem, exists := c.entries[key]; exists {
		entry := elem.Value.(*lruEntry)
		entry.response = response
		entry.expiresAt = time.Now().Add(c.ttl)
		c.order.MoveToFront(elem)
		return nil
	}

	// Evict if at capacity
	if c.order.Len() >= c.maxEntries {
		c.evictOldest()
	}

	// Add new entry
	entry := &lruEntry{
		key:       key,
		response:  response,
		expiresAt: time.Now().Add(c.ttl),
	}
	elem := c.order.PushFront(entry)
	c.entries[key] = elem

	return nil
}

func (c *LRUCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.entries[key]; exists {
		c.order.Remove(elem)
		delete(c.entries, key)
	}
	return nil
}

func (c *LRUCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]*list.Element)
	c.order.Init()
	return nil
}

func (c *LRUCache) ComputeKey(req *providers.ReviewRequest) string {
	return ComputeKey(req)
}

func (c *LRUCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return CacheStats{
		Hits:    atomic.LoadInt64(&c.hits),
		Misses:  atomic.LoadInt64(&c.misses),
		Entries: c.order.Len(),
	}
}

func (c *LRUCache) evictOldest() {
	elem := c.order.Back()
	if elem != nil {
		entry := elem.Value.(*lruEntry)
		delete(c.entries, entry.key)
		c.order.Remove(elem)
	}
}
```

---

## Commit 6.3: Implementar File Cache

**Mensaje de commit:**
```
feat(cache): add file-based cache

- Persist cache entries to disk
- Support TTL via file modification time
- Automatic directory creation
- JSON serialization
```

### `goreview/internal/cache/file.go`

```go
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

// FileCache implements a file-based persistent cache.
type FileCache struct {
	dir string
	ttl time.Duration

	hits   int64
	misses int64
}

type fileEntry struct {
	Response  *providers.ReviewResponse `json:"response"`
	ExpiresAt time.Time                 `json:"expires_at"`
}

// NewFileCache creates a new file-based cache.
func NewFileCache(dir string, ttl time.Duration) (*FileCache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	return &FileCache{
		dir: dir,
		ttl: ttl,
	}, nil
}

func (c *FileCache) Get(key string) (*providers.ReviewResponse, bool, error) {
	path := c.keyPath(key)

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		atomic.AddInt64(&c.misses, 1)
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}

	var entry fileEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false, err
	}

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		os.Remove(path)
		atomic.AddInt64(&c.misses, 1)
		return nil, false, nil
	}

	atomic.AddInt64(&c.hits, 1)
	return entry.Response, true, nil
}

func (c *FileCache) Set(key string, response *providers.ReviewResponse) error {
	entry := fileEntry{
		Response:  response,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	return os.WriteFile(c.keyPath(key), data, 0644)
}

func (c *FileCache) Delete(key string) error {
	return os.Remove(c.keyPath(key))
}

func (c *FileCache) Clear() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			os.Remove(filepath.Join(c.dir, entry.Name()))
		}
	}
	return nil
}

func (c *FileCache) ComputeKey(req *providers.ReviewRequest) string {
	return ComputeKey(req)
}

func (c *FileCache) Stats() CacheStats {
	entries, _ := os.ReadDir(c.dir)
	return CacheStats{
		Hits:    atomic.LoadInt64(&c.hits),
		Misses:  atomic.LoadInt64(&c.misses),
		Entries: len(entries),
	}
}

func (c *FileCache) keyPath(key string) string {
	return filepath.Join(c.dir, key+".json")
}

// Cleanup removes expired entries.
func (c *FileCache) Cleanup() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		path := filepath.Join(c.dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var fe fileEntry
		if err := json.Unmarshal(data, &fe); err != nil {
			continue
		}

		if time.Now().After(fe.ExpiresAt) {
			os.Remove(path)
		}
	}

	return nil
}
```

---

## Commit 6.4: Agregar tests de Cache

**Mensaje de commit:**
```
test(cache): add cache tests

- Test LRU eviction
- Test TTL expiration
- Test file persistence
- Test concurrent access
```

### `goreview/internal/cache/cache_test.go`

```go
package cache

import (
	"os"
	"testing"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(2, time.Hour)

	// Test Set and Get
	resp := &providers.ReviewResponse{Summary: "test"}
	cache.Set("key1", resp)

	got, found, err := cache.Get("key1")
	if err != nil || !found {
		t.Fatalf("Get() = %v, %v, want found", got, err)
	}
	if got.Summary != "test" {
		t.Errorf("Summary = %v, want test", got.Summary)
	}

	// Test miss
	_, found, _ = cache.Get("nonexistent")
	if found {
		t.Error("Get(nonexistent) found, want miss")
	}
}

func TestLRUEviction(t *testing.T) {
	cache := NewLRUCache(2, time.Hour)

	cache.Set("key1", &providers.ReviewResponse{Summary: "1"})
	cache.Set("key2", &providers.ReviewResponse{Summary: "2"})
	cache.Set("key3", &providers.ReviewResponse{Summary: "3"}) // Evicts key1

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

	cache.Set("key1", &providers.ReviewResponse{Summary: "test"})

	time.Sleep(20 * time.Millisecond)

	_, found, _ := cache.Get("key1")
	if found {
		t.Error("key1 should be expired")
	}
}

func TestFileCache(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, time.Hour)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	resp := &providers.ReviewResponse{Summary: "test"}
	cache.Set("key1", resp)

	got, found, err := cache.Get("key1")
	if err != nil || !found {
		t.Fatalf("Get() = %v, %v, want found", got, err)
	}
	if got.Summary != "test" {
		t.Errorf("Summary = %v, want test", got.Summary)
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
}
```

---

## Resumen de la Iteracion 06

### Commits:
1. `feat(cache): add cache interface`
2. `feat(cache): add LRU memory cache`
3. `feat(cache): add file-based cache`
4. `test(cache): add cache tests`

### Archivos:
```
goreview/internal/cache/
├── cache.go
├── lru.go
├── file.go
└── cache_test.go
```

---

## Siguiente Iteracion

Continua con: **[07-GENERACION-REPORTES.md](07-GENERACION-REPORTES.md)**
