package providers

import "context"

// RateLimiter implements a simple token bucket rate limiter.
// Placeholder - will be fully implemented in commit 4.4.
type RateLimiter struct {
	rps int
}

// NewRateLimiter creates a new rate limiter with the given RPS.
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{rps: rps}
}

// Wait blocks until a token is available or context is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	// Placeholder - no-op for now
	return nil
}
