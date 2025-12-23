import { logger } from '../utils/logger.js';
import { handlePullRequest } from './pullRequestHandler.js';
import { handleInstallation } from './installationHandler.js';

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

    case 'pull_request_review_comment':
      // Handle review comments
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
