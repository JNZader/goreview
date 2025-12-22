// Package rules provides custom rule definitions for code review.
package rules

// Rule defines a custom review rule.
type Rule interface {
	// ID returns the unique identifier for this rule.
	ID() string

	// Name returns the human-readable name of the rule.
	Name() string

	// Description returns a detailed description of what the rule checks.
	Description() string

	// Severity returns the default severity level.
	Severity() string

	// Enabled returns whether the rule is enabled by default.
	Enabled() bool
}
