---
description: Review staged code changes with AI
---

Run GoReview on staged changes to identify issues:

```bash
goreview review --staged --format markdown
```

If you want a specific focus, use mode flags:
- Security: `goreview review --staged --mode=security`
- Performance: `goreview review --staged --mode=perf`
- Clean code: `goreview review --staged --mode=clean`
- Documentation: `goreview review --staged --mode=docs`
- Tests: `goreview review --staged --mode=tests`

For root cause analysis, add `--trace`.
