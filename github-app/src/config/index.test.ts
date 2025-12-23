import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

describe('Config', () => {
  const originalEnv = process.env;

  beforeEach(() => {
    vi.resetModules();
    process.env = { ...originalEnv };
  });

  afterEach(() => {
    process.env = originalEnv;
  });

  it('should use empty strings for missing github env vars', async () => {
    delete process.env.GITHUB_APP_ID;
    delete process.env.GITHUB_PRIVATE_KEY;
    delete process.env.GITHUB_WEBHOOK_SECRET;

    const { config } = await import('./index.js');

    expect(config.github.appId).toBe('');
    expect(config.github.privateKey).toBe('');
    expect(config.github.webhookSecret).toBe('');
  });

  it('should use default values for optional env vars', async () => {
    const { config } = await import('./index.js');

    expect(config.port).toBe(3000);
    expect(config.logLevel).toBe('info');
    expect(config.ollama.baseUrl).toBe('http://localhost:11434');
    expect(config.ollama.model).toBe('codellama');
  });

  it('should parse env vars correctly', async () => {
    process.env.GITHUB_APP_ID = '12345';
    process.env.GITHUB_PRIVATE_KEY = 'test-private-key';
    process.env.GITHUB_WEBHOOK_SECRET = 'test-webhook-secret';
    process.env.PORT = '8080';
    process.env.NODE_ENV = 'production';
    process.env.LOG_LEVEL = 'debug';

    const { config } = await import('./index.js');

    expect(config.github.appId).toBe('12345');
    expect(config.github.privateKey).toBe('test-private-key');
    expect(config.github.webhookSecret).toBe('test-webhook-secret');
    expect(config.port).toBe(8080);
    expect(config.nodeEnv).toBe('production');
    expect(config.logLevel).toBe('debug');
    expect(config.isDevelopment).toBe(false);
  });

  it('should set isDevelopment based on NODE_ENV', async () => {
    process.env.NODE_ENV = 'development';

    const { config } = await import('./index.js');

    expect(config.isDevelopment).toBe(true);
  });
});
