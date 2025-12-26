import { Octokit } from '@octokit/rest';
/**
 * RelatedPR represents a PR that touched the same files
 */
export interface RelatedPR {
    number: number;
    title: string;
    url: string;
    author: string;
    state: 'open' | 'closed' | 'merged';
    mergedAt?: string;
    filesOverlap: string[];
    overlapPercentage: number;
}
/**
 * RelatedIssue represents an issue related to the files being changed
 */
export interface RelatedIssue {
    number: number;
    title: string;
    url: string;
    state: 'open' | 'closed';
    labels: string[];
    mentionedFiles: string[];
}
/**
 * RelatedContext contains all related PRs and issues
 */
export interface RelatedContext {
    relatedPRs: RelatedPR[];
    relatedIssues: RelatedIssue[];
    suggestedReviewers: string[];
}
/**
 * RelatedPRsService finds PRs and issues related to the current PR
 */
export declare class RelatedPRsService {
    private readonly octokit;
    private readonly owner;
    private readonly repo;
    constructor(octokit: Octokit, owner: string, repo: string);
    /**
     * Find related PRs and issues for a given PR
     */
    findRelated(prNumber: number): Promise<RelatedContext>;
    /**
     * Get files changed in a PR
     */
    private getChangedFiles;
    /**
     * Find PRs that touched the same files
     */
    private findRelatedPRs;
    /**
     * Find open issues that mention the changed files
     */
    private findRelatedIssues;
    /**
     * Extract suggested reviewers from related merged PRs
     */
    private extractSuggestedReviewers;
    /**
     * Generate markdown section for related PRs/issues
     */
    generateMarkdownSection(context: RelatedContext): string;
    private getPRStateEmoji;
}
/**
 * Create a RelatedPRsService instance
 */
export declare function createRelatedPRsService(octokit: Octokit, owner: string, repo: string): RelatedPRsService;
//# sourceMappingURL=relatedPRs.d.ts.map