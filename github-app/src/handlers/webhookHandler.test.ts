import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock the config FIRST to prevent process.exit
vi.mock('../config/index.js', () => ({
  config: {
    github: {
      appId: '123456',
      privateKey: 'test-key',
      webhookSecret: 'test-secret',
    },
    ai: {
      provider: 'gemini',
      model: 'gemini-2.0-flash',
    },
    server: {
      port: 3000,
    },
  },
}));

// Mock the logger
vi.mock('../utils/logger.js', () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock the handlers
vi.mock('./pullRequestHandler.js', () => ({
  handlePullRequest: vi.fn(),
}));

vi.mock('./installationHandler.js', () => ({
  handleInstallation: vi.fn(),
}));

// Mock comment handlers to avoid github service initialization
vi.mock('./commentHandler.js', () => ({
  handleIssueComment: vi.fn(),
  handlePullRequestReviewComment: vi.fn(),
}));

import { handleWebhook } from './webhookHandler.js';

import { handlePullRequest } from './pullRequestHandler.js';
import { handleInstallation } from './installationHandler.js';
import { logger } from '../utils/logger.js';

describe('handleWebhook', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should route pull_request events to handlePullRequest', async () => {
    const payload = { action: 'opened', pull_request: { number: 1 } };
    await handleWebhook('pull_request', payload);
    expect(handlePullRequest).toHaveBeenCalledWith('opened', payload);
  });

  it('should route installation events to handleInstallation', async () => {
    const payload = { action: 'created', installation: { id: 12345 } };
    await handleWebhook('installation', payload);
    expect(handleInstallation).toHaveBeenCalledWith('created', payload);
  });

  it('should route installation_repositories events to handleInstallation', async () => {
    const payload = { action: 'added', installation: { id: 12345 } };
    await handleWebhook('installation_repositories', payload);
    expect(handleInstallation).toHaveBeenCalledWith('added', payload);
  });

  it('should handle ping events', async () => {
    const payload = { zen: 'test' };
    await handleWebhook('ping', payload);
    expect(logger.info).toHaveBeenCalledWith('Ping event received');
  });

  it('should log debug for unhandled event types', async () => {
    const payload = { action: 'test' };
    await handleWebhook('unknown_event', payload);
    expect(logger.debug).toHaveBeenCalledWith({ event: 'unknown_event' }, 'Unhandled event type');
  });

  it('should handle pull_request_review_comment events', async () => {
    const payload = { action: 'created', comment: { id: 1 } };
    await handleWebhook('pull_request_review_comment', payload);
    // Should not throw and should not call other handlers
    expect(handlePullRequest).not.toHaveBeenCalled();
    expect(handleInstallation).not.toHaveBeenCalled();
  });
});
