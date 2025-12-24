import { logger } from '../utils/logger.js';
import { handlePullRequest } from './pullRequestHandler.js';
import { handleInstallation } from './installationHandler.js';
import { handleIssueComment, handlePullRequestReviewComment } from './commentHandler.js';
import type { IssueCommentEvent } from '@octokit/webhooks-types';

export type WebhookPayload = Record<string, unknown>;

/**
 * Route webhook events to appropriate handlers.
 */
export async function handleWebhook(
  event: string,
  payload: WebhookPayload
): Promise<void> {
  const action = payload.action as string | undefined;

  logger.debug({ event, action }, 'Processing webhook');

  switch (event) {
    case 'pull_request':
      await handlePullRequest(action, payload);
      break;

    case 'issue_comment':
      await handleIssueComment(payload as unknown as IssueCommentEvent);
      break;

    case 'pull_request_review_comment':
      await handlePullRequestReviewComment(payload as Parameters<typeof handlePullRequestReviewComment>[0]);
      break;

    case 'installation':
    case 'installation_repositories':
      await handleInstallation(action, payload);
      break;

    case 'ping':
      logger.info('Ping event received');
      break;

    default:
      logger.debug({ event }, 'Unhandled event type');
  }
}
