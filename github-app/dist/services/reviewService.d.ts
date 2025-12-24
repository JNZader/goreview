import { Octokit } from '@octokit/rest';
import { ReviewIssue } from './ai/index.js';
export interface PRReviewResult {
    filesReviewed: number;
    totalIssues: number;
    criticalIssues: number;
    files: FileReviewResult[];
    summary: string;
    overallScore: number;
}
export interface FileReviewResult {
    path: string;
    issues: ReviewIssue[];
    score: number;
}
/**
 * Service for reviewing pull requests.
 */
export declare class PRReviewService {
    /**
     * Review a pull request.
     */
    reviewPR(octokit: Octokit, owner: string, repo: string, pullNumber: number): Promise<PRReviewResult>;
    private filterFiles;
    private matchPattern;
    private reviewFiles;
    private generateSummary;
}
export declare const prReviewService: PRReviewService;
//# sourceMappingURL=reviewService.d.ts.map