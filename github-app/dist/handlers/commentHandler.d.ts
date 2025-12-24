/**
 * Handler for interactive comments.
 * Responds to @goreview mentions in issue/PR comments.
 */
import type { IssueCommentEvent } from '@octokit/webhooks-types';
/**
 * Handle issue comment events
 */
export declare function handleIssueComment(event: IssueCommentEvent): Promise<void>;
/**
 * Handle pull request review comment events
 * Allows responding to @mentions in review comments (on specific lines)
 */
export declare function handlePullRequestReviewComment(event: {
    action: string;
    comment: {
        id: number;
        body: string;
        user: {
            login: string;
            type: string;
        };
        diff_hunk?: string;
        path?: string;
        line?: number;
    };
    pull_request: {
        number: number;
    };
    repository: {
        full_name: string;
        owner: {
            login: string;
        };
        name: string;
    };
    installation?: {
        id: number;
    };
}): Promise<void>;
//# sourceMappingURL=commentHandler.d.ts.map