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

// retryablePatterns contains error patterns that should be retried
var retryablePatterns = []string{
	"connection refused",
	"connection reset",
	"timeout",
	"temporary failure",
	"no such host",
	"429", // Too Many Requests
	"500", // Internal Server Error
	"502", // Bad Gateway
	"503", // Service Unavailable
	"504", // Gateway Timeout
}

// nonRetryablePatterns contains error patterns that should NOT be retried
var nonRetryablePatterns = []string{
	"400", // Bad Request
	"401", // Unauthorized
	"403", // Forbidden
	"404", // Not Found
}

// IsRetryableError determines if an error should be retried
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check non-retryable patterns first
	for _, pattern := range nonRetryablePatterns {
		if strings.Contains(errStr, pattern) {
			return false
		}
	}

	// Check retryable patterns
	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
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
