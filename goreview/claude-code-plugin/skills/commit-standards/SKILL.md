---
name: commit-standards
description: Conventional Commits workflow using GoReview. Auto-invoked when creating commits, writing commit messages, or preparing changes for version control.
allowed-tools: Bash(goreview:commit*, git:*)
---

# Commit Standards

Use GoReview to generate standardized commit messages following Conventional Commits.

## Generate Commit Message

```bash
goreview commit
```

This analyzes staged changes and generates a message like:
```
feat(api): add user authentication endpoint

- Implement JWT token generation
- Add password hashing with bcrypt
- Create login/logout endpoints
```

## Commit Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `style` | Formatting (no code change) |
| `refactor` | Code restructuring |
| `test` | Adding tests |
| `chore` | Maintenance |

## Options

```bash
# Force specific type
goreview commit --type feat

# Add scope
goreview commit --scope api

# Mark as breaking change
goreview commit --breaking

# Execute commit directly
goreview commit --execute

# Amend last commit
goreview commit --amend
```

## Best Practices

1. Keep subject under 50 characters
2. Use imperative mood ("add" not "added")
3. Explain "why" in the body
4. Reference issues when applicable
5. One logical change per commit
