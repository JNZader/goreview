import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock the logger first
vi.mock('../utils/logger.js', () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock the config
vi.mock('../config/index.js', () => ({
  config: {
    nodeEnv: 'test',
    port: 3000,
    logLevel: 'info',
    isDevelopment: false,
    isProduction: false,
    github: {
      appId: 12345,
      privateKey: 'test-key',
      webhookSecret: 'test-secret',
    },
    ai: {
      provider: 'ollama',
      model: 'qwen2.5-coder:14b',
      ollamaBaseUrl: 'http://localhost:11434',
    },
    rateLimit: { rps: 10, burst: 20 },
    cache: { ttl: 3600000, maxEntries: 1000 },
    review: { maxFiles: 50, maxDiffSize: 500000, timeout: 300000 },
  },
}));

// Mock the github service
vi.mock('../services/github.js', () => ({
  getOctokit: vi.fn(),
  clearOctokitCache: vi.fn(),
}));

// Mock the repoConfig
vi.mock('../config/repoConfig.js', () => ({
  loadRepoConfig: vi.fn(),
  clearRepoConfigCache: vi.fn(),
}));

// Mock the job queue
vi.mock('../queue/jobQueue.js', () => ({
  jobQueue: {
    add: vi.fn().mockResolvedValue('job_123'),
    getJob: vi.fn(),
    getStats: vi.fn(),
  },
}));

import { handlePullRequest } from './pullRequestHandler.js';
import { logger } from '../utils/logger.js';
import { getOctokit } from '../services/github.js';
import { loadRepoConfig } from '../config/repoConfig.js';
import { jobQueue } from '../queue/jobQueue.js';

describe('handlePullRequest', () => {
  beforeEach(() => {
    vi.clearAllMocks();

    // Default mock implementations
    (getOctokit as ReturnType<typeof vi.fn>).mockResolvedValue({
      pulls: { get: vi.fn() },
      repos: { createCommitStatus: vi.fn() },
    });

    (loadRepoConfig as ReturnType<typeof vi.fn>).mockResolvedValue({
      review: {
        auto_review: true,
        max_files: 50,
      },
      comments: {
        inline: true,
        summary: true,
        request_changes: true,
        min_severity: 'warning',
      },
      labels: {
        add_on_issues: true,
        critical: 'critical-issues',
        reviewed: 'ai-reviewed',
      },
    });
  });

  const createPayload = (options: {
    action?: string;
    number?: number;
    draft?: boolean;
    changedFiles?: number;
    additions?: number;
    deletions?: number;
  } = {}) => ({
    action: options.action || 'opened',
    number: options.number || 42,
    pull_request: {
      number: options.number || 42,
      title: 'Test PR',
      body: 'Test description',
      state: 'open',
      draft: options.draft || false,
      head: { sha: 'abc123', ref: 'feature-branch' },
      base: { ref: 'main' },
      user: { login: 'testuser' },
      changed_files: options.changedFiles || 5,
      additions: options.additions || 100,
      deletions: options.deletions || 50,
    },
    repository: {
      id: 123,
      full_name: 'owner/repo',
      owner: { login: 'owner' },
      name: 'repo',
    },
    installation: { id: 12345 },
    sender: { login: 'testuser' },
  });

  it('should process PR opened event', async () => {
    const payload = createPayload({ action: 'opened' });

    await handlePullRequest('opened', payload);

    expect(logger.info).toHaveBeenCalledWith(
      expect.objectContaining({
        owner: 'owner',
        repo: 'repo',
        pullNumber: 42,
        action: 'opened',
      }),
      'Processing pull request'
    );
    expect(jobQueue.add).toHaveBeenCalledWith({
      type: 'pr_review',
      data: {
        installationId: 12345,
        owner: 'owner',
        repo: 'repo',
        pullNumber: 42,
        headSha: 'abc123',
      },
    });
  });

  it('should process PR synchronize event', async () => {
    const payload = createPayload({ action: 'synchronize' });

    await handlePullRequest('synchronize', payload);

    expect(jobQueue.add).toHaveBeenCalled();
    expect(logger.info).toHaveBeenCalledWith(
      expect.objectContaining({ action: 'synchronize' }),
      'Processing pull request'
    );
  });

  it('should process PR reopened event', async () => {
    const payload = createPayload({ action: 'reopened' });

    await handlePullRequest('reopened', payload);

    expect(jobQueue.add).toHaveBeenCalled();
  });

  it('should ignore irrelevant PR actions', async () => {
    const payload = createPayload({ action: 'closed' });

    await handlePullRequest('closed', payload);

    expect(logger.debug).toHaveBeenCalledWith(
      { action: 'closed' },
      'Ignoring PR action'
    );
    expect(jobQueue.add).not.toHaveBeenCalled();
  });

  it('should skip draft PRs', async () => {
    const payload = createPayload({ draft: true });

    await handlePullRequest('opened', payload);

    expect(logger.info).toHaveBeenCalledWith(
      { pullNumber: 42 },
      'Skipping draft PR'
    );
    expect(jobQueue.add).not.toHaveBeenCalled();
  });

  it('should skip when auto-review is disabled', async () => {
    (loadRepoConfig as ReturnType<typeof vi.fn>).mockResolvedValue({
      review: { auto_review: false, max_files: 50 },
      comments: {},
      labels: {},
    });

    const payload = createPayload();

    await handlePullRequest('opened', payload);

    expect(logger.info).toHaveBeenCalledWith(
      { owner: 'owner', repo: 'repo' },
      'Auto-review disabled for repository'
    );
    expect(jobQueue.add).not.toHaveBeenCalled();
  });

  it('should skip PRs with too many files', async () => {
    (loadRepoConfig as ReturnType<typeof vi.fn>).mockResolvedValue({
      review: { auto_review: true, max_files: 5 },
      comments: {},
      labels: {},
    });

    const payload = createPayload({ changedFiles: 10 });

    await handlePullRequest('opened', payload);

    expect(logger.info).toHaveBeenCalledWith(
      expect.objectContaining({
        pullNumber: 42,
        reason: 'Too many files: 10 > 5',
      }),
      'PR not eligible for review'
    );
    expect(jobQueue.add).not.toHaveBeenCalled();
  });

  it('should skip PRs with too many changes', async () => {
    const payload = createPayload({ additions: 8000, deletions: 3000 });

    await handlePullRequest('opened', payload);

    expect(logger.info).toHaveBeenCalledWith(
      expect.objectContaining({
        pullNumber: 42,
        reason: 'Too many changes: 11000 lines',
      }),
      'PR not eligible for review'
    );
    expect(jobQueue.add).not.toHaveBeenCalled();
  });

  it('should log when job is queued', async () => {
    const payload = createPayload();

    await handlePullRequest('opened', payload);

    expect(logger.info).toHaveBeenCalledWith(
      { owner: 'owner', repo: 'repo', pullNumber: 42 },
      'PR review job queued'
    );
  });
});
