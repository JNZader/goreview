---
description: Generate changelog from git commits
---

Generate a changelog based on git commits:

```bash
goreview changelog
```

Options:
- `--from v1.0.0` - Start from specific tag/commit
- `--to HEAD` - End at specific point
- `-o CHANGELOG.md` - Write to file

For JSON output:
```bash
goreview changelog --format json
```
