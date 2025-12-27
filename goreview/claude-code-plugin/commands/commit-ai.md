---
description: Generate AI-powered commit message
---

Generate a commit message using GoReview:

```bash
goreview commit
```

The message follows Conventional Commits format (feat, fix, docs, etc.).

To execute the commit directly, use:
```bash
goreview commit --execute
```

Optional flags:
- `--type feat` - Force commit type
- `--scope api` - Force scope
- `--breaking` - Mark as breaking change
