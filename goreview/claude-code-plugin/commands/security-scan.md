---
description: Deep security scan with root cause tracing
---

Run a comprehensive security scan with root cause analysis:

```bash
goreview review --staged --mode=security --personality=security-expert --trace
```

This will:
1. Analyze code for OWASP Top 10 vulnerabilities
2. Check for hardcoded secrets
3. Identify injection risks (SQL, XSS, command)
4. Trace each issue to its root cause
5. Provide actionable fix suggestions
