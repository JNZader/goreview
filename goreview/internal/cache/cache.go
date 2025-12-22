// Package cache provides caching for review results.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/JNZader/goreview/goreview/internal/providers"
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
	Hits      int64 `json:"hits"`
	Misses    int64 `json:"misses"`
	Entries   int   `json:"entries"`
	SizeBytes int64 `json:"size_bytes"`
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
