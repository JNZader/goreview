import { describe, it, expect, vi, beforeEach } from 'vitest';
import { handlePullRequest } from './pullRequestHandler.js';

// Mock the logger
vi.mock('../utils/logger.js', () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

import { logger } from '../utils/logger.js';

describe('handlePullRequest', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should log PR opened event', async () => {
    const payload = {
      action: 'opened',
      pull_request: { number: 42 },
      repository: { full_name: 'owner/repo' },
    };

    await handlePullRequest('opened', payload);

    expect(logger.info).toHaveBeenCalledWith(
      { action: 'opened', pr: 42, repo: 'owner/repo' },
      'Processing pull request event'
    );
    expect(logger.info).toHaveBeenCalledWith('PR opened or updated - review will be triggered');
  });

  it('should log PR synchronize event', async () => {
    const payload = {
      action: 'synchronize',
      pull_request: { number: 42 },
      repository: { full_name: 'owner/repo' },
    };

    await handlePullRequest('synchronize', payload);

    expect(logger.info).toHaveBeenCalledWith('PR opened or updated - review will be triggered');
  });

  it('should log PR closed event', async () => {
    const payload = {
      action: 'closed',
      pull_request: { number: 42 },
      repository: { full_name: 'owner/repo' },
    };

    await handlePullRequest('closed', payload);

    expect(logger.info).toHaveBeenCalledWith('PR closed');
  });

  it('should log debug for unhandled PR actions', async () => {
    const payload = {
      action: 'labeled',
      pull_request: { number: 42 },
      repository: { full_name: 'owner/repo' },
    };

    await handlePullRequest('labeled', payload);

    expect(logger.debug).toHaveBeenCalledWith({ action: 'labeled' }, 'Unhandled PR action');
  });

  it('should handle missing pull_request data gracefully', async () => {
    const payload = { action: 'opened' };

    await handlePullRequest('opened', payload);

    expect(logger.info).toHaveBeenCalledWith(
      { action: 'opened', pr: undefined, repo: undefined },
      'Processing pull request event'
    );
  });
});
