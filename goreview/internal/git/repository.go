package git

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Git command constants (SonarQube S1192)
const (
	unifiedContextFlag = "--unified=3"
)

// Repo implements Repository using git commands.
type Repo struct {
	path string
}

// NewRepo creates a new Repo.
func NewRepo(path string) (*Repo, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Verify it's a git repository
	repo := &Repo{path: absPath}
	if _, err := repo.GetRepoRoot(context.Background()); err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}

	return repo, nil
}

// runGit executes a git command and returns the output.
func (r *Repo) runGit(ctx context.Context, args ...string) (string, error) {
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

func (r *Repo) GetStagedDiff(ctx context.Context) (*Diff, error) {
	// Get staged diff
	output, err := r.runGit(ctx, "diff", "--cached", unifiedContextFlag)
	if err != nil {
		return nil, err
	}

	diff, err := ParseDiff(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse diff: %w", err)
	}

	return diff, nil
}

func (r *Repo) GetCommitDiff(ctx context.Context, sha string) (*Diff, error) {
	output, err := r.runGit(ctx, "show", sha, unifiedContextFlag, "--format=")
	if err != nil {
		return nil, err
	}

	return ParseDiff(output)
}

func (r *Repo) GetBranchDiff(ctx context.Context, baseBranch string) (*Diff, error) {
	// Get merge base
	mergeBase, err := r.runGit(ctx, "merge-base", baseBranch, "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to find merge base: %w", err)
	}

	mergeBase = strings.TrimSpace(mergeBase)
	output, err := r.runGit(ctx, "diff", mergeBase, "HEAD", unifiedContextFlag)
	if err != nil {
		return nil, err
	}

	return ParseDiff(output)
}

func (r *Repo) GetFileDiff(ctx context.Context, files []string) (*Diff, error) {
	args := append([]string{"diff", unifiedContextFlag, "--"}, files...)
	output, err := r.runGit(ctx, args...)
	if err != nil {
		return nil, err
	}

	return ParseDiff(output)
}

func (r *Repo) GetCurrentBranch(ctx context.Context) (string, error) {
	output, err := r.runGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *Repo) GetRepoRoot(ctx context.Context) (string, error) {
	output, err := r.runGit(ctx, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

func (r *Repo) IsClean(ctx context.Context) (bool, error) {
	output, err := r.runGit(ctx, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(output) == "", nil
}

// GetCommits returns commits between two refs (or all commits if from is empty).
func (r *Repo) GetCommits(ctx context.Context, from, to string) ([]Commit, error) {
	// Format: hash|short_hash|subject|body|author|email|date
	format := "%H|%h|%s|%b|%an|%ae|%aI"
	separator := "---COMMIT_SEPARATOR---"

	var args []string
	if from != "" && to != "" {
		args = []string{"log", from + ".." + to, "--format=" + format + separator, "--no-merges"}
	} else if from != "" {
		args = []string{"log", from + "..HEAD", "--format=" + format + separator, "--no-merges"}
	} else if to != "" {
		args = []string{"log", to, "--format=" + format + separator, "--no-merges"}
	} else {
		args = []string{"log", "--format=" + format + separator, "--no-merges"}
	}

	output, err := r.runGit(ctx, args...)
	if err != nil {
		return nil, err
	}

	return parseCommits(output, separator)
}

// GetTags returns all tags sorted by date (newest first).
func (r *Repo) GetTags(ctx context.Context) ([]Tag, error) {
	// Format: refname|hash|date|tagger
	format := "%(refname:short)|%(objectname:short)|%(creatordate:iso-strict)|%(taggername)"
	output, err := r.runGit(ctx, "tag", "--list", "--sort=-creatordate", "--format="+format)
	if err != nil {
		return nil, err
	}

	return parseTags(output)
}

// GetLatestTag returns the most recent tag.
func (r *Repo) GetLatestTag(ctx context.Context) (*Tag, error) {
	tags, err := r.GetTags(ctx)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, nil
	}
	return &tags[0], nil
}

// parseCommits parses the git log output into Commit structs.
func parseCommits(output, separator string) ([]Commit, error) {
	var commits []Commit
	entries := strings.Split(output, separator)

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.SplitN(entry, "|", 7)
		if len(parts) < 7 {
			continue
		}

		commits = append(commits, Commit{
			Hash:        parts[0],
			ShortHash:   parts[1],
			Subject:     parts[2],
			Body:        strings.TrimSpace(parts[3]),
			Author:      parts[4],
			AuthorEmail: parts[5],
			Date:        parts[6],
		})
	}

	return commits, nil
}

// parseTags parses the git tag output into Tag structs.
func parseTags(output string) ([]Tag, error) {
	var tags []Tag
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, "|", 4)
		if len(parts) < 3 {
			continue
		}

		tag := Tag{
			Name: parts[0],
			Hash: parts[1],
			Date: parts[2],
		}
		if len(parts) > 3 {
			tag.Tagger = parts[3]
		}
		tags = append(tags, tag)
	}

	return tags, nil
}
