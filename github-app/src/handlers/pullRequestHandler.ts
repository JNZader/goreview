import { getOctokit } from '../services/github.js';
import { prReviewService } from '../services/reviewService.js';
import { commentService } from '../services/commentService.js';
import { loadRepoConfig, RepoConfig } from '../config/repoConfig.js';
import { logger } from '../utils/logger.js';
import { jobQueue } from '../queue/jobQueue.js';

interface PullRequestPayload {
  action: string;
  number: number;
  pull_request: {
    number: number;
    title: string;
    body: string | null;
    state: string;
    draft: boolean;
    head: {
      sha: string;
      ref: string;
    };
    base: {
      ref: string;
    };
    user: {
      login: string;
    };
    changed_files: number;
    additions: number;
    deletions: number;
  };
  repository: {
    id: number;
    full_name: string;
    owner: {
      login: string;
    };
    name: string;
  };
  installation: {
    id: number;
  };
  sender: {
    login: string;
  };
}

/**
 * Handle pull request events.
 */
export async function handlePullRequest(
  action: string | undefined,
  payload: unknown
): Promise<void> {
  const pr = payload as PullRequestPayload;

  // Only handle relevant actions
  const relevantActions = ['opened', 'synchronize', 'reopened'];
  if (!action || !relevantActions.includes(action)) {
    logger.debug({ action }, 'Ignoring PR action');
    return;
  }

  const { repository, pull_request, installation } = pr;
  const owner = repository.owner.login;
  const repo = repository.name;
  const pullNumber = pull_request.number;

  logger.info({
    owner,
    repo,
    pullNumber,
    action,
    title: pull_request.title,
  }, 'Processing pull request');

  // Skip draft PRs
  if (pull_request.draft) {
    logger.info({ pullNumber }, 'Skipping draft PR');
    return;
  }

  // Get Octokit for this installation
  const octokit = await getOctokit(installation.id);

  // Load repository configuration
  const repoConfig = await loadRepoConfig(octokit, owner, repo);

  // Check if auto-review is enabled
  if (!repoConfig.review.auto_review) {
    logger.info({ owner, repo }, 'Auto-review disabled for repository');
    return;
  }

  // Validate PR is reviewable
  const validation = validatePR(pull_request, repoConfig);
  if (!validation.valid) {
    logger.info({
      pullNumber,
      reason: validation.reason,
    }, 'PR not eligible for review');
    return;
  }

  // Queue the review job
  await jobQueue.add({
    type: 'pr_review',
    data: {
      installationId: installation.id,
      owner,
      repo,
      pullNumber,
      headSha: pull_request.head.sha,
    },
  });

  logger.info({ owner, repo, pullNumber }, 'PR review job queued');
}

interface ValidationResult {
  valid: boolean;
  reason?: string;
}

function validatePR(
  pr: PullRequestPayload['pull_request'],
  config: RepoConfig
): ValidationResult {
  // Check if too many files
  if (pr.changed_files > config.review.max_files) {
    return {
      valid: false,
      reason: `Too many files: ${pr.changed_files} > ${config.review.max_files}`,
    };
  }

  // Check if too large
  const totalChanges = pr.additions + pr.deletions;
  if (totalChanges > 10000) {
    return {
      valid: false,
      reason: `Too many changes: ${totalChanges} lines`,
    };
  }

  return { valid: true };
}

/**
 * Process a queued PR review job.
 */
export async function processReviewJob(
  installationId: number,
  owner: string,
  repo: string,
  pullNumber: number,
  headSha: string
): Promise<void> {
  const startTime = Date.now();

  try {
    const octokit = await getOctokit(installationId);

    // Verify the PR still exists and SHA matches
    const { data: currentPR } = await octokit.pulls.get({
      owner,
      repo,
      pull_number: pullNumber,
    });

    if (currentPR.head.sha !== headSha) {
      logger.info({
        pullNumber,
        expectedSha: headSha,
        actualSha: currentPR.head.sha,
      }, 'PR has new commits, skipping stale review');
      return;
    }

    // Set commit status to pending
    await octokit.repos.createCommitStatus({
      owner,
      repo,
      sha: headSha,
      state: 'pending',
      context: 'ai-review',
      description: 'AI code review in progress...',
    });

    // Perform the review
    const result = await prReviewService.reviewPR(
      octokit,
      owner,
      repo,
      pullNumber
    );

    // Post review comments
    await commentService.postReview(
      octokit,
      owner,
      repo,
      pullNumber,
      result
    );

    // Update commit status
    const state = result.criticalIssues > 0 ? 'failure' : 'success';
    const description = result.criticalIssues > 0
      ? `Found ${result.criticalIssues} critical issue(s)`
      : `Review complete. Score: ${result.overallScore}/100`;

    await octokit.repos.createCommitStatus({
      owner,
      repo,
      sha: headSha,
      state,
      context: 'ai-review',
      description,
    });

    const duration = Date.now() - startTime;
    logger.info({
      owner,
      repo,
      pullNumber,
      duration,
      filesReviewed: result.filesReviewed,
      totalIssues: result.totalIssues,
    }, 'PR review completed successfully');

  } catch (error) {
    logger.error({ error, owner, repo, pullNumber }, 'PR review failed');

    // Try to update status to error
    try {
      const octokit = await getOctokit(installationId);
      await octokit.repos.createCommitStatus({
        owner,
        repo,
        sha: headSha,
        state: 'error',
        context: 'ai-review',
        description: 'Review failed. Please try again.',
      });
    } catch {
      // Ignore status update errors
    }

    throw error;
  }
}
