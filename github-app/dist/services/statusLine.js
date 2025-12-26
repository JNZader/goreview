import { logger } from '../utils/logger.js';
const DEFAULT_OPTIONS = {
    inactivityWarningHours: 48,
    inactivityCriticalHours: 72,
    showEmoji: true,
};
/**
 * Generate a status line summary for a PR review.
 */
export function generateStatusLine(status, options = {}) {
    const opts = { ...DEFAULT_OPTIONS, ...options };
    const parts = [];
    // Score indicator
    const scoreEmoji = getScoreEmoji(status.score, opts.showEmoji);
    parts.push(`${scoreEmoji} Score: ${status.score}/100`);
    // Issues summary
    const issuesEmoji = getIssuesEmoji(status.criticalIssues, status.totalIssues);
    if (opts.showEmoji) {
        parts.push(`${issuesEmoji} Issues: ${status.totalIssues}`);
    }
    else {
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
export function generateStatusBlock(status, options = {}) {
    const opts = { ...DEFAULT_OPTIONS, ...options };
    const lines = [
        // Header
        '## GoReview Status',
        '',
        // Main status line
        `> ${generateStatusLine(status, opts)}`,
        '',
    ];
    // Inactivity warning
    const inactivityWarning = checkInactivity(status.lastReviewAt, opts);
    if (inactivityWarning) {
        lines.push(`⚠️ **Warning:** ${inactivityWarning}`, '');
    }
    // Progress section (if there's history)
    if (status.reviewRound > 1) {
        lines.push('### Progress Since Last Review', '');
        if (status.resolvedIssues > 0) {
            lines.push(`✅ ${status.resolvedIssues} issue(s) resolved`);
        }
        if (status.persistentIssues > 0) {
            lines.push(`⏳ ${status.persistentIssues} issue(s) still pending`);
        }
        const newIssues = status.totalIssues - status.persistentIssues;
        if (newIssues > 0) {
            lines.push(`🆕 ${newIssues} new issue(s) found`);
        }
        lines.push('');
    }
    // Critical issues warning
    if (status.criticalIssues > 0) {
        lines.push(`🚨 **${status.criticalIssues} critical issue(s) require immediate attention**`, '');
    }
    return lines.join('\n');
}
/**
 * Parse existing review comment to extract previous status.
 */
export function parseExistingStatus(commentBody) {
    try {
        // Look for status data in HTML comment
        const statusRegex = /<!--goreview-status:(.*?)-->/;
        const match = statusRegex.exec(commentBody);
        if (match?.[1]) {
            return JSON.parse(match[1]);
        }
        return null;
    }
    catch (error) {
        logger.debug({ error }, 'Failed to parse existing status');
        return null;
    }
}
/**
 * Embed status data in comment for future parsing.
 */
export function embedStatusData(status) {
    return `<!--goreview-status:${JSON.stringify(status)}-->`;
}
/**
 * Fetch previous review status from PR comments.
 */
export async function getPreviousStatus(octokit, owner, repo, pullNumber) {
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
    }
    catch (error) {
        logger.error({ error, owner, repo, pullNumber }, 'Failed to fetch previous status');
        return null;
    }
}
/**
 * Calculate the review round based on previous status.
 */
export function calculateReviewRound(previousStatus) {
    if (!previousStatus?.reviewRound) {
        return 1;
    }
    return previousStatus.reviewRound + 1;
}
/**
 * Compare current issues with previous to find resolved/persistent.
 */
export function compareIssues(currentIssueIds, previousIssueIds) {
    const currentSet = new Set(currentIssueIds);
    const previousSet = new Set(previousIssueIds);
    let resolved = 0;
    let persistent = 0;
    for (const id of previousSet) {
        if (currentSet.has(id)) {
            persistent++;
        }
        else {
            resolved++;
        }
    }
    return { resolved, persistent };
}
// Helper functions
function getScoreEmoji(score, showEmoji) {
    if (!showEmoji)
        return '';
    if (score >= 90)
        return '🏆';
    if (score >= 80)
        return '✅';
    if (score >= 60)
        return '⚠️';
    return '🔴';
}
function getProgressEmoji(percentage) {
    if (percentage >= 80)
        return '🎯';
    if (percentage >= 50)
        return '📈';
    if (percentage >= 20)
        return '🔄';
    return '⏳';
}
function getIssuesEmoji(criticalIssues, totalIssues) {
    if (criticalIssues > 0)
        return '🔴';
    if (totalIssues > 0)
        return '🟡';
    return '🟢';
}
function getSeverityEmoji(severity) {
    switch (severity) {
        case 'critical': return '🚨';
        case 'error': return '❌';
        case 'warning': return '⚠️';
        default: return 'ℹ️';
    }
}
function checkInactivity(lastReviewAt, options) {
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
export function generateIssueId(file, line, message) {
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
export function generatePersistentIssuesSection(currentIssues, previousIssues, currentRound) {
    const previousIds = new Set(previousIssues.map(i => i.id));
    const persistent = currentIssues.filter(i => previousIds.has(i.id));
    if (persistent.length === 0) {
        return '';
    }
    const lines = [
        '### Persistent Issues',
        '',
        'These issues were found in previous reviews and still need attention:',
        '',
    ];
    // Group by severity
    const bySeverity = {};
    for (const issue of persistent) {
        const sev = issue.severity;
        bySeverity[sev] ??= [];
        bySeverity[sev].push(issue);
    }
    const severityOrder = ['critical', 'error', 'warning', 'info'];
    for (const severity of severityOrder) {
        const issues = bySeverity[severity];
        if (!issues || issues.length === 0)
            continue;
        for (const issue of issues) {
            const roundsOld = currentRound - issue.firstSeenRound;
            const roundLabel = roundsOld > 0 ? ` (since Round ${issue.firstSeenRound})` : ' (new)';
            const emoji = getSeverityEmoji(severity);
            const location = issue.line ? `${issue.file}:${issue.line}` : issue.file;
            lines.push(`${emoji} **[${severity.toUpperCase()}]** \`${location}\`${roundLabel}`, `   ${issue.message}`, '');
        }
    }
    return lines.join('\n');
}
/**
 * Determine which files have been "approved" (no issues for N rounds).
 */
export function getApprovedFiles(currentFiles, issueRecords, previousApproved = [], requiredCleanRounds = 1) {
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
export function buildHandoffStatus(currentResult, previousStatus) {
    const currentRound = calculateReviewRound(previousStatus);
    const now = new Date().toISOString();
    // Build issue records with IDs using flatMap instead of nested loop with push
    const issueRecords = currentResult.files.flatMap((file) => file.issues.map((issue) => {
        const id = generateIssueId(file.path, issue.line, issue.message);
        // Check if this issue existed before
        const previousIssue = previousStatus?.issueRecords?.find((p) => p.id === id);
        const firstSeenRound = previousIssue?.firstSeenRound ?? currentRound;
        return {
            id,
            file: file.path,
            line: issue.line,
            severity: issue.severity,
            type: issue.type,
            message: issue.message,
            firstSeenRound,
        };
    }));
    // Calculate resolved/persistent
    const previousIds = new Set(previousStatus?.issueRecords?.map(i => i.id) ?? []);
    const currentIds = new Set(issueRecords.map(i => i.id));
    let resolved = 0;
    let persistent = 0;
    for (const id of previousIds) {
        if (currentIds.has(id)) {
            persistent++;
        }
        else {
            resolved++;
        }
    }
    // Get approved files
    const allFiles = currentResult.files.map(f => f.path);
    const approvedFiles = getApprovedFiles(allFiles, issueRecords, previousStatus?.approvedFiles);
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
export function generateHandoffBlock(status, previousStatus, options = {}) {
    const opts = { ...DEFAULT_OPTIONS, ...options };
    const lines = buildHandoffHeader(status, opts);
    appendInactivityWarning(lines, status, opts);
    appendProgressSection(lines, status, previousStatus);
    appendCriticalWarning(lines, status);
    appendApprovedFilesSection(lines, status);
    return lines.join('\n');
}
function buildHandoffHeader(status, opts) {
    return [
        embedStatusData(status),
        '',
        '## GoReview Status',
        '',
        `> ${generateStatusLine(status, opts)}`,
        '',
    ];
}
function appendInactivityWarning(lines, status, opts) {
    const inactivityWarning = checkInactivity(status.lastReviewAt, opts);
    if (inactivityWarning) {
        lines.push(`⚠️ **Warning:** ${inactivityWarning}`, '');
    }
}
function appendProgressSection(lines, status, previousStatus) {
    if (status.reviewRound <= 1 || !previousStatus) {
        return;
    }
    lines.push('### Progress Since Last Review', '');
    appendProgressStats(lines, status);
    appendPersistentIssuesDetail(lines, status, previousStatus);
}
function appendProgressStats(lines, status) {
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
}
function appendPersistentIssuesDetail(lines, status, previousStatus) {
    if (!status.issueRecords || !previousStatus.issueRecords) {
        return;
    }
    const persistentSection = generatePersistentIssuesSection(status.issueRecords, previousStatus.issueRecords, status.reviewRound);
    if (persistentSection) {
        lines.push(persistentSection);
    }
}
function appendCriticalWarning(lines, status) {
    if (status.criticalIssues > 0) {
        lines.push(`🚨 **${status.criticalIssues} critical issue(s) require immediate attention**`, '');
    }
}
function appendApprovedFilesSection(lines, status) {
    if (!status.approvedFiles || status.approvedFiles.length === 0) {
        return;
    }
    lines.push('<details>', `<summary>✅ ${status.approvedFiles.length} file(s) approved</summary>`, '', ...status.approvedFiles.slice(0, 10).map((file) => `- \`${file}\``));
    if (status.approvedFiles.length > 10) {
        lines.push(`- ... and ${status.approvedFiles.length - 10} more`);
    }
    lines.push('</details>', '');
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
//# sourceMappingURL=statusLine.js.map