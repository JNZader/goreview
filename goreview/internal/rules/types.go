package rules

// Rule defines a code review rule.
type Rule struct {
	ID          string   `yaml:"id" json:"id"`
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Category    Category `yaml:"category" json:"category"`
	Severity    Severity `yaml:"severity" json:"severity"`
	Languages   []string `yaml:"languages" json:"languages"`
	Patterns    []string `yaml:"patterns" json:"patterns"` // File patterns
	Enabled     bool     `yaml:"enabled" json:"enabled"`
	Message     string   `yaml:"message" json:"message"`
	Suggestion  string   `yaml:"suggestion" json:"suggestion"`
}

// Category categorizes rules.
type Category string

const (
	CategorySecurity     Category = "security"
	CategoryPerformance  Category = "performance"
	CategoryBestPractice Category = "best_practice"
	CategoryStyle        Category = "style"
	CategoryBug          Category = "bug"
	CategoryMaintenance  Category = "maintenance"
)

// Severity indicates rule importance.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// RuleSet contains a collection of rules.
type RuleSet struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	Rules       []Rule `yaml:"rules" json:"rules"`
}

// Preset defines a collection of enabled rules.
type Preset struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Includes    []string `yaml:"includes" json:"includes"` // Rule IDs
	Excludes    []string `yaml:"excludes" json:"excludes"` // Rule IDs
}
