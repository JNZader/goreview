import { clearOctokitCache } from '../services/github.js';
import { clearRepoConfigCache } from '../config/repoConfig.js';
import { logger } from '../utils/logger.js';
/** Parse owner and name from full repo name */
function parseRepoFullName(fullName) {
    const parts = fullName.split('/');
    const owner = parts[0] ?? '';
    const name = parts[1] ?? '';
    return owner && name ? { owner, name } : null;
}
/** Log and process repository changes */
function processRepos(installationId, repositories, action, clearCache) {
    if (!repositories)
        return;
    for (const repo of repositories) {
        logger.info({ installationId, repo: repo.full_name, private: repo.private }, `Repository ${action}`);
        if (clearCache) {
            const parsed = parseRepoFullName(repo.full_name);
            if (parsed)
                clearRepoConfigCache(parsed.owner, parsed.name);
        }
    }
}
/**
 * Handle installation events.
 */
export async function handleInstallation(action, payload) {
    const installation = payload;
    const { installation: inst, repositories } = installation;
    logger.info({
        action,
        installationId: inst.id,
        account: inst.account.login,
        accountType: inst.account.type,
    }, 'Installation event received');
    switch (action) {
        case 'created':
            await handleInstallationCreated(inst, repositories);
            break;
        case 'deleted':
            await handleInstallationDeleted(inst);
            break;
        case 'added':
            await handleRepositoriesAdded(inst, repositories);
            break;
        case 'removed':
            await handleRepositoriesRemoved(inst, repositories);
            break;
        default:
            logger.debug({ action }, 'Unhandled installation action');
    }
}
async function handleInstallationCreated(installation, repositories) {
    logger.info({
        installationId: installation.id,
        account: installation.account.login,
        repoCount: repositories?.length || 0,
    }, 'App installed');
    // Log installed repositories (no cache to clear on new install)
    processRepos(installation.id, repositories, 'added', false);
}
async function handleInstallationDeleted(installation) {
    logger.info({
        installationId: installation.id,
        account: installation.account.login,
    }, 'App uninstalled');
    // Clear caches
    clearOctokitCache(installation.id);
    // Could clean up any stored data for this installation
}
async function handleRepositoriesAdded(installation, repositories) {
    processRepos(installation.id, repositories, 'added', true);
}
async function handleRepositoriesRemoved(installation, repositories) {
    processRepos(installation.id, repositories, 'removed', true);
}
//# sourceMappingURL=installationHandler.js.map