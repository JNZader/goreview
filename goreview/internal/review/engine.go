package review

import (
	"context"
	"fmt"
	"runtime"
	"sync"
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
	TotalIssues int           `json:"total_issues"`
	Duration    time.Duration `json:"duration"`
	Files       []FileResult  `json:"files"`
	Stats       git.DiffStats `json:"stats"`
	Summary     string        `json:"summary,omitempty"`
}

// FileResult contains review results for a single file.
type FileResult struct {
	File     string                    `json:"file"`
	Response *providers.ReviewResponse `json:"response,omitempty"`
	Error    error                     `json:"error,omitempty"`
	Cached   bool                      `json:"cached"`
}

// Run executes the review process.
func (e *Engine) Run(ctx context.Context) (*Result, error) {
	// 1. Get diff based on mode
	diff, err := e.getDiff(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	if len(diff.Files) == 0 {
		return &Result{Summary: "No changes found to review."}, nil
	}

	// 2. Filter files to review
	filesToReview := e.filterFiles(diff.Files)
	if len(filesToReview) == 0 {
		return &Result{Summary: "No reviewable files in changes."}, nil
	}

	// 3. Process files concurrently
	concurrency := e.calculateOptimalConcurrency()
	semaphore := make(chan struct{}, concurrency)
	resultsChan := make(chan *FileResult, len(filesToReview))

	var wg sync.WaitGroup
	start := time.Now()

	for _, file := range filesToReview {
		wg.Add(1)
		go func(f git.FileDiff) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := e.reviewFile(ctx, f)
			resultsChan <- result
		}(file)
	}

	// Wait and close
	wg.Wait()
	close(resultsChan)

	// 4. Aggregate results
	finalResult := &Result{
		Stats:    diff.Stats,
		Files:    make([]FileResult, 0, len(filesToReview)),
		Duration: time.Since(start),
	}

	for result := range resultsChan {
		finalResult.Files = append(finalResult.Files, *result)
		if result.Response != nil {
			finalResult.TotalIssues += len(result.Response.Issues)
		}
	}

	return finalResult, nil
}

func (e *Engine) getDiff(ctx context.Context) (*git.Diff, error) {
	switch e.cfg.Review.Mode {
	case "staged":
		return e.gitRepo.GetStagedDiff(ctx)
	case "commit":
		return e.gitRepo.GetCommitDiff(ctx, e.cfg.Review.Commit)
	case "branch":
		return e.gitRepo.GetBranchDiff(ctx, e.cfg.Git.BaseBranch)
	case "files":
		return e.gitRepo.GetFileDiff(ctx, e.cfg.Review.Files)
	default:
		return nil, fmt.Errorf("unknown review mode: %s", e.cfg.Review.Mode)
	}
}

func (e *Engine) filterFiles(files []git.FileDiff) []git.FileDiff {
	result := make([]git.FileDiff, 0, len(files))
	for _, f := range files {
		// Skip deleted and binary files
		if f.Status == git.FileDeleted || f.IsBinary {
			continue
		}
		// Skip ignored patterns
		if e.shouldIgnore(f.Path) {
			continue
		}
		result = append(result, f)
	}
	return result
}

func (e *Engine) shouldIgnore(path string) bool {
	for _, pattern := range e.cfg.Git.IgnorePatterns {
		if matchPattern(pattern, path) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, path string) bool {
	// Simple glob matching
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(path) >= len(prefix) && path[:len(prefix)] == prefix
	}
	return pattern == path
}

func (e *Engine) calculateOptimalConcurrency() int {
	if e.cfg.Review.MaxConcurrency > 0 {
		return e.cfg.Review.MaxConcurrency
	}

	// Auto-detect based on CPU cores
	cpus := runtime.NumCPU()
	optimal := cpus * 2
	if optimal > 10 {
		optimal = 10
	}
	if optimal < 1 {
		optimal = 1
	}
	return optimal
}

func (e *Engine) reviewFile(ctx context.Context, file git.FileDiff) *FileResult {
	// Build review request
	req := &providers.ReviewRequest{
		Diff:     formatDiff(file),
		Language: file.Language,
		FilePath: file.Path,
	}

	// Check cache
	if e.cache != nil {
		key := e.cache.ComputeKey(req)
		if cached, found, _ := e.cache.Get(key); found {
			return &FileResult{
				File:     file.Path,
				Response: cached,
				Cached:   true,
			}
		}
	}

	// Call provider
	resp, err := e.provider.Review(ctx, req)
	if err != nil {
		return &FileResult{
			File: file.Path,
			Error: fmt.Errorf("review failed for %s (lang=%s, size=%d bytes): %w",
				file.Path, file.Language, len(req.Diff), err),
		}
	}

	// Store in cache
	if e.cache != nil {
		key := e.cache.ComputeKey(req)
		_ = e.cache.Set(key, resp)
	}

	return &FileResult{
		File:     file.Path,
		Response: resp,
		Cached:   false,
	}
}

func formatDiff(file git.FileDiff) string {
	var result string
	for _, hunk := range file.Hunks {
		result += hunk.Header + "\n"
		for _, line := range hunk.Lines {
			prefix := " "
			if line.Type == git.LineAddition {
				prefix = "+"
			} else if line.Type == git.LineDeletion {
				prefix = "-"
			}
			result += prefix + line.Content + "\n"
		}
	}
	return result
}
