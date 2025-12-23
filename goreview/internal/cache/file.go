package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/JNZader/goreview/goreview/internal/providers"
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

	return os.WriteFile(c.keyPath(key), data, 0600)
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
