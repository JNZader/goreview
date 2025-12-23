import { logger } from '../utils/logger.js';
import { WebhookPayload } from './webhookHandler.js';
import { clearOctokitCache } from '../services/github.js';

/**
 * Handle installation events.
 */
export async function handleInstallation(
  action: string | undefined,
  payload: WebhookPayload
): Promise<void> {
  const installation = payload.installation as Record<string, unknown> | undefined;
  const installationId = installation?.id as number | undefined;

  logger.info({
    action,
    installationId,
    account: (installation?.account as Record<string, unknown>)?.login,
  }, 'Processing installation event');

  switch (action) {
    case 'created':
      logger.info({ installationId }, 'New installation created');
      break;

    case 'deleted':
      if (installationId) {
        clearOctokitCache(installationId);
      }
      logger.info({ installationId }, 'Installation deleted');
      break;

    case 'suspend':
      logger.info({ installationId }, 'Installation suspended');
      break;

    case 'unsuspend':
      logger.info({ installationId }, 'Installation unsuspended');
      break;

    default:
      logger.debug({ action }, 'Unhandled installation action');
  }
}
