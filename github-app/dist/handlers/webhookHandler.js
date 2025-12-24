import { logger } from '../utils/logger.js';
import { handlePullRequest } from './pullRequestHandler.js';
import { handleInstallation } from './installationHandler.js';
import { handleIssueComment, handlePullRequestReviewComment } from './commentHandler.js';
/**
 * Route webhook events to appropriate handlers.
 */
export async function handleWebhook(event, payload) {
    const action = payload.action;
    logger.debug({ event, action }, 'Processing webhook');
    switch (event) {
        case 'pull_request':
            await handlePullRequest(action, payload);
            break;
        case 'issue_comment':
            await handleIssueComment(payload);
            break;
        case 'pull_request_review_comment':
            await handlePullRequestReviewComment(payload);
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
//# sourceMappingURL=webhookHandler.js.map