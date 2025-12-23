import { describe, it, expect, vi, beforeEach } from 'vitest';
import { loadRepoConfig, clearRepoConfigCache } from './repoConfig.js';

// Mock logger
vi.mock('../utils/logger.js', () => ({
  logger: {
    debug: vi.fn(),
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

// Mock Octokit
const mockOctokit = {
  repos: {
    getContent: vi.fn(),
  },
};

describe('Repository Configuration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    clearRepoConfigCache('owner', 'repo');
  });

  it('returns defaults when no config file', async () => {
    mockOctokit.repos.getContent.mockRejectedValue({ status: 404 });

    const config = await loadRepoConfig(
      mockOctokit as any,
      'owner',
      'repo'
    );

    expect(config.review.enabled).toBe(true);
    expect(config.rules.preset).toBe('standard');
  });

  it('parses valid YAML config', async () => {
    const yamlContent = `
version: "1.0"
review:
  enabled: true
  max_files: 100
rules:
  preset: strict
`;

    mockOctokit.repos.getContent.mockResolvedValue({
      data: {
        content: Buffer.from(yamlContent).toString('base64'),
      },
    });

    const config = await loadRepoConfig(
      mockOctokit as any,
      'owner',
      'repo'
    );

    expect(config.review.max_files).toBe(100);
    expect(config.rules.preset).toBe('strict');
  });

  it('caches configuration', async () => {
    const yamlContent = 'version: "1.0"';

    mockOctokit.repos.getContent.mockResolvedValue({
      data: {
        content: Buffer.from(yamlContent).toString('base64'),
      },
    });

    await loadRepoConfig(mockOctokit as any, 'owner', 'repo');
    await loadRepoConfig(mockOctokit as any, 'owner', 'repo');

    // Should only call API once due to caching
    expect(mockOctokit.repos.getContent).toHaveBeenCalledTimes(1);
  });

  it('clears cache for specific repo', async () => {
    const yamlContent = 'version: "1.0"';

    mockOctokit.repos.getContent.mockResolvedValue({
      data: {
        content: Buffer.from(yamlContent).toString('base64'),
      },
    });

    await loadRepoConfig(mockOctokit as any, 'owner', 'repo');
    clearRepoConfigCache('owner', 'repo');
    await loadRepoConfig(mockOctokit as any, 'owner', 'repo');

    // Should call API twice after cache clear
    expect(mockOctokit.repos.getContent).toHaveBeenCalledTimes(2);
  });

  it('applies default values for missing config fields', async () => {
    const yamlContent = `
review:
  enabled: false
`;

    mockOctokit.repos.getContent.mockResolvedValue({
      data: {
        content: Buffer.from(yamlContent).toString('base64'),
      },
    });

    const config = await loadRepoConfig(
      mockOctokit as any,
      'owner',
      'repo'
    );

    expect(config.review.enabled).toBe(false);
    expect(config.review.auto_review).toBe(true); // Default
    expect(config.comments.inline).toBe(true); // Default
    expect(config.labels.critical).toBe('needs-attention'); // Default
  });

  it('handles different refs', async () => {
    const yamlContent = 'version: "1.0"';

    mockOctokit.repos.getContent.mockResolvedValue({
      data: {
        content: Buffer.from(yamlContent).toString('base64'),
      },
    });

    await loadRepoConfig(mockOctokit as any, 'owner', 'repo', 'main');
    await loadRepoConfig(mockOctokit as any, 'owner', 'repo', 'develop');

    // Should call API twice for different refs
    expect(mockOctokit.repos.getContent).toHaveBeenCalledTimes(2);
  });
});
