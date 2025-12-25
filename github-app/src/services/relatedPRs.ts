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
export class RelatedPRsService {
  private octokit: Octokit;
  private owner: string;
  private repo: string;

  constructor(octokit: Octokit, owner: string, repo: string) {
    this.octokit = octokit;
    this.owner = owner;
    this.repo = repo;
  }

  /**
   * Find related PRs and issues for a given PR
   */
  async findRelated(prNumber: number): Promise<RelatedContext> {
    // Get files changed in this PR
    const changedFiles = await this.getChangedFiles(prNumber);

    if (changedFiles.length === 0) {
      return { relatedPRs: [], relatedIssues: [], suggestedReviewers: [] };
    }

    // Find related PRs in parallel
    const [relatedPRs, relatedIssues] = await Promise.all([
      this.findRelatedPRs(prNumber, changedFiles),
      this.findRelatedIssues(changedFiles),
    ]);

    // Extract suggested reviewers from related merged PRs
    const suggestedReviewers = this.extractSuggestedReviewers(relatedPRs);

    return { relatedPRs, relatedIssues, suggestedReviewers };
  }

  /**
   * Get files changed in a PR
   */
  private async getChangedFiles(prNumber: number): Promise<string[]> {
    try {
      const { data: files } = await this.octokit.pulls.listFiles({
        owner: this.owner,
        repo: this.repo,
        pull_number: prNumber,
        per_page: 100,
      });

      return files.map(f => f.filename);
    } catch (error) {
      console.error('Error getting PR files:', error);
      return [];
    }
  }

  /**
   * Find PRs that touched the same files
   */
  private async findRelatedPRs(
    currentPR: number,
    changedFiles: string[]
  ): Promise<RelatedPR[]> {
    const relatedPRs: RelatedPR[] = [];
    const seenPRs = new Set<number>([currentPR]);

    // Get recently closed/merged PRs
    try {
      const { data: recentPRs } = await this.octokit.pulls.list({
        owner: this.owner,
        repo: this.repo,
        state: 'all',
        sort: 'updated',
        direction: 'desc',
        per_page: 50,
      });

      for (const pr of recentPRs) {
        if (seenPRs.has(pr.number)) continue;
        seenPRs.add(pr.number);

        // Get files for this PR
        const prFiles = await this.getChangedFiles(pr.number);
        const overlap = changedFiles.filter(f => prFiles.includes(f));

        if (overlap.length > 0) {
          const overlapPercentage = (overlap.length / changedFiles.length) * 100;

          relatedPRs.push({
            number: pr.number,
            title: pr.title,
            url: pr.html_url,
            author: pr.user?.login || 'unknown',
            state: pr.merged_at ? 'merged' : pr.state as 'open' | 'closed',
            mergedAt: pr.merged_at || undefined,
            filesOverlap: overlap,
            overlapPercentage,
          });
        }
      }
    } catch (error) {
      console.error('Error finding related PRs:', error);
    }

    // Sort by overlap percentage
    return relatedPRs
      .sort((a, b) => b.overlapPercentage - a.overlapPercentage)
      .slice(0, 10);
  }

  /**
   * Find open issues that mention the changed files
   */
  private async findRelatedIssues(changedFiles: string[]): Promise<RelatedIssue[]> {
    const relatedIssues: RelatedIssue[] = [];

    try {
      // Search for open issues
      const { data: issues } = await this.octokit.issues.listForRepo({
        owner: this.owner,
        repo: this.repo,
        state: 'open',
        per_page: 100,
      });

      for (const issue of issues) {
        // Skip PRs (they're also returned as issues)
        if (issue.pull_request) continue;

        // Check if issue mentions any of the changed files
        const issueText = `${issue.title} ${issue.body || ''}`.toLowerCase();
        const mentionedFiles = changedFiles.filter(file => {
          const fileName = file.split('/').pop()?.toLowerCase() || '';
          const dirName = file.split('/').slice(-2, -1)[0]?.toLowerCase() || '';
          return issueText.includes(fileName) || issueText.includes(dirName);
        });

        if (mentionedFiles.length > 0) {
          relatedIssues.push({
            number: issue.number,
            title: issue.title,
            url: issue.html_url,
            state: issue.state as 'open' | 'closed',
            labels: issue.labels.map(l => (typeof l === 'string' ? l : l.name || '')),
            mentionedFiles,
          });
        }
      }
    } catch (error) {
      console.error('Error finding related issues:', error);
    }

    return relatedIssues.slice(0, 10);
  }

  /**
   * Extract suggested reviewers from related merged PRs
   */
  private extractSuggestedReviewers(relatedPRs: RelatedPR[]): string[] {
    const reviewerCounts = new Map<string, number>();

    for (const pr of relatedPRs) {
      if (pr.state === 'merged' && pr.author) {
        const count = reviewerCounts.get(pr.author) || 0;
        reviewerCounts.set(pr.author, count + 1);
      }
    }

    return Array.from(reviewerCounts.entries())
      .sort((a, b) => b[1] - a[1])
      .slice(0, 3)
      .map(([author]) => author);
  }

  /**
   * Generate markdown section for related PRs/issues
   */
  generateMarkdownSection(context: RelatedContext): string {
    const sections: string[] = [];

    if (context.relatedPRs.length > 0) {
      sections.push('### Related PRs\n');
      sections.push('PRs that modified the same files:\n');

      for (const pr of context.relatedPRs.slice(0, 5)) {
        const stateEmoji = pr.state === 'merged' ? '🟣' : pr.state === 'open' ? '🟢' : '🔴';
        const overlap = `${pr.overlapPercentage.toFixed(0)}% overlap`;
        sections.push(`- ${stateEmoji} #${pr.number}: ${pr.title} (${overlap})`);
      }
      sections.push('');
    }

    if (context.relatedIssues.length > 0) {
      sections.push('### Related Issues\n');
      sections.push('Open issues that may be addressed by this PR:\n');

      for (const issue of context.relatedIssues.slice(0, 5)) {
        const labels = issue.labels.length > 0
          ? ` [${issue.labels.slice(0, 3).join(', ')}]`
          : '';
        sections.push(`- #${issue.number}: ${issue.title}${labels}`);
      }
      sections.push('');
    }

    if (context.suggestedReviewers.length > 0) {
      sections.push('### Suggested Reviewers\n');
      sections.push('Based on previous contributions to these files:\n');
      sections.push(`- ${context.suggestedReviewers.map(r => `@${r}`).join(', ')}`);
      sections.push('');
    }

    return sections.join('\n');
  }
}

/**
 * Create a RelatedPRsService instance
 */
export function createRelatedPRsService(
  octokit: Octokit,
  owner: string,
  repo: string
): RelatedPRsService {
  return new RelatedPRsService(octokit, owner, repo);
}
