#!/bin/bash
# Quick review script for GoReview plugin
# Usage: ./quick-review.sh [mode]

MODE="${1:-standard}"

echo "🔍 Running GoReview quick scan..."

# Run review with minimal preset for speed
goreview review --staged --preset=minimal --format markdown

# Exit with review result
exit $?
