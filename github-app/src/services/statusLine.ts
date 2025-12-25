import { Octokit } from '@octokit/rest';
import { logger } from '../utils/logger.js';

/**
 * StatusLine service for generating PR review status summaries.
 * Provides visual indicators of review progress and warnings.
 */

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
}

export interface StatusLineOptions {
  /** Warning threshold for hours since last update */
  inactivityWarningHours?: number;
  /** Critical threshold for hours since last update */
  inactivityCriticalHours?: number;
  /** Show emoji indicators */
  showEmoji?: boolean;
}

const DEFAULT_OPTIONS: StatusLineOptions = {
  inactivityWarningHours: 48,
  inactivityCriticalHours: 72,
  showEmoji: true,
};

/**
 * Generate a status line summary for a PR review.
 */
export function generateStatusLine(
  status: ReviewStatus,
  options: StatusLineOptions = {}
): string {
  const opts = { ...DEFAULT_OPTIONS, ...options };
  const parts: string[] = [];

  // Score indicator
  const scoreEmoji = getScoreEmoji(status.score, opts.showEmoji);
  parts.push(`${scoreEmoji} Score: ${status.score}/100`);

  // Issues summary
  const issuesEmoji = status.criticalIssues > 0 ? '🔴' : status.totalIssues > 0 ? '🟡' : '🟢';
  if (opts.showEmoji) {
    parts.push(`${issuesEmoji} Issues: ${status.totalIssues}`);
  } else {
    parts.push(`Issues: ${status.totalIssues}`);
  }

  // Critical issues
  if (status.criticalIssues > 0) {
    parts.push(`Critical: ${status.criticalIssues}`);
  }

  // Progress (if there's history)
  if (status.reviewRound > 1) {
    const resolved = status.resolvedIssues;
    const persistent = status.persistentIssues;
    const total = resolved + persistent;

    if (total > 0) {
      const percentage = Math.round((resolved / total) * 100);
      const progressEmoji = opts.showEmoji ? getProgressEmoji(percentage) : '';
      parts.push(`${progressEmoji} Progress: ${resolved}/${total} resolved (${percentage}%)`.trim());
    }
  }

  // Review round
  parts.push(`Round: ${status.reviewRound}`);

  return parts.join(' | ');
}

/**
 * Generate a detailed status block for PR comments.
 */
export function generateStatusBlock(
  status: ReviewStatus,
  options: StatusLineOptions = {}
): string {
  const opts = { ...DEFAULT_OPTIONS, ...options };
  const lines: string[] = [];

  // Header
  lines.push('## GoReview Status');
  lines.push('');

  // Main status line
  lines.push(`> ${generateStatusLine(status, opts)}`);
  lines.push('');

  // Inactivity warning
  const inactivityWarning = checkInactivity(status.lastReviewAt, opts);
  if (inactivityWarning) {
    lines.push(`⚠️ **Warning:** ${inactivityWarning}`);
    lines.push('');
  }

  // Progress section (if there's history)
  if (status.reviewRound > 1) {
    lines.push('### Progress Since Last Review');
    lines.push('');

    if (status.resolvedIssues > 0) {
      lines.push(`✅ ${status.resolvedIssues} issue(s) resolved`);
    }
    if (status.persistentIssues > 0) {
      lines.push(`⏳ ${status.persistentIssues} issue(s) still pending`);
    }
    if (status.totalIssues > status.persistentIssues) {
      const newIssues = status.totalIssues - status.persistentIssues;
      if (newIssues > 0) {
        lines.push(`🆕 ${newIssues} new issue(s) found`);
      }
    }
    lines.push('');
  }

  // Critical issues warning
  if (status.criticalIssues > 0) {
    lines.push(`🚨 **${status.criticalIssues} critical issue(s) require immediate attention**`);
    lines.push('');
  }

  return lines.join('\n');
}

/**
 * Parse existing review comment to extract previous status.
 */
export function parseExistingStatus(commentBody: string): Partial<ReviewStatus> | null {
  try {
    // Look for status data in HTML comment
    const match = commentBody.match(/<!--goreview-status:(.*?)-->/);
    if (match) {
      return JSON.parse(match[1]);
    }
    return null;
  } catch (error) {
    logger.debug({ error }, 'Failed to parse existing status');
    return null;
  }
}

/**
 * Embed status data in comment for future parsing.
 */
export function embedStatusData(status: ReviewStatus): string {
  return `<!--goreview-status:${JSON.stringify(status)}-->`;
}

/**
 * Fetch previous review status from PR comments.
 */
export async function getPreviousStatus(
  octokit: Octokit,
  owner: string,
  repo: string,
  pullNumber: number
): Promise<Partial<ReviewStatus> | null> {
  try {
    const { data: comments } = await octokit.issues.listComments({
      owner,
      repo,
      issue_number: pullNumber,
      per_page: 100,
    });

    // Find the most recent GoReview comment
    for (let i = comments.length - 1; i >= 0; i--) {
      const comment = comments[i];
      if (comment.body?.includes('<!--goreview-status:')) {
        const status = parseExistingStatus(comment.body);
        if (status) {
          return status;
        }
      }
    }

    return null;
  } catch (error) {
    logger.error({ error, owner, repo, pullNumber }, 'Failed to fetch previous status');
    return null;
  }
}

/**
 * Calculate the review round based on previous status.
 */
export function calculateReviewRound(previousStatus: Partial<ReviewStatus> | null): number {
  if (!previousStatus?.reviewRound) {
    return 1;
  }
  return previousStatus.reviewRound + 1;
}

/**
 * Compare current issues with previous to find resolved/persistent.
 */
export function compareIssues(
  currentIssueIds: string[],
  previousIssueIds: string[]
): { resolved: number; persistent: number } {
  const currentSet = new Set(currentIssueIds);
  const previousSet = new Set(previousIssueIds);

  let resolved = 0;
  let persistent = 0;

  for (const id of previousSet) {
    if (currentSet.has(id)) {
      persistent++;
    } else {
      resolved++;
    }
  }

  return { resolved, persistent };
}

// Helper functions

function getScoreEmoji(score: number, showEmoji: boolean | undefined): string {
  if (!showEmoji) return '';
  if (score >= 90) return '🏆';
  if (score >= 80) return '✅';
  if (score >= 60) return '⚠️';
  return '🔴';
}

function getProgressEmoji(percentage: number): string {
  if (percentage >= 80) return '🎯';
  if (percentage >= 50) return '📈';
  if (percentage >= 20) return '🔄';
  return '⏳';
}

function checkInactivity(
  lastReviewAt: string,
  options: StatusLineOptions
): string | null {
  const lastReview = new Date(lastReviewAt);
  const now = new Date();
  const hoursSinceReview = (now.getTime() - lastReview.getTime()) / (1000 * 60 * 60);

  if (hoursSinceReview >= (options.inactivityCriticalHours ?? 72)) {
    return `This PR has been inactive for ${Math.floor(hoursSinceReview)} hours. Please address pending issues or close the PR.`;
  }

  if (hoursSinceReview >= (options.inactivityWarningHours ?? 48)) {
    return `This PR has pending issues for ${Math.floor(hoursSinceReview)} hours.`;
  }

  return null;
}

export const statusLineService = {
  generateStatusLine,
  generateStatusBlock,
  parseExistingStatus,
  embedStatusData,
  getPreviousStatus,
  calculateReviewRound,
  compareIssues,
};
