package providers

import (
	"fmt"
	"strings"
)

// Validation constants
const (
	// MaxDiffSize is the maximum allowed diff size in bytes (500KB)
	MaxDiffSize = 500 * 1024
	// MaxFilePathLength is the maximum file path length
	MaxFilePathLength = 500
	// MaxLanguageLength is the maximum language name length
	MaxLanguageLength = 50
	// MaxContextLength is the maximum context length
	MaxContextLength = 10000
)

// Error message formats (SonarQube S1192)
const (
	errTooLongChars  = "too long: %d chars (max %d)"
	errTooLargeBytes = "too large: %d bytes (max %d)"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

// ValidateReviewRequest validates a review request before sending to AI provider
func ValidateReviewRequest(req *ReviewRequest) error {
	if req == nil {
		return &ValidationError{Field: "request", Message: "cannot be nil"}
	}

	// Empty diff is valid - returns empty response
	if len(req.Diff) == 0 {
		return nil
	}

	// Check diff size
	if len(req.Diff) > MaxDiffSize {
		return &ValidationError{
			Field:   "diff",
			Message: fmt.Sprintf(errTooLargeBytes, len(req.Diff), MaxDiffSize),
		}
	}

	// Validate file path (prevent path traversal)
	if req.FilePath != "" {
		if len(req.FilePath) > MaxFilePathLength {
			return &ValidationError{
				Field:   "file_path",
				Message: fmt.Sprintf(errTooLongChars, len(req.FilePath), MaxFilePathLength),
			}
		}
		if strings.Contains(req.FilePath, "..") {
			return &ValidationError{
				Field:   "file_path",
				Message: "contains path traversal pattern",
			}
		}
	}

	// Validate language
	if len(req.Language) > MaxLanguageLength {
		return &ValidationError{
			Field:   "language",
			Message: fmt.Sprintf(errTooLongChars, len(req.Language), MaxLanguageLength),
		}
	}

	// Validate context
	if len(req.Context) > MaxContextLength {
		return &ValidationError{
			Field:   "context",
			Message: fmt.Sprintf(errTooLongChars, len(req.Context), MaxContextLength),
		}
	}

	return nil
}

// ValidateDiff validates a diff string for commit message generation
func ValidateDiff(diff string) error {
	if len(diff) == 0 {
		return &ValidationError{Field: "diff", Message: "cannot be empty"}
	}
	if len(diff) > MaxDiffSize {
		return &ValidationError{
			Field:   "diff",
			Message: fmt.Sprintf(errTooLargeBytes, len(diff), MaxDiffSize),
		}
	}
	return nil
}
