#!/bin/bash
# Deep security scan script for GoReview plugin
# Usage: ./security-scan.sh [files...]

echo "🔒 Running GoReview security scan..."

# Run comprehensive security review with root cause tracing
if [ $# -eq 0 ]; then
    goreview review --staged --mode=security --personality=security-expert --trace --format markdown
else
    goreview review --mode=security --personality=security-expert --trace --format markdown "$@"
fi

exit $?
