#!/bin/bash
# Pre-commit check script for GoReview plugin
# Usage: ./pre-commit-check.sh
# Returns non-zero if critical issues found

echo "🔍 GoReview pre-commit check..."

# Run review and capture JSON output
RESULT=$(goreview review --staged --preset=standard --format json 2>/dev/null)

# Count critical and error issues
CRITICAL=$(echo "$RESULT" | python3 -c "
import json, sys
try:
    data = json.load(sys.stdin)
    count = sum(
        1 for f in data.get('files', [])
        for i in f.get('response', {}).get('issues', [])
        if i.get('severity') in ['critical', 'error']
    )
    print(count)
except:
    print(0)
" 2>/dev/null || echo "0")

if [ "$CRITICAL" -gt 0 ]; then
    echo "❌ Found $CRITICAL critical/error issues. Please fix before committing."
    echo ""
    goreview review --staged --preset=standard --format markdown
    exit 1
else
    echo "✅ No critical issues found."
    exit 0
fi
