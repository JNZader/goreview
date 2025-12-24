/**
 * AI Provider interface and types.
 * Ported from goreview/internal/providers/types.go
 */
// Shared prompt templates (avoid duplication across providers)
export const COMMIT_MESSAGE_PROMPT = `Generate a conventional commit message for this diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, chore

Diff:
%DIFF%

Return ONLY the commit message, nothing else.`;
/** Build commit message prompt with diff */
export function buildCommitPrompt(diff) {
    return COMMIT_MESSAGE_PROMPT.replace('%DIFF%', diff);
}
/**
 * Build the review prompt for AI models.
 */
export function buildReviewPrompt(request) {
    return `You are an expert code reviewer. Analyze this code and identify issues.

File: ${request.filePath}
Language: ${request.language}
${request.context ? `Context: ${request.context}` : ''}

Code changes:
${request.diff}

Return a JSON object with this structure:
{
  "issues": [
    {
      "id": "1",
      "type": "bug|security|performance|style|best_practice",
      "severity": "info|warning|error|critical",
      "message": "description of the issue",
      "suggestion": "how to fix it",
      "line": 10
    }
  ],
  "summary": "brief summary of the review",
  "score": 85
}

Only report real issues, not style nitpicks. Focus on:
- Security vulnerabilities
- Bugs and logic errors
- Performance problems
- Best practices violations

Return valid JSON only.`;
}
/**
 * Parse AI response to ReviewResponse.
 */
export function parseReviewResponse(response) {
    try {
        const parsed = JSON.parse(response);
        return {
            issues: parsed.issues || [],
            summary: parsed.summary || 'No summary provided',
            score: typeof parsed.score === 'number' ? parsed.score : 70,
        };
    }
    catch {
        // Fallback response
        return {
            issues: [],
            summary: response.slice(0, 200),
            score: 70,
        };
    }
}
//# sourceMappingURL=types.js.map