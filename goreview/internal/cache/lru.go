package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JNZader/goreview/goreview/internal/providers"
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
		_ = c.Delete(key)
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

func (c *LRUCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
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
