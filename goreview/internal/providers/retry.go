package providers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts
	MaxRetries int
	// InitialDelay is the initial delay between retries
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration
	// Multiplier is the backoff multiplier
	Multiplier float64
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 500 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// IsRetryableError determines if an error should be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Network errors (retryable)
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "temporary failure") ||
		strings.Contains(errStr, "no such host") {
		return true
	}

	// HTTP status codes in error message (retryable: 429, 500, 502, 503, 504)
	if strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "500") ||
		strings.Contains(errStr, "502") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "504") {
		return true
	}

	// Client errors (not retryable: 400, 401, 403, 404)
	if strings.Contains(errStr, "400") ||
		strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "404") {
		return false
	}

	return false
}

// IsRetryableStatusCode checks if HTTP status code should be retried
func IsRetryableStatusCode(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusInternalServerError ||
		statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable ||
		statusCode == http.StatusGatewayTimeout
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// WithRetry executes a function with retry logic
func WithRetry(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryableError(err) {
			return fmt.Errorf("non-retryable error: %w", err)
		}

		// Check if we've exhausted retries
		if attempt == config.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt)))
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Wait with context
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}

// RetryableReviewFunc is a review function that can be retried
type RetryableReviewFunc func() (*ReviewResponse, error)

// WithRetryReview executes a review function with retry logic
func WithRetryReview(ctx context.Context, config RetryConfig, fn RetryableReviewFunc) (*ReviewResponse, error) {
	var lastErr error
	var result *ReviewResponse

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		var err error
		result, err = fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Check if error is retryable
		if !IsRetryableError(err) {
			return nil, fmt.Errorf("non-retryable error: %w", err)
		}

		// Check if we've exhausted retries
		if attempt == config.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff
		delay := time.Duration(float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt)))
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Wait with context
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return nil, fmt.Errorf("max retries (%d) exceeded: %w", config.MaxRetries, lastErr)
}
