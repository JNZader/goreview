package review

import (
	"context"
	"testing"

	"github.com/JNZader/goreview/goreview/internal/config"
	"github.com/JNZader/goreview/goreview/internal/git"
	"github.com/JNZader/goreview/goreview/internal/providers"
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

func TestEngineFilterDeletedFiles(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Review.Mode = "staged"

	repo := &MockRepository{
		StagedDiff: &git.Diff{
			Files: []git.FileDiff{
				{Path: "deleted.go", Status: git.FileDeleted},
				{Path: "modified.go", Status: git.FileModified, Language: "go"},
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
		t.Errorf("len(Files) = %d, want 1 (deleted should be filtered)", len(result.Files))
	}
}

func TestEngineFilterBinaryFiles(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Review.Mode = "staged"

	repo := &MockRepository{
		StagedDiff: &git.Diff{
			Files: []git.FileDiff{
				{Path: "image.png", IsBinary: true, Status: git.FileAdded},
				{Path: "code.go", Status: git.FileModified, Language: "go"},
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
		t.Errorf("len(Files) = %d, want 1 (binary should be filtered)", len(result.Files))
	}
}

func TestCalculateOptimalConcurrency(t *testing.T) {
	cfg := config.DefaultConfig()
	engine := NewEngine(cfg, nil, nil, nil, nil)

	// Test auto-detection (should be > 0)
	concurrency := engine.calculateOptimalConcurrency()
	if concurrency < 1 {
		t.Errorf("calculateOptimalConcurrency() = %d, want >= 1", concurrency)
	}

	// Test configured value
	cfg.Review.MaxConcurrency = 3
	concurrency = engine.calculateOptimalConcurrency()
	if concurrency != 3 {
		t.Errorf("calculateOptimalConcurrency() = %d, want 3", concurrency)
	}
}

func TestMatchPattern(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		{"vendor/*", "vendor/lib/file.go", true},
		{"vendor/*", "src/file.go", false},
		{"*.md", "*.md", true},
		{"README.md", "README.md", true},
		{"README.md", "CHANGELOG.md", false},
	}

	for _, tt := range tests {
		got := matchPattern(tt.pattern, tt.path)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}
