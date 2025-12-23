# Iteracion 05: Motor de Review

## Objetivos

- Engine como orquestador principal
- Procesamiento paralelo de archivos
- Integracion con Git, Provider y Cache
- Control de concurrencia optimo
- Agregacion de resultados

## Tiempo Estimado: 8 horas

---

## Commit 5.1: Crear estructura del Engine

**Mensaje de commit:**
```
feat(review): add engine struct

- Create Engine as main orchestrator
- Inject dependencies (git, provider, cache, rules)
- Define Result and FileResult types
```

### `goreview/internal/review/engine.go`

```go
package review

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/cache"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/git"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/rules"
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
```

---

## Commit 5.2: Implementar Run con concurrencia

**Mensaje de commit:**
```
feat(review): add concurrent file processing

- Process files in parallel with semaphore
- Calculate optimal concurrency from CPU
- Aggregate results from goroutines
- Handle errors per file
```

### Agregar a `goreview/internal/review/engine.go`:

```go
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
	var result []git.FileDiff
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
	if pattern[len(pattern)-1] == '*' {
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
			File:  file.Path,
			Error: err,
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
```

---

## Commit 5.3: Agregar tests del Engine

**Mensaje de commit:**
```
test(review): add engine tests

- Test with mock provider
- Test concurrent processing
- Test cache integration
- Test error handling
```

### `goreview/internal/review/engine_test.go`

```go
package review

import (
	"context"
	"testing"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/git"
	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/providers"
)

// MockProvider for testing
type MockProvider struct {
	ReviewFunc func(ctx context.Context, req *providers.ReviewRequest) (*providers.ReviewResponse, error)
}

func (m *MockProvider) Name() string { return "mock" }
func (m *MockProvider) Review(ctx context.Context, req *providers.ReviewRequest) (*providers.ReviewResponse, error) {
	if m.ReviewFunc != nil {
		return m.ReviewFunc(ctx, req)
	}
	return &providers.ReviewResponse{
		Issues:  []providers.Issue{{ID: "1", Message: "Test issue"}},
		Summary: "Test summary",
		Score:   85,
	}, nil
}
func (m *MockProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	return "feat: test", nil
}
func (m *MockProvider) GenerateDocumentation(ctx context.Context, diff, context string) (string, error) {
	return "# Doc", nil
}
func (m *MockProvider) HealthCheck(ctx context.Context) error { return nil }
func (m *MockProvider) Close() error                          { return nil }

// MockRepository for testing
type MockRepository struct {
	StagedDiff *git.Diff
}

func (m *MockRepository) GetStagedDiff(ctx context.Context) (*git.Diff, error) {
	return m.StagedDiff, nil
}
func (m *MockRepository) GetCommitDiff(ctx context.Context, sha string) (*git.Diff, error) {
	return m.StagedDiff, nil
}
func (m *MockRepository) GetBranchDiff(ctx context.Context, base string) (*git.Diff, error) {
	return m.StagedDiff, nil
}
func (m *MockRepository) GetFileDiff(ctx context.Context, files []string) (*git.Diff, error) {
	return m.StagedDiff, nil
}
func (m *MockRepository) GetCurrentBranch(ctx context.Context) (string, error) { return "main", nil }
func (m *MockRepository) GetRepoRoot(ctx context.Context) (string, error)      { return "/repo", nil }
func (m *MockRepository) IsClean(ctx context.Context) (bool, error)            { return true, nil }

func TestEngineRun(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Review.Mode = "staged"

	repo := &MockRepository{
		StagedDiff: &git.Diff{
			Files: []git.FileDiff{
				{Path: "main.go", Language: "go", Status: git.FileModified},
			},
		},
	}

	provider := &MockProvider{}
	engine := NewEngine(cfg, repo, provider, nil, nil)

	result, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if len(result.Files) != 1 {
		t.Errorf("len(Files) = %d, want 1", len(result.Files))
	}
}

func TestEngineEmptyDiff(t *testing.T) {
	cfg := config.DefaultConfig()
	repo := &MockRepository{StagedDiff: &git.Diff{}}
	provider := &MockProvider{}

	engine := NewEngine(cfg, repo, provider, nil, nil)
	result, err := engine.Run(context.Background())

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	if result.Summary != "No changes found to review." {
		t.Errorf("Summary = %q, want empty diff message", result.Summary)
	}
}
```

---

## Resumen de la Iteracion 05

### Commits:
1. `feat(review): add engine struct`
2. `feat(review): add concurrent file processing`
3. `test(review): add engine tests`

### Archivos:
```
goreview/internal/review/
├── engine.go
└── engine_test.go
```

---

## Siguiente Iteracion

Continua con: **[06-SISTEMA-CACHE.md](06-SISTEMA-CACHE.md)**
