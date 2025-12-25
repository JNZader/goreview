import { Octokit } from '@octokit/rest';
/**
 * StatusLine service for generating PR review status summaries.
 * Provides visual indicators of review progress and warnings.
 */
export interface ReviewIssueRecord {
    /** Unique identifier for the issue (hash of location + message) */
    id: string;
    /** File path where the issue was found */
    file: string;
    /** Line number in the file */
    line?: number;
    /** Issue severity */
    severity: 'info' | 'warning' | 'error' | 'critical';
    /** Issue type */
    type: string;
    /** Issue description */
    message: string;
    /** Round when first detected */
    firstSeenRound: number;
}
export interface ReviewStatus {
    /** Total issues found in current review */
    totalIssues: number;
    /** Critical issues requiring immediate attention */
    criticalIssues: number;
    /** Issues resolved since last review */
    resolvedIssues: number;
    /** Issues from previous review that persist */
    persistentIssues: number;
    /** Overall review score (0-100) */
    score: number;
    /** Time of last review in ISO format */
    lastReviewAt: string;
    /** Review round number */
    reviewRound: number;
    /** Detailed issue records for tracking */
    issueRecords?: ReviewIssueRecord[];
    /** Files that have been approved (no issues in last N rounds) */
    approvedFiles?: string[];
}
export interface StatusLineOptions {
    /** Warning threshold for hours since last update */
    inactivityWarningHours?: number;
    /** Critical threshold for hours since last update */
    inactivityCriticalHours?: number;
    /** Show emoji indicators */
    showEmoji?: boolean;
}
/**
 * Generate a status line summary for a PR review.
 */
export declare function generateStatusLine(status: ReviewStatus, options?: StatusLineOptions): string;
/**
 * Generate a detailed status block for PR comments.
 */
export declare function generateStatusBlock(status: ReviewStatus, options?: StatusLineOptions): string;
/**
 * Parse existing review comment to extract previous status.
 */
export declare function parseExistingStatus(commentBody: string): Partial<ReviewStatus> | null;
/**
 * Embed status data in comment for future parsing.
 */
export declare function embedStatusData(status: ReviewStatus): string;
/**
 * Fetch previous review status from PR comments.
 */
export declare function getPreviousStatus(octokit: Octokit, owner: string, repo: string, pullNumber: number): Promise<Partial<ReviewStatus> | null>;
/**
 * Calculate the review round based on previous status.
 */
export declare function calculateReviewRound(previousStatus: Partial<ReviewStatus> | null): number;
/**
 * Compare current issues with previous to find resolved/persistent.
 */
export declare function compareIssues(currentIssueIds: string[], previousIssueIds: string[]): {
    resolved: number;
    persistent: number;
};
/**
 * Generate a unique ID for an issue based on location and content.
 */
export declare function generateIssueId(file: string, line: number | undefined, message: string): string;
/**
 * Generate detailed persistent issues section for PR comment.
 */
export declare function generatePersistentIssuesSection(currentIssues: ReviewIssueRecord[], previousIssues: ReviewIssueRecord[], currentRound: number): string;
/**
 * Determine which files have been "approved" (no issues for N rounds).
 */
export declare function getApprovedFiles(currentFiles: string[], issueRecords: ReviewIssueRecord[], previousApproved?: string[], requiredCleanRounds?: number): string[];
/**
 * Build a complete handoff status for the next review round.
 */
export declare function buildHandoffStatus(currentResult: {
    score: number;
    totalIssues: number;
    criticalIssues: number;
    files: Array<{
        path: string;
        issues: Array<{
            severity: string;
            type: string;
            message: string;
            line?: number;
        }>;
    }>;
}, previousStatus: Partial<ReviewStatus> | null): ReviewStatus;
/**
 * Generate complete status block with handoff information.
 */
export declare function generateHandoffBlock(status: ReviewStatus, previousStatus: Partial<ReviewStatus> | null, options?: StatusLineOptions): string;
export declare const statusLineService: {
    generateStatusLine: typeof generateStatusLine;
    generateStatusBlock: typeof generateStatusBlock;
    parseExistingStatus: typeof parseExistingStatus;
    embedStatusData: typeof embedStatusData;
    getPreviousStatus: typeof getPreviousStatus;
    calculateReviewRound: typeof calculateReviewRound;
    compareIssues: typeof compareIssues;
    generateIssueId: typeof generateIssueId;
    generatePersistentIssuesSection: typeof generatePersistentIssuesSection;
    getApprovedFiles: typeof getApprovedFiles;
    buildHandoffStatus: typeof buildHandoffStatus;
    generateHandoffBlock: typeof generateHandoffBlock;
};
//# sourceMappingURL=statusLine.d.ts.map