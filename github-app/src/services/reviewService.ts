import { Octokit } from '@octokit/rest';
import { getProvider, ReviewIssue } from './ai/index.js';
import { loadRepoConfig, RepoConfig } from '../config/repoConfig.js';
import { config } from '../config/index.js';
import { logger } from '../utils/logger.js';
import { parseDiff, DiffFile } from '../utils/diffParser.js';

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
export class PRReviewService {
  /**
   * Review a pull request.
   */
  async reviewPR(
    octokit: Octokit,
    owner: string,
    repo: string,
    pullNumber: number
  ): Promise<PRReviewResult> {
    const startTime = Date.now();

    // Load repository configuration
    const repoConfig = await loadRepoConfig(octokit, owner, repo);

    if (!repoConfig.review.enabled) {
      logger.info({ owner, repo, pullNumber }, 'Reviews disabled for repository');
      throw new Error('Reviews are disabled for this repository');
    }

    // Get pull request diff
    const { data: diffData } = await octokit.pulls.get({
      owner,
      repo,
      pull_number: pullNumber,
      mediaType: { format: 'diff' },
    });

    const diff = diffData as unknown as string;

    // Parse diff into files
    const files = parseDiff(diff);

    // Filter files based on configuration
    const filesToReview = this.filterFiles(files, repoConfig);

    if (filesToReview.length === 0) {
      return {
        filesReviewed: 0,
        totalIssues: 0,
        criticalIssues: 0,
        files: [],
        summary: 'No files to review after applying filters.',
        overallScore: 100,
      };
    }

    // Check limits
    if (filesToReview.length > config.review.maxFiles) {
      logger.warn({
        owner, repo, pullNumber,
        fileCount: filesToReview.length,
        maxFiles: config.review.maxFiles,
      }, 'Too many files to review');

      throw new Error(`Too many files to review: ${filesToReview.length} > ${config.review.maxFiles}`);
    }

    // Review files concurrently with limit
    const results = await this.reviewFiles(filesToReview);

    // Aggregate results
    const totalIssues = results.reduce((sum, r) => sum + r.issues.length, 0);
    const criticalIssues = results.reduce(
      (sum, r) => sum + r.issues.filter(i => i.severity === 'critical').length,
      0
    );
    const avgScore = results.length > 0
      ? results.reduce((sum, r) => sum + r.score, 0) / results.length
      : 100;

    const duration = Date.now() - startTime;
    logger.info({
      owner, repo, pullNumber,
      filesReviewed: results.length,
      totalIssues,
      duration,
    }, 'PR review completed');

    return {
      filesReviewed: results.length,
      totalIssues,
      criticalIssues,
      files: results,
      summary: this.generateSummary(results),
      overallScore: Math.round(avgScore),
    };
  }

  private filterFiles(files: DiffFile[], repoConfig: RepoConfig): DiffFile[] {
    return files.filter(file => {
      // Skip deleted files
      if (file.status === 'deleted') return false;

      // Skip binary files
      if (file.isBinary) return false;

      // Check ignore patterns
      for (const pattern of repoConfig.review.ignore_patterns) {
        if (this.matchPattern(pattern, file.path)) {
          return false;
        }
      }

      // Check language filter
      if (repoConfig.review.languages && repoConfig.review.languages.length > 0) {
        if (!repoConfig.review.languages.includes(file.language)) {
          return false;
        }
      }

      return true;
    });
  }

  private matchPattern(pattern: string, path: string): boolean {
    // Simple glob matching
    if (pattern.endsWith('**')) {
      return path.startsWith(pattern.slice(0, -2));
    }
    if (pattern.endsWith('*')) {
      return path.startsWith(pattern.slice(0, -1));
    }
    return pattern === path;
  }

  private async reviewFiles(files: DiffFile[]): Promise<FileReviewResult[]> {
    const results: FileReviewResult[] = [];
    const concurrency = 3; // Limit concurrent reviews

    for (let i = 0; i < files.length; i += concurrency) {
      const batch = files.slice(i, i + concurrency);

      const batchResults = await Promise.all(
        batch.map(async (file) => {
          try {
            const response = await getProvider().review({
              diff: file.content,
              language: file.language,
              filePath: file.path,
            });

            return {
              path: file.path,
              issues: response.issues,
              score: response.score,
            };
          } catch (error) {
            logger.error({ error, file: file.path }, 'Failed to review file');
            return {
              path: file.path,
              issues: [],
              score: 70,
            };
          }
        })
      );

      results.push(...batchResults);
    }

    return results;
  }

  private generateSummary(results: FileReviewResult[]): string {
    const totalIssues = results.reduce((sum, r) => sum + r.issues.length, 0);
    const criticalCount = results.reduce(
      (sum, r) => sum + r.issues.filter(i => i.severity === 'critical').length,
      0
    );
    const errorCount = results.reduce(
      (sum, r) => sum + r.issues.filter(i => i.severity === 'error').length,
      0
    );

    if (totalIssues === 0) {
      return 'No issues found. The code looks good!';
    }

    const parts = [`Found ${totalIssues} issue(s) across ${results.length} file(s).`];

    if (criticalCount > 0) {
      parts.push(`**${criticalCount} critical issue(s)** require immediate attention.`);
    }
    if (errorCount > 0) {
      parts.push(`${errorCount} error(s) should be fixed before merging.`);
    }

    return parts.join(' ');
  }
}

export const prReviewService = new PRReviewService();
