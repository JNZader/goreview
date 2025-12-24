package review

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/JNZader/goreview/goreview/internal/cache"
	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/git"
	"github.com/JNZader/goreview/goreview/internal/logger"
	"github.com/JNZader/goreview/goreview/internal/providers"
	"github.com/JNZader/goreview/goreview/internal/rules"
	"github.com/JNZader/goreview/goreview/internal/worker"
)

const DefaultMaxConcurrency = 5

// Engine orchestrates the code review process.
type Engine struct {
	cfg      *config.Config
	gitRepo  git.Repository
	provider providers.Provider
	cache    cache.Cache
	rules    []rules.Rule
	log      *logger.Logger
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
		log:      logger.Default().WithPrefix("ENGINE"),
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

// reviewTask implements worker.Task for file reviews
type reviewTask struct {
	id       string
	file     git.FileDiff
	engine   *Engine
	result   *FileResult
	resultMu sync.Mutex
}

func newReviewTask(file git.FileDiff, engine *Engine) *reviewTask {
	return &reviewTask{
		id:     fmt.Sprintf("review:%s", file.Path),
		file:   file,
		engine: engine,
	}
}

func (t *reviewTask) ID() string {
	return t.id
}

func (t *reviewTask) Execute(ctx context.Context) error {
	result := t.engine.reviewFile(ctx, t.file)
	t.resultMu.Lock()
	t.result = result
	t.resultMu.Unlock()
	return result.Error
}

func (t *reviewTask) Result() *FileResult {
	t.resultMu.Lock()
	defer t.resultMu.Unlock()
	return t.result
}

// Run executes the review process using the worker pool.
func (e *Engine) Run(ctx context.Context) (*Result, error) {
	start := time.Now()

	// 1. Get diff based on mode
	e.log.Debug("Getting diff in mode: %s", e.cfg.Review.Mode)
	diff, err := e.getDiff(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	if len(diff.Files) == 0 {
		e.log.Info("No changes found to review")
		return &Result{Summary: "No changes found to review."}, nil
	}

	// 2. Filter files to review
	filesToReview := e.filterFiles(diff.Files)
	if len(filesToReview) == 0 {
		e.log.Info("No reviewable files in changes")
		return &Result{Summary: "No reviewable files in changes."}, nil
	}

	e.log.Info("Reviewing %d files with %d workers", len(filesToReview), e.calculateOptimalConcurrency())

	// 3. Initialize worker pool
	poolCfg := worker.Config{
		Workers:   e.calculateOptimalConcurrency(),
		QueueSize: len(filesToReview),
	}
	pool := worker.NewPool(poolCfg)
	pool.Start()

	// 4. Create and submit tasks
	tasks := make([]*reviewTask, 0, len(filesToReview))
	for _, file := range filesToReview {
		task := newReviewTask(file, e)
		tasks = append(tasks, task)
		if err := pool.Submit(task); err != nil {
			e.log.Error("Failed to submit task for %s: %v", file.Path, err)
		}
	}

	// 5. Collect results
	finalResult := &Result{
		Stats: diff.Stats,
		Files: make([]FileResult, 0, len(filesToReview)),
	}

	// Wait for all results
	resultsCollected := 0
	for resultsCollected < len(tasks) {
		select {
		case result := <-pool.Results():
			resultsCollected++
			// Find the corresponding task and get its result
			for _, task := range tasks {
				if task.ID() == result.TaskID {
					if fileResult := task.Result(); fileResult != nil {
						finalResult.Files = append(finalResult.Files, *fileResult)
						if fileResult.Response != nil {
							finalResult.TotalIssues += len(fileResult.Response.Issues)
						}
						if fileResult.Cached {
							e.log.Debug("Cache hit for %s", fileResult.File)
						}
					}
					break
				}
			}
		case <-ctx.Done():
			e.log.Warn("Review cancelled: %v", ctx.Err())
			pool.Stop()
			return nil, ctx.Err()
		}
	}

	// 6. Stop pool gracefully
	pool.StopWait()

	finalResult.Duration = time.Since(start)

	stats := pool.Stats()
	e.log.Info("Review completed: %d files, %d issues, %d errors in %v",
		len(finalResult.Files), finalResult.TotalIssues, stats.Errors, finalResult.Duration)

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
			e.log.Debug("Ignoring file: %s", f.Path)
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
	return DefaultMaxConcurrency
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
		e.log.Error("Review failed for %s (lang=%s, size=%d bytes): %v",
			file.Path, file.Language, len(req.Diff), err)
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
