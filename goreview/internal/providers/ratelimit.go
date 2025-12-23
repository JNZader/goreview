package providers

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter.
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the given RPS.
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(rps),
		maxTokens:  float64(rps),
		refillRate: float64(rps),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		r.refill()

		if r.tokens >= 1 {
			r.tokens--
			r.mu.Unlock()
			return nil
		}

		waitTime := time.Duration(float64(time.Second) / r.refillRate)
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue loop
		}
	}
}

func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.tokens += elapsed * r.refillRate
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}
	r.lastRefill = now
}
