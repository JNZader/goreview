/**
 * AI Provider interface and types.
 * Ported from goreview/internal/providers/types.go
 */
export declare const COMMIT_MESSAGE_PROMPT = "Generate a conventional commit message for this diff.\nFormat: <type>(<scope>): <description>\nTypes: feat, fix, docs, style, refactor, perf, test, chore\n\nDiff:\n%DIFF%\n\nReturn ONLY the commit message, nothing else.";
/** Build commit message prompt with diff */
export declare function buildCommitPrompt(diff: string): string;
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
export declare function buildReviewPrompt(request: ReviewRequest): string;
/**
 * Parse AI response to ReviewResponse.
 */
export declare function parseReviewResponse(response: string): ReviewResponse;
//# sourceMappingURL=types.d.ts.map