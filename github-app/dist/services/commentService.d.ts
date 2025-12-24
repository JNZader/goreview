import { Octokit } from '@octokit/rest';
import { PRReviewResult } from './reviewService.js';
export interface CommentOptions {
    inline: boolean;
    summary: boolean;
    requestChanges: boolean;
    minSeverity: 'info' | 'warning' | 'error' | 'critical';
}
/**
 * Service for posting review comments to GitHub.
 */
export declare class CommentService {
    /**
     * Post review results to a pull request.
     */
    postReview(octokit: Octokit, owner: string, repo: string, pullNumber: number, result: PRReviewResult): Promise<void>;
    private buildReviewBody;
    private buildInlineComments;
    private formatIssueComment;
    private getSeverityEmoji;
    private addLabel;
}
export declare const commentService: CommentService;
//# sourceMappingURL=commentService.d.ts.map