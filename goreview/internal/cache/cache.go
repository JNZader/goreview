// Package cache provides caching for review results.
package cache

import "github.com/JNZader/goreview/goreview/internal/providers"

// Cache defines the interface for caching review results.
type Cache interface {
	// Get retrieves a cached review response.
	Get(key string) (*providers.ReviewResponse, bool, error)

	// Set stores a review response in the cache.
	Set(key string, response *providers.ReviewResponse) error

	// ComputeKey generates a cache key from a review request.
	ComputeKey(req *providers.ReviewRequest) string

	// Clear removes all cached entries.
	Clear() error

	// Close releases cache resources.
	Close() error
}
