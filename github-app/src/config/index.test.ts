import { describe, it, expect } from 'vitest';
import { envSchema } from './schema.js';

describe('Environment Schema', () => {
  const validEnv = {
    GITHUB_APP_ID: '12345',
    GITHUB_PRIVATE_KEY: '-----BEGIN RSA PRIVATE KEY-----\\ntest\\n-----END RSA PRIVATE KEY-----',
    GITHUB_WEBHOOK_SECRET: 'a-very-long-secret-key-here',
  };

  it('validates required fields', () => {
    const result = envSchema.safeParse(validEnv);
    expect(result.success).toBe(true);
  });

  it('rejects missing required fields', () => {
    const result = envSchema.safeParse({});
    expect(result.success).toBe(false);
  });

  it('applies defaults', () => {
    const result = envSchema.safeParse(validEnv);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.PORT).toBe(3000);
      expect(result.data.NODE_ENV).toBe('development');
      expect(result.data.AI_PROVIDER).toBe('ollama');
    }
  });

  it('transforms private key newlines', () => {
    const result = envSchema.safeParse(validEnv);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.GITHUB_PRIVATE_KEY).toContain('\n');
    }
  });

  it('parses duration strings', () => {
    const envWithDuration = {
      ...validEnv,
      CACHE_TTL: '30m',
    };
    const result = envSchema.safeParse(envWithDuration);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.CACHE_TTL).toBe(30 * 60 * 1000);
    }
  });

  it('rejects invalid duration format', () => {
    const envWithBadDuration = {
      ...validEnv,
      CACHE_TTL: 'invalid',
    };
    const result = envSchema.safeParse(envWithBadDuration);
    expect(result.success).toBe(false);
  });

  it('coerces numeric strings', () => {
    const envWithStrings = {
      ...validEnv,
      PORT: '8080',
      GITHUB_APP_ID: '99999',
    };
    const result = envSchema.safeParse(envWithStrings);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.PORT).toBe(8080);
      expect(result.data.GITHUB_APP_ID).toBe(99999);
    }
  });
});
