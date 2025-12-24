/**
 * AI Provider interface and types.
 * Ported from goreview/internal/providers/types.go
 */

export interface ReviewRequest {
  diff: string;
  language: string;
  filePath: string;
  context?: string;
}

export interface ReviewIssue {
  id: string;
  type: 'bug' | 'security' | 'performance' | 'style' | 'best_practice';
  severity: 'info' | 'warning' | 'error' | 'critical';
  message: string;
  suggestion?: string;
  line?: number;
  endLine?: number;
}

export interface ReviewResponse {
  issues: ReviewIssue[];
  summary: string;
  score: number;
  tokensUsed?: number;
  processingTimeMs?: number;
}

/**
 * AI Provider interface - all providers must implement this.
 */
export interface AIProvider {
  /** Provider name */
  readonly name: string;

  /** Review code and return issues */
  review(request: ReviewRequest): Promise<ReviewResponse>;

  /** Generate a commit message from diff */
  generateCommitMessage(diff: string): Promise<string>;

  /** General chat/conversation for interactive responses */
  chat(prompt: string): Promise<string>;

  /** Check if provider is available */
  healthCheck(): Promise<boolean>;
}

/**
 * Build the review prompt for AI models.
 */
export function buildReviewPrompt(request: ReviewRequest): string {
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
export function parseReviewResponse(response: string): ReviewResponse {
  try {
    const parsed = JSON.parse(response);
    return {
      issues: parsed.issues || [],
      summary: parsed.summary || 'No summary provided',
      score: typeof parsed.score === 'number' ? parsed.score : 70,
    };
  } catch {
    // Fallback response
    return {
      issues: [],
      summary: response.slice(0, 200),
      score: 70,
    };
  }
}
