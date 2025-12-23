/**
 * Tests for secure secret handling utilities
 */

import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  encryptSecret,
  decryptSecret,
  maskSecret,
  maskSecrets,
  safeStringify,
  secureCompare,
  generateSecureToken,
  generateUrlSafeToken,
  hashValue,
  createHmac,
  verifyHmac,
} from './secrets.js';

describe('secrets', () => {
  // ===========================================================================
  // Encryption Tests
  // ===========================================================================

  describe('encryptSecret / decryptSecret', () => {
    const testPassword = 'test-master-password';
    const plaintext = 'super-secret-value';

    beforeEach(() => {
      vi.stubEnv('NODE_ENV', 'test');
    });

    afterEach(() => {
      vi.unstubAllEnvs();
    });

    it('should encrypt and decrypt a secret', () => {
      const encrypted = encryptSecret(plaintext, testPassword);
      const decrypted = decryptSecret(encrypted, testPassword);

      expect(decrypted).toBe(plaintext);
    });

    it('should produce different ciphertext each time (random IV)', () => {
      const encrypted1 = encryptSecret(plaintext, testPassword);
      const encrypted2 = encryptSecret(plaintext, testPassword);

      expect(encrypted1.encrypted).not.toBe(encrypted2.encrypted);
      expect(encrypted1.iv).not.toBe(encrypted2.iv);
      expect(encrypted1.salt).not.toBe(encrypted2.salt);
    });

    it('should fail decryption with wrong password', () => {
      const encrypted = encryptSecret(plaintext, testPassword);

      expect(() => decryptSecret(encrypted, 'wrong-password')).toThrow();
    });

    it('should fail decryption with tampered ciphertext', () => {
      const encrypted = encryptSecret(plaintext, testPassword);
      encrypted.encrypted = encrypted.encrypted.replace('a', 'b');

      expect(() => decryptSecret(encrypted, testPassword)).toThrow();
    });

    it('should fail decryption with tampered auth tag', () => {
      const encrypted = encryptSecret(plaintext, testPassword);
      encrypted.authTag = encrypted.authTag.replace('a', 'b');

      expect(() => decryptSecret(encrypted, testPassword)).toThrow();
    });

    it('should handle unicode characters', () => {
      const unicodeText = 'Hello 世界 🔐 مرحبا';
      const encrypted = encryptSecret(unicodeText, testPassword);
      const decrypted = decryptSecret(encrypted, testPassword);

      expect(decrypted).toBe(unicodeText);
    });

    it('should handle empty string', () => {
      const encrypted = encryptSecret('', testPassword);
      const decrypted = decryptSecret(encrypted, testPassword);

      expect(decrypted).toBe('');
    });

    it('should throw in production without ENCRYPTION_KEY', () => {
      vi.stubEnv('NODE_ENV', 'production');
      vi.stubEnv('ENCRYPTION_KEY', '');

      expect(() => encryptSecret('test')).toThrow('ENCRYPTION_KEY');
    });

    it('should use ENCRYPTION_KEY from environment when available', () => {
      vi.stubEnv('ENCRYPTION_KEY', 'env-key-for-testing');

      const encrypted = encryptSecret(plaintext);
      const decrypted = decryptSecret(encrypted);

      expect(decrypted).toBe(plaintext);
    });
  });

  // ===========================================================================
  // Masking Tests
  // ===========================================================================

  describe('maskSecret', () => {
    it('should mask middle part of secret', () => {
      const secret = 'abcdefghijklmnop';
      const masked = maskSecret(secret);

      expect(masked).toMatch(/^abcd\*+mnop$/);
      expect(masked.length).toBe(secret.length);
    });

    it('should respect showFirst/showLast options', () => {
      const secret = 'abcdefghijklmnop';
      const masked = maskSecret(secret, { showFirst: 2, showLast: 2 });

      expect(masked).toMatch(/^ab\*+op$/);
    });

    it('should use custom mask character', () => {
      const secret = 'abcdefghijklmnop';
      const masked = maskSecret(secret, { maskChar: 'X' });

      expect(masked).toMatch(/^abcdX+mnop$/);
    });

    it('should mask short secrets completely', () => {
      const shortSecret = 'short';
      const masked = maskSecret(shortSecret, { minLength: 12 });

      expect(masked).toBe('********');
    });

    it('should handle empty string', () => {
      const masked = maskSecret('');

      expect(masked).toBe('********');
    });

    it('should handle null/undefined gracefully', () => {
      expect(maskSecret(null as unknown as string)).toBe('********');
      expect(maskSecret(undefined as unknown as string)).toBe('********');
    });
  });

  describe('maskSecrets', () => {
    it('should mask GitHub PAT tokens', () => {
      // GitHub PAT format: ghp_ + 36 alphanumeric chars
      const text = 'token: ghp_abcdefghijklmnopqrstuvwxyz1234567890';
      const masked = maskSecrets(text);

      expect(masked).not.toContain('ghp_abcdefghijklmnopqrstuvwxyz1234567890');
      expect(masked).toContain('****');
    });

    it('should mask GitHub App tokens', () => {
      // GitHub App token format: ghs_ + 36 alphanumeric chars
      const text = 'ghs_abcdefghijklmnopqrstuvwxyz1234567890';
      const masked = maskSecrets(text);

      expect(masked).not.toContain('ghs_abcdefghijklmnopqrstuvwxyz1234567890');
    });

    it('should mask Bearer tokens', () => {
      const text = 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9';
      const masked = maskSecrets(text);

      expect(masked).not.toContain('eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9');
    });

    it('should mask private keys', () => {
      const text = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0m59l2u9iDnMbrXHfqkOrn2dVQ
-----END RSA PRIVATE KEY-----`;
      const masked = maskSecrets(text);

      expect(masked).not.toContain('MIIEowIBAAKCAQEA0m59l2u9iDnMbrXHfqkOrn2dVQ');
    });

    it('should mask email addresses', () => {
      const text = 'Contact: user@example.com for support';
      const masked = maskSecrets(text);

      expect(masked).not.toContain('user@example.com');
    });

    it('should handle multiple sensitive patterns', () => {
      // Use correct format: ghp_ + 36 chars
      const text = 'PAT: ghp_abcdefghijklmnopqrstuvwxyz1234567890, email: test@test.com';
      const masked = maskSecrets(text);

      expect(masked).not.toContain('ghp_abcdefghijklmnopqrstuvwxyz1234567890');
      expect(masked).not.toContain('test@test.com');
    });

    it('should return original for non-string input', () => {
      expect(maskSecrets(null as unknown as string)).toBeNull();
      expect(maskSecrets(123 as unknown as string)).toBe(123);
    });
  });

  describe('safeStringify', () => {
    it('should mask sensitive object values', () => {
      const obj = {
        username: 'john',
        password: 'secret123',
        apiKey: 'sk-abcdefgh12345678',
      };

      const result = safeStringify(obj);
      const parsed = JSON.parse(result);

      expect(parsed.username).toBe('john');
      expect(parsed.password).toContain('****');
      expect(parsed.apiKey).toContain('****');
    });

    it('should handle nested objects', () => {
      const obj = {
        user: {
          name: 'john',
          credentials: {
            token: 'secret-token',
          },
        },
      };

      const result = safeStringify(obj);
      const parsed = JSON.parse(result);

      expect(parsed.user.name).toBe('john');
      expect(parsed.user.credentials.token).toContain('****');
    });

    it('should handle circular references', () => {
      const obj: Record<string, unknown> = { name: 'test' };
      obj.self = obj;

      const result = safeStringify(obj);

      expect(result).toContain('[Circular]');
    });

    it('should use custom sensitive keys', () => {
      const obj = {
        customSecret: 'my-value',
        normalField: 'visible',
      };

      const result = safeStringify(obj, ['customSecret']);
      const parsed = JSON.parse(result);

      expect(parsed.customSecret).toContain('****');
      expect(parsed.normalField).toBe('visible');
    });
  });

  // ===========================================================================
  // Secure Compare Tests
  // ===========================================================================

  describe('secureCompare', () => {
    it('should return true for equal strings', () => {
      expect(secureCompare('abc123', 'abc123')).toBe(true);
    });

    it('should return false for different strings', () => {
      expect(secureCompare('abc123', 'abc124')).toBe(false);
    });

    it('should return false for different length strings', () => {
      expect(secureCompare('abc', 'abcd')).toBe(false);
    });

    it('should return false for non-string inputs', () => {
      expect(secureCompare(null as unknown as string, 'test')).toBe(false);
      expect(secureCompare('test', null as unknown as string)).toBe(false);
      expect(secureCompare(123 as unknown as string, '123')).toBe(false);
    });

    it('should handle empty strings', () => {
      expect(secureCompare('', '')).toBe(true);
      expect(secureCompare('', 'a')).toBe(false);
    });
  });

  // ===========================================================================
  // Token Generation Tests
  // ===========================================================================

  describe('generateSecureToken', () => {
    it('should generate token of specified length', () => {
      const token = generateSecureToken(32);

      expect(token).toHaveLength(64); // hex encoding doubles length
    });

    it('should generate unique tokens', () => {
      const tokens = new Set<string>();
      for (let i = 0; i < 100; i++) {
        tokens.add(generateSecureToken());
      }

      expect(tokens.size).toBe(100);
    });

    it('should only contain hex characters', () => {
      const token = generateSecureToken();

      expect(token).toMatch(/^[a-f0-9]+$/);
    });
  });

  describe('generateUrlSafeToken', () => {
    it('should generate token of specified length', () => {
      const token = generateUrlSafeToken(32);

      expect(token).toHaveLength(32);
    });

    it('should be URL safe (no +, /, =)', () => {
      // Generate many tokens to increase chance of catching unsafe chars
      for (let i = 0; i < 100; i++) {
        const token = generateUrlSafeToken(64);

        expect(token).not.toContain('+');
        expect(token).not.toContain('/');
        expect(token).not.toContain('=');
      }
    });
  });

  // ===========================================================================
  // Hashing Tests
  // ===========================================================================

  describe('hashValue', () => {
    it('should produce consistent hash', () => {
      const hash1 = hashValue('test');
      const hash2 = hashValue('test');

      expect(hash1).toBe(hash2);
    });

    it('should produce different hash for different input', () => {
      const hash1 = hashValue('test1');
      const hash2 = hashValue('test2');

      expect(hash1).not.toBe(hash2);
    });

    it('should produce 64 character hex string', () => {
      const hash = hashValue('test');

      expect(hash).toHaveLength(64);
      expect(hash).toMatch(/^[a-f0-9]+$/);
    });
  });

  describe('createHmac / verifyHmac', () => {
    const secret = 'my-secret-key';
    const value = 'data-to-sign';

    it('should create and verify HMAC', () => {
      const signature = createHmac(value, secret);
      const isValid = verifyHmac(value, signature, secret);

      expect(isValid).toBe(true);
    });

    it('should fail verification with wrong value', () => {
      const signature = createHmac(value, secret);
      const isValid = verifyHmac('wrong-data', signature, secret);

      expect(isValid).toBe(false);
    });

    it('should fail verification with wrong secret', () => {
      const signature = createHmac(value, secret);
      const isValid = verifyHmac(value, signature, 'wrong-secret');

      expect(isValid).toBe(false);
    });

    it('should fail verification with tampered signature', () => {
      const signature = createHmac(value, secret);
      const tamperedSig = signature.replace('a', 'b');
      const isValid = verifyHmac(value, tamperedSig, secret);

      expect(isValid).toBe(false);
    });

    it('should produce 64 character hex signature', () => {
      const signature = createHmac(value, secret);

      expect(signature).toHaveLength(64);
      expect(signature).toMatch(/^[a-f0-9]+$/);
    });
  });
});
