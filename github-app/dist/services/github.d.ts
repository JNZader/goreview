import { Octokit } from '@octokit/rest';
/**
 * Get an authenticated Octokit instance for an installation.
 */
export declare function getOctokit(installationId: number): Promise<Octokit>;
/**
 * Clear cached Octokit instance for an installation.
 */
export declare function clearOctokitCache(installationId: number): void;
/**
 * Get the app's Octokit instance (not installation-specific).
 */
export declare function getAppOctokit(): Octokit;
//# sourceMappingURL=github.d.ts.map