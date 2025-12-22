package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitRepository implements Repository using git commands.
type GitRepository struct {
	path string
}

// NewGitRepository creates a new GitRepository.
func NewGitRepository(path string) (*GitRepository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify it's a git repository
	repo := &GitRepository{path: absPath}
	if _, err := repo.GetRepoRoot(context.Background()); err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	return repo, nil
}

// runGit executes a git command and returns the output.
func (r *GitRepository) runGit(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.path

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Include stderr in error message for debugging
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg != "" {
			return "", fmt.Errorf("git %s: %w: %s", args[0], err, errMsg)
		}
		return "", fmt.Errorf("git %s: %w", args[0], err)
	}

	return stdout.String(), nil
}

func (r *GitRepository) GetStagedDiff(ctx context.Context) (*Diff, error) {
	// Get staged diff
	output, err := r.runGit(ctx, "diff", "--cached", "--unified=3")
	if err != nil {
		return nil, err
	}

	diff, err := ParseDiff(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	return diff, nil
}

func (r *GitRepository) GetCommitDiff(ctx context.Context, sha string) (*Diff, error) {
	output, err := r.runGit(ctx, "show", sha, "--unified=3", "--format=")
	if err != nil {
		return nil, err
	}

	return ParseDiff(output)
}

func (r *GitRepository) GetBranchDiff(ctx context.Context, baseBranch string) (*Diff, error) {
	// Get merge base
	mergeBase, err := r.runGit(ctx, "merge-base", baseBranch, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to find merge base: %w", err)
	}

	mergeBase = strings.TrimSpace(mergeBase)
	output, err := r.runGit(ctx, "diff", mergeBase, "HEAD", "--unified=3")
	if err != nil {
		return nil, err
	}

	return ParseDiff(output)
}

func (r *GitRepository) GetFileDiff(ctx context.Context, files []string) (*Diff, error) {
	args := append([]string{"diff", "--unified=3", "--"}, files...)
	output, err := r.runGit(ctx, args...)
	if err != nil {
		return nil, err
	}

	return ParseDiff(output)
}

func (r *GitRepository) GetCurrentBranch(ctx context.Context) (string, error) {
	output, err := r.runGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *GitRepository) GetRepoRoot(ctx context.Context) (string, error) {
	output, err := r.runGit(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *GitRepository) IsClean(ctx context.Context) (bool, error) {
	output, err := r.runGit(ctx, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) == "", nil
}
