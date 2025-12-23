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

import { logger } from '../utils/logger.js';
import { clearOctokitCache } from '../services/github.js';

describe('handleInstallation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should log installation created event', async () => {
    const payload = {
      action: 'created',
      installation: { id: 12345, account: { login: 'testuser' } },
    };

    await handleInstallation('created', payload);

    expect(logger.info).toHaveBeenCalledWith(
      { action: 'created', installationId: 12345, account: 'testuser' },
      'Processing installation event'
    );
    expect(logger.info).toHaveBeenCalledWith({ installationId: 12345 }, 'New installation created');
  });

  it('should clear cache on installation deleted', async () => {
    const payload = {
      action: 'deleted',
      installation: { id: 12345, account: { login: 'testuser' } },
    };

    await handleInstallation('deleted', payload);

    expect(clearOctokitCache).toHaveBeenCalledWith(12345);
    expect(logger.info).toHaveBeenCalledWith({ installationId: 12345 }, 'Installation deleted');
  });

  it('should not clear cache when installationId is missing on delete', async () => {
    const payload = {
      action: 'deleted',
      installation: { account: { login: 'testuser' } },
    };

    await handleInstallation('deleted', payload);

    expect(clearOctokitCache).not.toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith({ installationId: undefined }, 'Installation deleted');
  });

  it('should log installation suspended event', async () => {
    const payload = {
      action: 'suspend',
      installation: { id: 12345, account: { login: 'testuser' } },
    };

    await handleInstallation('suspend', payload);

    expect(logger.info).toHaveBeenCalledWith({ installationId: 12345 }, 'Installation suspended');
  });

  it('should log installation unsuspended event', async () => {
    const payload = {
      action: 'unsuspend',
      installation: { id: 12345, account: { login: 'testuser' } },
    };

    await handleInstallation('unsuspend', payload);

    expect(logger.info).toHaveBeenCalledWith({ installationId: 12345 }, 'Installation unsuspended');
  });

  it('should log debug for unhandled installation actions', async () => {
    const payload = {
      action: 'new_permissions_accepted',
      installation: { id: 12345, account: { login: 'testuser' } },
    };

    await handleInstallation('new_permissions_accepted', payload);

    expect(logger.debug).toHaveBeenCalledWith(
      { action: 'new_permissions_accepted' },
      'Unhandled installation action'
    );
  });

  it('should handle missing installation data gracefully', async () => {
    const payload = { action: 'created' };

    await handleInstallation('created', payload);

    expect(logger.info).toHaveBeenCalledWith(
      { action: 'created', installationId: undefined, account: undefined },
      'Processing installation event'
    );
  });
});
