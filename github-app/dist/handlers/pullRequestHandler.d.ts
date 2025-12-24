/**
 * Handle pull request events.
 */
export declare function handlePullRequest(action: string | undefined, payload: unknown): Promise<void>;
/**
 * Process a queued PR review job.
 */
export declare function processReviewJob(installationId: number, owner: string, repo: string, pullNumber: number, headSha: string): Promise<void>;
//# sourceMappingURL=pullRequestHandler.d.ts.map