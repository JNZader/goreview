package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RegisterGoReviewTools registers all GoReview tools with the MCP server.
func RegisterGoReviewTools(s *Server) {
	registerReviewTools(s)
	registerUtilityTools(s)
	registerDocTools(s)
}

// registerReviewTools registers review and fix related tools.
func registerReviewTools(s *Server) {
	s.RegisterTool(&Tool{
		Name:        "goreview_review",
		Description: "Analyze code changes and identify issues. Supports staged changes, specific commits, or branch comparisons.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target": map[string]interface{}{
					"type":        "string",
					"description": "What to review: 'staged', 'HEAD', a commit SHA, or a branch name",
					"default":     "staged",
				},
				"mode": map[string]interface{}{
					"type":        "string",
					"description": "Review mode: security, perf, clean, docs, tests, or comma-separated combination",
					"enum":        []string{"security", "perf", "clean", "docs", "tests"},
				},
				"personality": map[string]interface{}{
					"type":        "string",
					"description": "Reviewer personality: senior, strict, friendly, security-expert",
					"enum":        []string{"senior", "strict", "friendly", "security-expert"},
				},
				"files": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Specific files to review (optional)",
				},
				"trace": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable root cause tracing",
					"default":     false,
				},
			},
		},
	}, handleReview)

	s.RegisterTool(&Tool{
		Name:        "goreview_fix",
		Description: "Automatically fix issues found in code review.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target": map[string]interface{}{
					"type":        "string",
					"description": "What to fix: 'staged' or specific file path",
					"default":     "staged",
				},
				"severity": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Only fix issues with these severities",
				},
				"types": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Only fix issues of these types",
				},
				"dryRun": map[string]interface{}{
					"type":        "boolean",
					"description": "Show what would be fixed without applying",
					"default":     false,
				},
			},
		},
	}, handleFix)
}

// registerUtilityTools registers commit, search, and stats tools.
func registerUtilityTools(s *Server) {
	s.RegisterTool(&Tool{
		Name:        "goreview_commit",
		Description: "Generate an AI-powered commit message following Conventional Commits format.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Force commit type (feat, fix, docs, style, refactor, test, chore)",
					"enum":        []string{"feat", "fix", "docs", "style", "refactor", "test", "chore"},
				},
				"scope": map[string]interface{}{
					"type":        "string",
					"description": "Force commit scope",
				},
				"breaking": map[string]interface{}{
					"type":        "boolean",
					"description": "Mark as breaking change",
					"default":     false,
				},
			},
		},
	}, handleCommit)

	s.RegisterTool(&Tool{
		Name:        "goreview_search",
		Description: "Search through past code reviews and findings.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query (full-text search)",
				},
				"severity": map[string]interface{}{
					"type":        "string",
					"description": "Filter by severity",
					"enum":        []string{"critical", "error", "warning", "info"},
				},
				"file": map[string]interface{}{
					"type":        "string",
					"description": "Filter by file path pattern",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum results to return",
					"default":     10,
				},
			},
			"required": []string{"query"},
		},
	}, handleSearch)

	s.RegisterTool(&Tool{
		Name:        "goreview_stats",
		Description: "Get code review statistics and project health metrics.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"period": map[string]interface{}{
					"type":        "string",
					"description": "Time period: today, week, month, all",
					"default":     "week",
					"enum":        []string{"today", "week", "month", "all"},
				},
				"groupBy": map[string]interface{}{
					"type":        "string",
					"description": "Group results by: file, severity, type, author",
					"enum":        []string{"file", "severity", "type", "author"},
				},
			},
		},
	}, handleStats)
}

// registerDocTools registers changelog and documentation tools.
func registerDocTools(s *Server) {
	s.RegisterTool(&Tool{
		Name:        "goreview_changelog",
		Description: "Generate changelog from git commits.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"from": map[string]interface{}{
					"type":        "string",
					"description": "Starting point (tag or commit)",
				},
				"to": map[string]interface{}{
					"type":        "string",
					"description": "Ending point (default: HEAD)",
					"default":     "HEAD",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format: markdown, json",
					"default":     "markdown",
					"enum":        []string{"markdown", "json"},
				},
			},
		},
	}, handleChangelog)

	s.RegisterTool(&Tool{
		Name:        "goreview_doc",
		Description: "Generate documentation for code changes.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"target": map[string]interface{}{
					"type":        "string",
					"description": "What to document: 'staged', commit SHA, or file path",
					"default":     "staged",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Documentation type: changes, changelog, api, readme",
					"default":     "changes",
					"enum":        []string{"changes", "changelog", "api", "readme"},
				},
				"style": map[string]interface{}{
					"type":        "string",
					"description": "Documentation style: markdown, jsdoc, godoc",
					"default":     "markdown",
					"enum":        []string{"markdown", "jsdoc", "godoc"},
				},
			},
		},
	}, handleDoc)
}

// Tool handlers

func handleReview(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	args := []string{"review"}
	args = appendTargetArgs(args, params)
	args = appendReviewOptions(args, params)
	args = append(args, "--format", "json")
	return runGoReview(ctx, args)
}

// appendTargetArgs appends target-related arguments based on params.
func appendTargetArgs(args []string, params map[string]interface{}) []string {
	target, _ := params["target"].(string)
	switch {
	case target == "" || target == "staged":
		return append(args, "--staged")
	case target == "HEAD":
		return append(args, "--commit", "HEAD")
	case strings.HasPrefix(target, "origin/") || !strings.Contains(target, "/") && len(target) < 40:
		return append(args, "--branch", target)
	default:
		return append(args, "--commit", target)
	}
}

// appendReviewOptions appends optional review arguments.
func appendReviewOptions(args []string, params map[string]interface{}) []string {
	if mode, ok := params["mode"].(string); ok && mode != "" {
		args = append(args, "--mode", mode)
	}
	if personality, ok := params["personality"].(string); ok && personality != "" {
		args = append(args, "--personality", personality)
	}
	if trace, ok := params["trace"].(bool); ok && trace {
		args = append(args, "--trace")
	}
	if files, ok := params["files"].([]interface{}); ok {
		for _, f := range files {
			if fs, ok := f.(string); ok {
				args = append(args, fs)
			}
		}
	}
	return args
}

func handleCommit(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	args := []string{"commit"}

	if commitType, ok := params["type"].(string); ok && commitType != "" {
		args = append(args, "--type", commitType)
	}

	if scope, ok := params["scope"].(string); ok && scope != "" {
		args = append(args, "--scope", scope)
	}

	if breaking, ok := params["breaking"].(bool); ok && breaking {
		args = append(args, "--breaking")
	}

	return runGoReview(ctx, args)
}

func handleFix(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	args := []string{"fix"}

	target, _ := params["target"].(string)
	if target == "" || target == "staged" {
		args = append(args, "--staged")
	} else {
		args = append(args, target)
	}

	if dryRun, ok := params["dryRun"].(bool); ok && dryRun {
		args = append(args, "--dry-run")
	}

	if severities, ok := params["severity"].([]interface{}); ok {
		for _, s := range severities {
			if ss, ok := s.(string); ok {
				args = append(args, "--severity", ss)
			}
		}
	}

	if types, ok := params["types"].([]interface{}); ok {
		for _, t := range types {
			if ts, ok := t.(string); ok {
				args = append(args, "--type", ts)
			}
		}
	}

	return runGoReview(ctx, args)
}

func handleSearch(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	args := []string{"search"}

	if query, ok := params["query"].(string); ok && query != "" {
		args = append(args, query)
	}

	if severity, ok := params["severity"].(string); ok && severity != "" {
		args = append(args, "--severity", severity)
	}

	if file, ok := params["file"].(string); ok && file != "" {
		args = append(args, "--file", file)
	}

	if limit, ok := params["limit"].(float64); ok {
		args = append(args, "--limit", fmt.Sprintf("%d", int(limit)))
	}

	args = append(args, "--format", "json")

	return runGoReview(ctx, args)
}

func handleStats(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	args := []string{"stats"}

	if period, ok := params["period"].(string); ok && period != "" {
		args = append(args, "--period", period)
	}

	if groupBy, ok := params["groupBy"].(string); ok && groupBy != "" {
		args = append(args, "--by-"+groupBy)
	}

	args = append(args, "--format", "json")

	return runGoReview(ctx, args)
}

func handleChangelog(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	args := []string{"changelog"}

	if from, ok := params["from"].(string); ok && from != "" {
		args = append(args, "--from", from)
	}

	if to, ok := params["to"].(string); ok && to != "" {
		args = append(args, "--to", to)
	}

	if format, ok := params["format"].(string); ok && format != "" {
		args = append(args, "--format", format)
	}

	return runGoReview(ctx, args)
}

func handleDoc(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	args := []string{"doc"}

	target, _ := params["target"].(string)
	if target == "" || target == "staged" {
		args = append(args, "--staged")
	} else {
		args = append(args, target)
	}

	if docType, ok := params["type"].(string); ok && docType != "" {
		args = append(args, "--type", docType)
	}

	if style, ok := params["style"].(string); ok && style != "" {
		args = append(args, "--style", style)
	}

	return runGoReview(ctx, args)
}

// runGoReview executes the goreview binary with the given arguments.
func runGoReview(ctx context.Context, args []string) (interface{}, error) {
	// Find goreview binary
	binary, err := findGoReviewBinary()
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, binary, args...) // #nosec G204 - binary path is validated by findGoReviewBinary
	cmd.Dir, _ = os.Getwd()

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("goreview error: %s", stderr.String())
		}
		return nil, fmt.Errorf("goreview failed: %w", err)
	}

	// Try to parse as JSON for structured output
	output := stdout.String()
	var jsonResult interface{}
	if err := json.Unmarshal([]byte(output), &jsonResult); err == nil {
		return jsonResult, nil
	}

	return output, nil
}

// findGoReviewBinary finds the goreview binary path.
func findGoReviewBinary() (string, error) {
	// Check if we're running as the goreview binary itself
	exe, err := os.Executable()
	if err == nil {
		base := filepath.Base(exe)
		if strings.HasPrefix(base, "goreview") {
			return exe, nil
		}
	}

	// Look in PATH
	if path, err := exec.LookPath("goreview"); err == nil {
		return path, nil
	}

	// Check common locations
	homeDir, _ := os.UserHomeDir()
	locations := []string{
		filepath.Join(homeDir, "go", "bin", "goreview"),
		filepath.Join(homeDir, ".local", "bin", "goreview"),
		"/usr/local/bin/goreview",
		"/usr/bin/goreview",
	}

	for _, loc := range locations {
		if _, err := os.Stat(loc); err == nil {
			return loc, nil
		}
	}

	return "", fmt.Errorf("goreview binary not found in PATH or common locations")
}
