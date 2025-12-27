---
name: goreview-workflow
description: Code review and quality workflow using GoReview CLI. Auto-invoked when reviewing code, checking quality, analyzing changes, or preparing commits. Provides structured review process with security, performance, and best practices analysis.
allowed-tools: Read, Bash(goreview:*, git:*)
---

# GoReview Workflow

Use GoReview CLI for systematic code analysis and quality assurance.

## Quick Commands

### Review Staged Changes
```bash
goreview review --staged
```

### Security-Focused Review
```bash
goreview review --staged --mode=security --trace
```

### Performance Analysis
```bash
goreview review --staged --mode=perf
```

### Generate Commit Message
```bash
goreview commit
```

### Auto-Fix Issues
```bash
goreview fix --staged
```

## Review Modes

| Mode | Focus |
|------|-------|
| `security` | OWASP vulnerabilities, secrets, injections |
| `perf` | N+1 queries, complexity, memory leaks |
| `clean` | SOLID, DRY, naming, code smells |
| `docs` | Missing comments, documentation |
| `tests` | Coverage, edge cases, mocking |

## Reviewer Personalities

| Personality | Style |
|-------------|-------|
| `senior` | Mentoring, explains "why" |
| `strict` | Direct, demanding |
| `friendly` | Encouraging, positive |
| `security-expert` | Paranoid, worst-case |

## Combined Usage

```bash
# Full review with security focus and root cause tracing
goreview review --staged --mode=security,perf --personality=senior --trace

# TDD enforcement
goreview review --staged --require-tests --min-coverage=80

# Export to Obsidian
goreview review --staged --export-obsidian
```

## Output Formats

- `markdown` - Human readable (default)
- `json` - Structured for processing
- `sarif` - IDE integration

## Integration Tips

1. Run review before committing
2. Use `--trace` for complex issues
3. Check `goreview stats` for patterns
4. Search history with `goreview search`
