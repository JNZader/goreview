package review

import (
	"time"

	"github.com/JNZader/goreview/goreview/internal/cache"
	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/git"
	"github.com/JNZader/goreview/goreview/internal/providers"
	"github.com/JNZader/goreview/goreview/internal/rules"
)

const DefaultMaxConcurrency = 5

// Engine orchestrates the code review process.
type Engine struct {
	cfg      *config.Config
	gitRepo  git.Repository
	provider providers.Provider
	cache    cache.Cache
	rules    []rules.Rule
}

// NewEngine creates a new review engine.
func NewEngine(
	cfg *config.Config,
	gitRepo git.Repository,
	provider providers.Provider,
	c cache.Cache,
	r []rules.Rule,
) *Engine {
	return &Engine{
		cfg:      cfg,
		gitRepo:  gitRepo,
		provider: provider,
		cache:    c,
		rules:    r,
	}
}

// Result contains the complete review results.
type Result struct {
	TotalIssues int             `json:"total_issues"`
	Duration    time.Duration   `json:"duration"`
	Files       []FileResult    `json:"files"`
	Stats       git.DiffStats   `json:"stats"`
	Summary     string          `json:"summary,omitempty"`
}

// FileResult contains review results for a single file.
type FileResult struct {
	File     string                    `json:"file"`
	Response *providers.ReviewResponse `json:"response,omitempty"`
	Error    error                     `json:"error,omitempty"`
	Cached   bool                      `json:"cached"`
}
