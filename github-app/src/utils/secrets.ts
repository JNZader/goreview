/**
 * Secure secret handling utilities
 * Provides encryption, masking, and secure comparison functions
 */

import crypto from 'node:crypto';

// =============================================================================
// Constants
// =============================================================================

const ALGORITHM = 'aes-256-gcm';
const KEY_LENGTH = 32; // 256 bits
const IV_LENGTH = 16; // 128 bits
const AUTH_TAG_LENGTH = 16; // 128 bits
const SALT_LENGTH = 32;
const PBKDF2_ITERATIONS = 100000;

// Patterns for sensitive data detection
const SENSITIVE_PATTERNS = [
  { pattern: /ghp_[a-zA-Z0-9]{36}/g, name: 'GitHub PAT' },
  { pattern: /ghs_[a-zA-Z0-9]{36}/g, name: 'GitHub App Token' },
  { pattern: /ghr_[a-zA-Z0-9]{36}/g, name: 'GitHub Refresh Token' },
  { pattern: /github_pat_[a-zA-Z0-9]{22}_[a-zA-Z0-9]{59}/g, name: 'GitHub Fine-grained PAT' },
  { pattern: /Bearer\s+[a-zA-Z0-9._-]+/gi, name: 'Bearer Token' },
  { pattern: /-----BEGIN[A-Z ]+PRIVATE KEY-----[\s\S]*?-----END[A-Z ]+PRIVATE KEY-----/g, name: 'Private Key' },
  { pattern: /sk-[a-zA-Z0-9]{48}/g, name: 'API Key' },
  { pattern: /[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}/g, name: 'Email' },
  { pattern: /password['":\s]*[=:]\s*['"]?[^'"\s,}]+['"]?/gi, name: 'Password' },
  { pattern: /secret['":\s]*[=:]\s*['"]?[^'"\s,}]+['"]?/gi, name: 'Secret' },
];

// =============================================================================
// Types
// =============================================================================

export interface EncryptedData {
  encrypted: string;
  iv: string;
  authTag: string;
  salt: string;
}

export interface SecretMaskOptions {
  showFirst?: number;
  showLast?: number;
  maskChar?: string;
  minLength?: number;
}

// =============================================================================
// Key Derivation
// =============================================================================

/**
 * Derives an encryption key from a password using PBKDF2
 */
function deriveKey(password: string, salt: Buffer): Buffer {
  return crypto.pbkdf2Sync(password, salt, PBKDF2_ITERATIONS, KEY_LENGTH, 'sha256');
}

/**
 * Gets the master encryption key from environment
 * Falls back to a derived key if ENCRYPTION_KEY is not set
 */
function getMasterKey(): string {
  const key = process.env.ENCRYPTION_KEY;
  if (!key) {
    // In production, this should be a required environment variable
    if (process.env.NODE_ENV === 'production') {
      throw new Error('ENCRYPTION_KEY environment variable is required in production');
    }
    // For development/testing, use a derived key (not secure for production)
    return 'dev-encryption-key-not-for-production';
  }
  return key;
}

// =============================================================================
// Encryption Functions
// =============================================================================

/**
 * Encrypts a secret using AES-256-GCM
 * @param plaintext - The secret to encrypt
 * @param masterPassword - Optional master password (defaults to env var)
 * @returns Encrypted data with IV, auth tag, and salt
 */
export function encryptSecret(plaintext: string, masterPassword?: string): EncryptedData {
  const password = masterPassword ?? getMasterKey();
  const salt = crypto.randomBytes(SALT_LENGTH);
  const iv = crypto.randomBytes(IV_LENGTH);
  const key = deriveKey(password, salt);

  const cipher = crypto.createCipheriv(ALGORITHM, key, iv);

  let encrypted = cipher.update(plaintext, 'utf8', 'hex');
  encrypted += cipher.final('hex');

  const authTag = cipher.getAuthTag();

  return {
    encrypted,
    iv: iv.toString('hex'),
    authTag: authTag.toString('hex'),
    salt: salt.toString('hex'),
  };
}

/**
 * Decrypts a secret encrypted with encryptSecret
 * @param encryptedData - The encrypted data object
 * @param masterPassword - Optional master password (defaults to env var)
 * @returns Decrypted plaintext
 */
export function decryptSecret(encryptedData: EncryptedData, masterPassword?: string): string {
  const password = masterPassword ?? getMasterKey();
  const salt = Buffer.from(encryptedData.salt, 'hex');
  const iv = Buffer.from(encryptedData.iv, 'hex');
  const authTag = Buffer.from(encryptedData.authTag, 'hex');
  const key = deriveKey(password, salt);

  const decipher = crypto.createDecipheriv(ALGORITHM, key, iv);
  decipher.setAuthTag(authTag);

  let decrypted = decipher.update(encryptedData.encrypted, 'hex', 'utf8');
  decrypted += decipher.final('utf8');

  return decrypted;
}

// =============================================================================
// Masking Functions
// =============================================================================

/**
 * Masks a secret string, showing only first/last characters
 * @param secret - The secret to mask
 * @param options - Masking options
 * @returns Masked string
 */
export function maskSecret(secret: string, options: SecretMaskOptions = {}): string {
  const {
    showFirst = 4,
    showLast = 4,
    maskChar = '*',
    minLength = 12,
  } = options;

  if (!secret || secret.length < minLength) {
    return maskChar.repeat(8);
  }

  const firstPart = secret.slice(0, showFirst);
  const lastPart = secret.slice(-showLast);
  const middleLength = Math.max(secret.length - showFirst - showLast, 4);

  return `${firstPart}${maskChar.repeat(middleLength)}${lastPart}`;
}

/**
 * Scans text for sensitive data and masks it
 * @param text - The text to scan and mask
 * @returns Text with sensitive data masked
 */
export function maskSecrets(text: string): string {
  if (!text || typeof text !== 'string') {
    return text;
  }

  let masked = text;

  for (const { pattern } of SENSITIVE_PATTERNS) {
    masked = masked.replace(pattern, (match) => maskSecret(match));
  }

  return masked;
}

/**
 * Safely stringifies an object, masking sensitive values
 * @param obj - Object to stringify
 * @param sensitiveKeys - Keys to mask (default: common sensitive keys)
 * @returns JSON string with masked values
 */
export function safeStringify(
  obj: unknown,
  sensitiveKeys: string[] = ['password', 'secret', 'token', 'key', 'apiKey', 'privateKey', 'authorization']
): string {
  const seen = new WeakSet();

  const replacer = (key: string, value: unknown): unknown => {
    // Handle circular references
    if (typeof value === 'object' && value !== null) {
      if (seen.has(value)) {
        return '[Circular]';
      }
      seen.add(value);
    }

    // Mask sensitive keys
    if (sensitiveKeys.some((k) => key.toLowerCase().includes(k.toLowerCase()))) {
      if (typeof value === 'string') {
        return maskSecret(value);
      }
      return '[REDACTED]';
    }

    // Mask string values that look sensitive
    if (typeof value === 'string') {
      return maskSecrets(value);
    }

    return value;
  };

  return JSON.stringify(obj, replacer, 2);
}

// =============================================================================
// Secure Comparison
// =============================================================================

/**
 * Performs timing-safe string comparison to prevent timing attacks
 * @param a - First string
 * @param b - Second string
 * @returns True if strings are equal
 */
export function secureCompare(a: string, b: string): boolean {
  if (typeof a !== 'string' || typeof b !== 'string') {
    return false;
  }

  // Use a fixed-length comparison to prevent length-based timing attacks
  const bufA = Buffer.from(a);
  const bufB = Buffer.from(b);

  // If lengths differ, compare against itself to maintain constant time
  if (bufA.length !== bufB.length) {
    crypto.timingSafeEqual(bufA, bufA);
    return false;
  }

  return crypto.timingSafeEqual(bufA, bufB);
}

// =============================================================================
// Token Generation
// =============================================================================

/**
 * Generates a cryptographically secure random token
 * @param length - Token length in bytes (output will be hex, so 2x chars)
 * @returns Hex-encoded random token
 */
export function generateSecureToken(length: number = 32): string {
  return crypto.randomBytes(length).toString('hex');
}

/**
 * Generates a URL-safe random token
 * @param length - Desired output length
 * @returns Base64url-encoded random token
 */
export function generateUrlSafeToken(length: number = 32): string {
  // Generate extra bytes to account for base64 encoding overhead
  const bytes = crypto.randomBytes(Math.ceil(length * 0.75));
  return bytes
    .toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=/g, '')
    .slice(0, length);
}

// =============================================================================
// Hashing Functions
// =============================================================================

/**
 * Creates a SHA-256 hash of a value
 * @param value - Value to hash
 * @returns Hex-encoded hash
 */
export function hashValue(value: string): string {
  return crypto.createHash('sha256').update(value).digest('hex');
}

/**
 * Creates a HMAC-SHA256 signature
 * @param value - Value to sign
 * @param secret - Secret key
 * @returns Hex-encoded HMAC
 */
export function createHmac(value: string, secret: string): string {
  return crypto.createHmac('sha256', secret).update(value).digest('hex');
}

/**
 * Verifies a HMAC-SHA256 signature using timing-safe comparison
 * @param value - Original value
 * @param signature - Signature to verify
 * @param secret - Secret key
 * @returns True if signature is valid
 */
export function verifyHmac(value: string, signature: string, secret: string): boolean {
  const expected = createHmac(value, secret);
  return secureCompare(expected, signature);
}

// =============================================================================
// Exports
// =============================================================================

export const secrets = {
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
};

export default secrets;
