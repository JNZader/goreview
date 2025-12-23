import { Octokit } from '@octokit/rest';
import { createAppAuth } from '@octokit/auth-app';
import { config } from '../config/index.js';
import { logger } from '../utils/logger.js';

// Octokit instance cache per installation
const octokitCache = new Map<number, Octokit>();

/**
 * Get an authenticated Octokit instance for an installation.
 */
export async function getOctokit(installationId: number): Promise<Octokit> {
  if (octokitCache.has(installationId)) {
    return octokitCache.get(installationId)!;
  }

  const octokit = new Octokit({
    authStrategy: createAppAuth,
    auth: {
      appId: config.github.appId,
      privateKey: config.github.privateKey,
      installationId,
    },
    log: {
      debug: (msg: string) => logger.debug(msg),
      info: (msg: string) => logger.info(msg),
      warn: (msg: string) => logger.warn(msg),
      error: (msg: string) => logger.error(msg),
    },
  });

  octokitCache.set(installationId, octokit);
  return octokit;
}

/**
 * Clear cached Octokit instance for an installation.
 */
export function clearOctokitCache(installationId: number): void {
  octokitCache.delete(installationId);
}

/**
 * Get the app's Octokit instance (not installation-specific).
 */
export function getAppOctokit(): Octokit {
  return new Octokit({
    authStrategy: createAppAuth,
    auth: {
      appId: config.github.appId,
      privateKey: config.github.privateKey,
    },
  });
}
