package providers

import "context"

// Provider defines the interface for AI/LLM providers.
type Provider interface {
	// Name returns the provider name.
	Name() string

	// Review analyzes code and returns issues.
	Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error)

	// GenerateCommitMessage generates a commit message from diff.
	GenerateCommitMessage(ctx context.Context, diff string) (string, error)

	// GenerateDocumentation generates documentation from diff.
	GenerateDocumentation(ctx context.Context, diff, context string) (string, error)

	// HealthCheck verifies the provider is available.
	HealthCheck(ctx context.Context) error

	// Close releases any resources.
	Close() error
}

// ReviewRequest contains the input for a code review.
type ReviewRequest struct {
	Diff        string       `json:"diff"`
	Language    string       `json:"language"`
	FilePath    string       `json:"file_path"`
	FileContent string       `json:"file_content,omitempty"`
	Context     string       `json:"context,omitempty"`
	Rules       []string     `json:"rules,omitempty"`
	Personality string       `json:"personality,omitempty"`
	Modes       []ReviewMode `json:"modes,omitempty"`
}

// ReviewResponse contains the review results.
type ReviewResponse struct {
	Issues         []Issue `json:"issues"`
	Summary        string  `json:"summary"`
	Score          int     `json:"score"` // 0-100
	TokensUsed     int     `json:"tokens_used"`
	ProcessingTime int64   `json:"processing_time_ms"`
}

// Issue represents a code review issue.
type Issue struct {
	ID         string    `json:"id"`
	Type       IssueType `json:"type"`
	Severity   Severity  `json:"severity"`
	Message    string    `json:"message"`
	Suggestion string    `json:"suggestion,omitempty"`
	Location   *Location `json:"location,omitempty"`
	RuleID     string    `json:"rule_id,omitempty"`
	FixedCode  string    `json:"fixed_code,omitempty"`
}

// Location represents a position in code.
type Location struct {
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	StartCol  int    `json:"start_col,omitempty"`
	EndCol    int    `json:"end_col,omitempty"`
}

// IssueType categorizes the type of issue.
type IssueType string

const (
	IssueTypeBug          IssueType = "bug"
	IssueTypeSecurity     IssueType = "security"
	IssueTypePerformance  IssueType = "performance"
	IssueTypeStyle        IssueType = "style"
	IssueTypeMaintenance  IssueType = "maintenance"
	IssueTypeBestPractice IssueType = "best_practice"
)

// Severity indicates the importance of an issue.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)
