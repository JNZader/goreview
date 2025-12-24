import { describe, it, expect, vi, beforeEach } from 'vitest';
import { handleInstallation } from './installationHandler.js';

// Mock the logger
vi.mock('../utils/logger.js', () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock the github service
vi.mock('../services/github.js', () => ({
  clearOctokitCache: vi.fn(),
}));

// Mock the repoConfig service
vi.mock('../config/repoConfig.js', () => ({
  clearRepoConfigCache: vi.fn(),
}));

import { logger } from '../utils/logger.js';
import { clearOctokitCache } from '../services/github.js';
import { clearRepoConfigCache } from '../config/repoConfig.js';

describe('handleInstallation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const createPayload = (
    action: string,
    installationId: number = 12345,
    accountLogin: string = 'testuser',
    accountType: string = 'User',
    repositories?: Array<{ id: number; name: string; full_name: string; private: boolean }>
  ) => ({
    action,
    installation: {
      id: installationId,
      account: { login: accountLogin, type: accountType },
    },
    repositories,
    sender: { login: 'sender' },
  });

  it('should log installation created event', async () => {
    const payload = createPayload('created', 12345, 'testuser', 'User', [
      { id: 1, name: 'repo1', full_name: 'testuser/repo1', private: false },
    ]);

    await handleInstallation('created', payload);

    expect(logger.info).toHaveBeenCalledWith(
      {
        action: 'created',
        installationId: 12345,
        account: 'testuser',
        accountType: 'User',
      },
      'Installation event received'
    );
    expect(logger.info).toHaveBeenCalledWith(
      {
        installationId: 12345,
        account: 'testuser',
        repoCount: 1,
      },
      'App installed'
    );
  });

  it('should log each repository on installation created', async () => {
    const payload = createPayload('created', 12345, 'testuser', 'User', [
      { id: 1, name: 'repo1', full_name: 'testuser/repo1', private: false },
      { id: 2, name: 'repo2', full_name: 'testuser/repo2', private: true },
    ]);

    await handleInstallation('created', payload);

    expect(logger.info).toHaveBeenCalledWith(
      {
        installationId: 12345,
        repo: 'testuser/repo1',
        private: false,
      },
      'Repository added'
    );
    expect(logger.info).toHaveBeenCalledWith(
      {
        installationId: 12345,
        repo: 'testuser/repo2',
        private: true,
      },
      'Repository added'
    );
  });

  it('should clear octokit cache on installation deleted', async () => {
    const payload = createPayload('deleted');

    await handleInstallation('deleted', payload);

    expect(clearOctokitCache).toHaveBeenCalledWith(12345);
    expect(logger.info).toHaveBeenCalledWith(
      {
        installationId: 12345,
        account: 'testuser',
      },
      'App uninstalled'
    );
  });

  it('should clear repo config cache on repositories added', async () => {
    const payload = createPayload('added', 12345, 'testuser', 'User', [
      { id: 1, name: 'repo1', full_name: 'testuser/repo1', private: false },
    ]);

    await handleInstallation('added', payload);

    expect(clearRepoConfigCache).toHaveBeenCalledWith('testuser', 'repo1');
    expect(logger.info).toHaveBeenCalledWith(
      {
        installationId: 12345,
        repo: 'testuser/repo1',
        private: false,
      },
      'Repository added'
    );
  });

  it('should clear repo config cache on repositories removed', async () => {
    const payload = createPayload('removed', 12345, 'testuser', 'User', [
      { id: 1, name: 'repo1', full_name: 'testuser/repo1', private: false },
    ]);

    await handleInstallation('removed', payload);

    expect(clearRepoConfigCache).toHaveBeenCalledWith('testuser', 'repo1');
    expect(logger.info).toHaveBeenCalledWith(
      {
        installationId: 12345,
        repo: 'testuser/repo1',
        private: false,
      },
      'Repository removed'
    );
  });

  it('should log debug for unhandled installation actions', async () => {
    const payload = createPayload('suspend');

    await handleInstallation('suspend', payload);

    expect(logger.debug).toHaveBeenCalledWith(
      { action: 'suspend' },
      'Unhandled installation action'
    );
  });

  it('should handle organization accounts', async () => {
    const payload = createPayload('created', 99999, 'my-org', 'Organization');

    await handleInstallation('created', payload);

    expect(logger.info).toHaveBeenCalledWith(
      {
        action: 'created',
        installationId: 99999,
        account: 'my-org',
        accountType: 'Organization',
      },
      'Installation event received'
    );
  });
});
