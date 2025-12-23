import { logger } from '../utils/logger.js';
import { WebhookPayload } from './webhookHandler.js';

/**
 * Handle pull request events.
 */
export async function handlePullRequest(
  action: string | undefined,
  payload: WebhookPayload
): Promise<void> {
  const pr = payload.pull_request as Record<string, unknown> | undefined;
  const repo = payload.repository as Record<string, unknown> | undefined;

  logger.info({
    action,
    pr: pr?.number,
    repo: repo?.full_name,
  }, 'Processing pull request event');

  switch (action) {
    case 'opened':
    case 'synchronize':
      // Will trigger code review
      logger.info('PR opened or updated - review will be triggered');
      break;

    case 'closed':
      logger.info('PR closed');
      break;

    default:
      logger.debug({ action }, 'Unhandled PR action');
  }
}
