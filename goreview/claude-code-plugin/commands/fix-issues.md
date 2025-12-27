---
description: Auto-fix code issues found by review
---

Automatically fix issues found in code review:

```bash
goreview fix --staged
```

To preview what would be fixed without applying:
```bash
goreview fix --staged --dry-run
```

Filter by severity or type:
```bash
goreview fix --staged --severity critical,error
goreview fix --staged --type security,bug
```
