import { clearOctokitCache } from '../services/github.js';
import { clearRepoConfigCache } from '../config/repoConfig.js';
import { logger } from '../utils/logger.js';

interface InstallationPayload {
  action: string;
  installation: {
    id: number;
    account: {
      login: string;
      type: string;
    };
  };
  repositories?: Array<{
    id: number;
    name: string;
    full_name: string;
    private: boolean;
  }>;
  sender: {
    login: string;
  };
}

/**
 * Handle installation events.
 */
export async function handleInstallation(
  action: string | undefined,
  payload: unknown
): Promise<void> {
  const installation = payload as InstallationPayload;
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

async function handleInstallationCreated(
  installation: InstallationPayload['installation'],
  repositories: InstallationPayload['repositories']
): Promise<void> {
  logger.info({
    installationId: installation.id,
    account: installation.account.login,
    repoCount: repositories?.length || 0,
  }, 'App installed');

  // Log installed repositories
  if (repositories) {
    for (const repo of repositories) {
      logger.info({
        installationId: installation.id,
        repo: repo.full_name,
        private: repo.private,
      }, 'Repository added to installation');
    }
  }

  // Could send welcome message, initialize settings, etc.
}

async function handleInstallationDeleted(
  installation: InstallationPayload['installation']
): Promise<void> {
  logger.info({
    installationId: installation.id,
    account: installation.account.login,
  }, 'App uninstalled');

  // Clear caches
  clearOctokitCache(installation.id);

  // Could clean up any stored data for this installation
}

async function handleRepositoriesAdded(
  installation: InstallationPayload['installation'],
  repositories: InstallationPayload['repositories']
): Promise<void> {
  if (!repositories) return;

  for (const repo of repositories) {
    logger.info({
      installationId: installation.id,
      repo: repo.full_name,
    }, 'Repository added');

    // Clear any cached config for this repo
    const parts = repo.full_name.split('/');
    const owner = parts[0] ?? '';
    const name = parts[1] ?? '';
    if (owner && name) {
      clearRepoConfigCache(owner, name);
    }
  }
}

async function handleRepositoriesRemoved(
  installation: InstallationPayload['installation'],
  repositories: InstallationPayload['repositories']
): Promise<void> {
  if (!repositories) return;

  for (const repo of repositories) {
    logger.info({
      installationId: installation.id,
      repo: repo.full_name,
    }, 'Repository removed');

    // Clear cached config
    const parts = repo.full_name.split('/');
    const owner = parts[0] ?? '';
    const name = parts[1] ?? '';
    if (owner && name) {
      clearRepoConfigCache(owner, name);
    }
  }
}
