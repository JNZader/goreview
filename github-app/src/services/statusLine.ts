import { Octokit } from '@octokit/rest';
import { logger } from '../utils/logger.js';

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
  const issuesEmoji = getIssuesEmoji(status.criticalIssues, status.totalIssues);
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
    const statusRegex = /<!--goreview-status:(.*?)-->/;
    const match = statusRegex.exec(commentBody);
    if (match?.[1]) {
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
      const body = comment?.body;
      if (body?.includes('<!--goreview-status:')) {
        const status = parseExistingStatus(body);
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

function getIssuesEmoji(criticalIssues: number, totalIssues: number): string {
  if (criticalIssues > 0) return '🔴';
  if (totalIssues > 0) return '🟡';
  return '🟢';
}

function getSeverityEmoji(severity: string): string {
  switch (severity) {
    case 'critical': return '🚨';
    case 'error': return '❌';
    case 'warning': return '⚠️';
    default: return 'ℹ️';
  }
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

/**
 * Generate a unique ID for an issue based on location and content.
 */
export function generateIssueId(
  file: string,
  line: number | undefined,
  message: string
): string {
  const content = `${file}:${line ?? 0}:${message.slice(0, 100)}`;
  // Simple hash function for consistency
  let hash = 0;
  for (let i = 0; i < content.length; i++) {
    const char = content.codePointAt(i) ?? 0;
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // Convert to 32bit integer
  }
  return `issue-${Math.abs(hash).toString(16)}`;
}

/**
 * Generate detailed persistent issues section for PR comment.
 */
export function generatePersistentIssuesSection(
  currentIssues: ReviewIssueRecord[],
  previousIssues: ReviewIssueRecord[],
  currentRound: number
): string {
  const previousIds = new Set(previousIssues.map(i => i.id));
  const persistent = currentIssues.filter(i => previousIds.has(i.id));

  if (persistent.length === 0) {
    return '';
  }

  const lines: string[] = [];
  lines.push('### Persistent Issues');
  lines.push('');
  lines.push('These issues were found in previous reviews and still need attention:');
  lines.push('');

  // Group by severity
  const bySeverity: Record<string, ReviewIssueRecord[]> = {};
  for (const issue of persistent) {
    const sev = issue.severity;
    if (!bySeverity[sev]) {
      bySeverity[sev] = [];
    }
    const arr = bySeverity[sev];
    if (arr) arr.push(issue);
  }

  const severityOrder = ['critical', 'error', 'warning', 'info'];

  for (const severity of severityOrder) {
    const issues = bySeverity[severity];
    if (!issues || issues.length === 0) continue;

    for (const issue of issues) {
      const roundsOld = currentRound - issue.firstSeenRound;
      const roundLabel = roundsOld > 0 ? ` (since Round ${issue.firstSeenRound})` : ' (new)';
      const emoji = getSeverityEmoji(severity);

      const location = issue.line ? `${issue.file}:${issue.line}` : issue.file;
      lines.push(`${emoji} **[${severity.toUpperCase()}]** \`${location}\`${roundLabel}`);
      lines.push(`   ${issue.message}`);
      lines.push('');
    }
  }

  return lines.join('\n');
}

/**
 * Determine which files have been "approved" (no issues for N rounds).
 */
export function getApprovedFiles(
  currentFiles: string[],
  issueRecords: ReviewIssueRecord[],
  previousApproved: string[] = [],
  requiredCleanRounds: number = 1
): string[] {
  // Files with current issues
  const filesWithIssues = new Set(issueRecords.map(i => i.file));

  // Files that are clean in this round
  const cleanFiles = currentFiles.filter(f => !filesWithIssues.has(f));

  // Combine with previously approved files (that still exist)
  const allApproved = new Set([
    ...previousApproved.filter(f => currentFiles.includes(f)),
    ...cleanFiles,
  ]);

  return Array.from(allApproved);
}

/**
 * Build a complete handoff status for the next review round.
 */
export function buildHandoffStatus(
  currentResult: {
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
  },
  previousStatus: Partial<ReviewStatus> | null
): ReviewStatus {
  const currentRound = calculateReviewRound(previousStatus);
  const now = new Date().toISOString();

  // Build issue records with IDs
  const issueRecords: ReviewIssueRecord[] = [];
  for (const file of currentResult.files) {
    for (const issue of file.issues) {
      const id = generateIssueId(file.path, issue.line, issue.message);

      // Check if this issue existed before
      const previousIssue = previousStatus?.issueRecords?.find(p => p.id === id);
      const firstSeenRound = previousIssue?.firstSeenRound ?? currentRound;

      issueRecords.push({
        id,
        file: file.path,
        line: issue.line,
        severity: issue.severity as ReviewIssueRecord['severity'],
        type: issue.type,
        message: issue.message,
        firstSeenRound,
      });
    }
  }

  // Calculate resolved/persistent
  const previousIds = new Set(previousStatus?.issueRecords?.map(i => i.id) ?? []);
  const currentIds = new Set(issueRecords.map(i => i.id));

  let resolved = 0;
  let persistent = 0;
  for (const id of previousIds) {
    if (currentIds.has(id)) {
      persistent++;
    } else {
      resolved++;
    }
  }

  // Get approved files
  const allFiles = currentResult.files.map(f => f.path);
  const approvedFiles = getApprovedFiles(
    allFiles,
    issueRecords,
    previousStatus?.approvedFiles
  );

  return {
    totalIssues: currentResult.totalIssues,
    criticalIssues: currentResult.criticalIssues,
    resolvedIssues: resolved,
    persistentIssues: persistent,
    score: currentResult.score,
    lastReviewAt: now,
    reviewRound: currentRound,
    issueRecords,
    approvedFiles,
  };
}

/**
 * Generate complete status block with handoff information.
 */
export function generateHandoffBlock(
  status: ReviewStatus,
  previousStatus: Partial<ReviewStatus> | null,
  options: StatusLineOptions = {}
): string {
  const opts = { ...DEFAULT_OPTIONS, ...options };
  const lines: string[] = [];

  // Status data embed (hidden)
  lines.push(embedStatusData(status));
  lines.push('');

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
  if (status.reviewRound > 1 && previousStatus) {
    lines.push('### Progress Since Last Review');
    lines.push('');

    if (status.resolvedIssues > 0) {
      lines.push(`✅ **${status.resolvedIssues}** issue(s) resolved`);
    }
    if (status.persistentIssues > 0) {
      lines.push(`⏳ **${status.persistentIssues}** issue(s) still pending`);
    }
    const newIssues = status.totalIssues - status.persistentIssues;
    if (newIssues > 0) {
      lines.push(`🆕 **${newIssues}** new issue(s) found`);
    }
    lines.push('');

    // Persistent issues detail
    if (status.issueRecords && previousStatus.issueRecords) {
      const persistentSection = generatePersistentIssuesSection(
        status.issueRecords,
        previousStatus.issueRecords,
        status.reviewRound
      );
      if (persistentSection) {
        lines.push(persistentSection);
      }
    }
  }

  // Critical issues warning
  if (status.criticalIssues > 0) {
    lines.push(`🚨 **${status.criticalIssues} critical issue(s) require immediate attention**`);
    lines.push('');
  }

  // Approved files (if any)
  if (status.approvedFiles && status.approvedFiles.length > 0) {
    lines.push('<details>');
    lines.push(`<summary>✅ ${status.approvedFiles.length} file(s) approved</summary>`);
    lines.push('');
    for (const file of status.approvedFiles.slice(0, 10)) {
      lines.push(`- \`${file}\``);
    }
    if (status.approvedFiles.length > 10) {
      lines.push(`- ... and ${status.approvedFiles.length - 10} more`);
    }
    lines.push('</details>');
    lines.push('');
  }

  return lines.join('\n');
}

export const statusLineService = {
  generateStatusLine,
  generateStatusBlock,
  parseExistingStatus,
  embedStatusData,
  getPreviousStatus,
  calculateReviewRound,
  compareIssues,
  generateIssueId,
  generatePersistentIssuesSection,
  getApprovedFiles,
  buildHandoffStatus,
  generateHandoffBlock,
};
