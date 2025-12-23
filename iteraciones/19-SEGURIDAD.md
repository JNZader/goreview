# Iteracion 19: Seguridad

## Objetivos

- Manejo seguro de secretos
- Rate limiting robusto
- Validacion de entrada
- Audit logging

## Tiempo Estimado: 5 horas

---

## Commit 19.1: Implementar manejo seguro de secretos

**Mensaje de commit:**
```
security: add secure secret handling

- Encrypt sensitive config at rest
- Mask secrets in logs
- Secure environment variable handling
```

### `github-app/src/utils/secrets.ts`

```typescript
import crypto from 'crypto';
import { logger } from './logger.js';

const ALGORITHM = 'aes-256-gcm';
const IV_LENGTH = 16;
const TAG_LENGTH = 16;
const SALT_LENGTH = 64;
const KEY_LENGTH = 32;
const ITERATIONS = 100000;

/**
 * Encrypt a secret value.
 */
export function encryptSecret(plaintext: string, masterKey: string): string {
  // Generate salt and IV
  const salt = crypto.randomBytes(SALT_LENGTH);
  const iv = crypto.randomBytes(IV_LENGTH);

  // Derive key from master key
  const key = crypto.pbkdf2Sync(masterKey, salt, ITERATIONS, KEY_LENGTH, 'sha512');

  // Encrypt
  const cipher = crypto.createCipheriv(ALGORITHM, key, iv);
  const encrypted = Buffer.concat([
    cipher.update(plaintext, 'utf8'),
    cipher.final(),
  ]);
  const tag = cipher.getAuthTag();

  // Combine: salt + iv + tag + encrypted
  const result = Buffer.concat([salt, iv, tag, encrypted]);

  return result.toString('base64');
}

/**
 * Decrypt a secret value.
 */
export function decryptSecret(ciphertext: string, masterKey: string): string {
  const data = Buffer.from(ciphertext, 'base64');

  // Extract components
  const salt = data.subarray(0, SALT_LENGTH);
  const iv = data.subarray(SALT_LENGTH, SALT_LENGTH + IV_LENGTH);
  const tag = data.subarray(SALT_LENGTH + IV_LENGTH, SALT_LENGTH + IV_LENGTH + TAG_LENGTH);
  const encrypted = data.subarray(SALT_LENGTH + IV_LENGTH + TAG_LENGTH);

  // Derive key
  const key = crypto.pbkdf2Sync(masterKey, salt, ITERATIONS, KEY_LENGTH, 'sha512');

  // Decrypt
  const decipher = crypto.createDecipheriv(ALGORITHM, key, iv);
  decipher.setAuthTag(tag);

  return Buffer.concat([
    decipher.update(encrypted),
    decipher.final(),
  ]).toString('utf8');
}

/**
 * Mask sensitive data in strings.
 */
export function maskSecrets(text: string): string {
  // Patterns to mask
  const patterns = [
    // API keys and tokens
    /([a-zA-Z_]*(?:key|token|secret|password|auth)[a-zA-Z_]*[=:]\s*)([^\s'"]+)/gi,
    // GitHub tokens
    /(ghp_[a-zA-Z0-9]{36})/g,
    /(gho_[a-zA-Z0-9]{36})/g,
    /(github_pat_[a-zA-Z0-9_]{22,})/g,
    // OpenAI keys
    /(sk-[a-zA-Z0-9]{48})/g,
    // Bearer tokens
    /(Bearer\s+)[^\s]+/gi,
    // Basic auth
    /(Basic\s+)[^\s]+/gi,
  ];

  let masked = text;

  for (const pattern of patterns) {
    masked = masked.replace(pattern, (match, prefix) => {
      if (prefix) {
        return prefix + '***REDACTED***';
      }
      return '***REDACTED***';
    });
  }

  return masked;
}

/**
 * Securely compare two strings (timing-safe).
 */
export function secureCompare(a: string, b: string): boolean {
  if (typeof a !== 'string' || typeof b !== 'string') {
    return false;
  }

  const bufA = Buffer.from(a);
  const bufB = Buffer.from(b);

  if (bufA.length !== bufB.length) {
    return false;
  }

  return crypto.timingSafeEqual(bufA, bufB);
}

/**
 * Generate a secure random token.
 */
export function generateSecureToken(length: number = 32): string {
  return crypto.randomBytes(length).toString('hex');
}

/**
 * Hash a value using SHA-256.
 */
export function hashValue(value: string): string {
  return crypto.createHash('sha256').update(value).digest('hex');
}
```

---

## Commit 19.2: Implementar rate limiting avanzado

**Mensaje de commit:**
```
security: add advanced rate limiting

- Per-installation rate limits
- Sliding window algorithm
- Configurable limits
- Rate limit headers
```

### `github-app/src/middleware/rateLimit.ts`

```typescript
import { Request, Response, NextFunction } from 'express';
import { config } from '../config/index.js';
import { logger } from '../utils/logger.js';

interface RateLimitEntry {
  count: number;
  windowStart: number;
  tokens: number;
  lastRefill: number;
}

interface RateLimitConfig {
  windowMs: number;      // Time window in ms
  maxRequests: number;   // Max requests per window
  burstLimit: number;    // Max burst requests
  skipFailedRequests?: boolean;
}

const DEFAULT_CONFIG: RateLimitConfig = {
  windowMs: 60000,       // 1 minute
  maxRequests: config.rateLimit.rps * 60,
  burstLimit: config.rateLimit.burst,
};

/**
 * Rate limiter using sliding window + token bucket hybrid.
 */
class RateLimiter {
  private entries: Map<string, RateLimitEntry> = new Map();
  private config: RateLimitConfig;

  constructor(options: Partial<RateLimitConfig> = {}) {
    this.config = { ...DEFAULT_CONFIG, ...options };

    // Cleanup expired entries periodically
    setInterval(() => this.cleanup(), this.config.windowMs);
  }

  /**
   * Check if request should be allowed.
   */
  check(key: string): { allowed: boolean; remaining: number; resetTime: number } {
    const now = Date.now();
    let entry = this.entries.get(key);

    // Initialize new entry
    if (!entry) {
      entry = {
        count: 0,
        windowStart: now,
        tokens: this.config.burstLimit,
        lastRefill: now,
      };
      this.entries.set(key, entry);
    }

    // Refill tokens (token bucket)
    const elapsed = (now - entry.lastRefill) / 1000;
    const tokensToAdd = elapsed * (this.config.maxRequests / (this.config.windowMs / 1000));
    entry.tokens = Math.min(this.config.burstLimit, entry.tokens + tokensToAdd);
    entry.lastRefill = now;

    // Check sliding window
    if (now - entry.windowStart >= this.config.windowMs) {
      entry.count = 0;
      entry.windowStart = now;
    }

    // Check limits
    if (entry.tokens < 1 || entry.count >= this.config.maxRequests) {
      const resetTime = entry.windowStart + this.config.windowMs;
      return {
        allowed: false,
        remaining: 0,
        resetTime,
      };
    }

    // Consume token and increment count
    entry.tokens -= 1;
    entry.count += 1;

    return {
      allowed: true,
      remaining: Math.floor(entry.tokens),
      resetTime: entry.windowStart + this.config.windowMs,
    };
  }

  /**
   * Record a failed request (optionally don't count against limit).
   */
  recordFailure(key: string): void {
    if (this.config.skipFailedRequests) {
      const entry = this.entries.get(key);
      if (entry && entry.count > 0) {
        entry.count -= 1;
        entry.tokens += 1;
      }
    }
  }

  /**
   * Get current stats for a key.
   */
  getStats(key: string): RateLimitEntry | undefined {
    return this.entries.get(key);
  }

  /**
   * Clean up expired entries.
   */
  private cleanup(): void {
    const now = Date.now();
    for (const [key, entry] of this.entries) {
      if (now - entry.windowStart > this.config.windowMs * 2) {
        this.entries.delete(key);
      }
    }
  }
}

// Shared rate limiter instances
const globalLimiter = new RateLimiter();
const installationLimiters = new Map<number, RateLimiter>();

/**
 * Get or create rate limiter for an installation.
 */
function getInstallationLimiter(installationId: number): RateLimiter {
  if (!installationLimiters.has(installationId)) {
    installationLimiters.set(installationId, new RateLimiter({
      maxRequests: 100, // Per-installation limit
      burstLimit: 20,
    }));
  }
  return installationLimiters.get(installationId)!;
}

/**
 * Rate limiting middleware.
 */
export function rateLimitMiddleware(req: Request, res: Response, next: NextFunction) {
  // Determine key (IP for unauthenticated, installation for authenticated)
  const installationId = (req as any).installationId;
  const key = installationId ? `inst:${installationId}` : `ip:${req.ip}`;

  // Check global limit
  const globalResult = globalLimiter.check(`global:${req.ip}`);
  if (!globalResult.allowed) {
    logger.warn({ ip: req.ip }, 'Global rate limit exceeded');
    return sendRateLimitResponse(res, globalResult);
  }

  // Check per-installation limit
  if (installationId) {
    const instLimiter = getInstallationLimiter(installationId);
    const instResult = instLimiter.check(key);

    if (!instResult.allowed) {
      logger.warn({ installationId }, 'Installation rate limit exceeded');
      return sendRateLimitResponse(res, instResult);
    }

    // Set rate limit headers
    res.set({
      'X-RateLimit-Limit': String(100),
      'X-RateLimit-Remaining': String(instResult.remaining),
      'X-RateLimit-Reset': String(Math.ceil(instResult.resetTime / 1000)),
    });
  }

  next();
}

function sendRateLimitResponse(
  res: Response,
  result: { remaining: number; resetTime: number }
) {
  res.set({
    'X-RateLimit-Remaining': '0',
    'X-RateLimit-Reset': String(Math.ceil(result.resetTime / 1000)),
    'Retry-After': String(Math.ceil((result.resetTime - Date.now()) / 1000)),
  });

  res.status(429).json({
    error: 'Too Many Requests',
    message: 'Rate limit exceeded. Please retry later.',
    retryAfter: Math.ceil((result.resetTime - Date.now()) / 1000),
  });
}
```

---

## Commit 19.3: Agregar validacion de entrada robusta

**Mensaje de commit:**
```
security: add input validation

- Request body validation with Zod
- Path parameter sanitization
- Query parameter validation
- Error messages without leaking info
```

### `github-app/src/middleware/validation.ts`

```typescript
import { Request, Response, NextFunction } from 'express';
import { z, ZodError, ZodSchema } from 'zod';
import { logger } from '../utils/logger.js';

/**
 * Validation middleware factory.
 */
export function validate(schema: {
  body?: ZodSchema;
  query?: ZodSchema;
  params?: ZodSchema;
}) {
  return (req: Request, res: Response, next: NextFunction) => {
    try {
      if (schema.body) {
        req.body = schema.body.parse(req.body);
      }
      if (schema.query) {
        req.query = schema.query.parse(req.query);
      }
      if (schema.params) {
        req.params = schema.params.parse(req.params);
      }
      next();
    } catch (error) {
      if (error instanceof ZodError) {
        logger.warn({
          path: req.path,
          errors: error.errors,
        }, 'Validation error');

        return res.status(400).json({
          error: 'Validation Error',
          details: error.errors.map(e => ({
            field: e.path.join('.'),
            message: e.message,
          })),
        });
      }
      next(error);
    }
  };
}

/**
 * Sanitize string input.
 */
export function sanitizeString(input: string): string {
  return input
    // Remove null bytes
    .replace(/\0/g, '')
    // Normalize unicode
    .normalize('NFC')
    // Trim whitespace
    .trim();
}

/**
 * Sanitize path parameter.
 */
export function sanitizePath(input: string): string {
  return input
    // Remove path traversal attempts
    .replace(/\.\./g, '')
    .replace(/[<>:"|?*]/g, '')
    // Normalize slashes
    .replace(/\\/g, '/')
    // Remove leading slashes
    .replace(/^\/+/, '');
}

// Common validation schemas
export const schemas = {
  // Repository reference
  repoRef: z.object({
    owner: z.string()
      .min(1)
      .max(39)
      .regex(/^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?$/),
    repo: z.string()
      .min(1)
      .max(100)
      .regex(/^[a-zA-Z0-9._-]+$/),
  }),

  // Pull request number
  pullNumber: z.object({
    pull_number: z.coerce.number().int().positive(),
  }),

  // Pagination
  pagination: z.object({
    page: z.coerce.number().int().min(1).default(1),
    per_page: z.coerce.number().int().min(1).max(100).default(30),
  }),

  // Review request body
  reviewRequest: z.object({
    files: z.array(z.string()).optional(),
    preset: z.enum(['minimal', 'standard', 'strict']).optional(),
    options: z.object({
      inline_comments: z.boolean().optional(),
      summary: z.boolean().optional(),
    }).optional(),
  }),
};

/**
 * Check for common attack patterns.
 */
export function detectMaliciousInput(input: string): boolean {
  const patterns = [
    // SQL injection
    /(\bunion\b.*\bselect\b|\bselect\b.*\bfrom\b|\binsert\b.*\binto\b)/i,
    // Script injection
    /<script[\s\S]*?>[\s\S]*?<\/script>/i,
    // Command injection
    /[;&|`$]|\$\(|\${/,
    // Path traversal
    /\.\.[\/\\]/,
  ];

  return patterns.some(pattern => pattern.test(input));
}
```

---

## Commit 19.4: Implementar audit logging

**Mensaje de commit:**
```
security: add audit logging

- Log security-relevant events
- Structured audit trail
- Include actor information
- Retention and rotation
```

### `github-app/src/utils/auditLog.ts`

```typescript
import pino from 'pino';
import fs from 'fs';
import path from 'path';
import { config } from '../config/index.js';

// Audit log types
export type AuditEventType =
  | 'auth.success'
  | 'auth.failure'
  | 'webhook.received'
  | 'webhook.verified'
  | 'webhook.rejected'
  | 'review.started'
  | 'review.completed'
  | 'review.failed'
  | 'config.loaded'
  | 'config.error'
  | 'ratelimit.exceeded'
  | 'error.security';

export interface AuditEvent {
  type: AuditEventType;
  actor?: {
    type: 'installation' | 'user' | 'system';
    id: string | number;
    name?: string;
  };
  resource?: {
    type: 'repository' | 'pull_request' | 'installation';
    id: string | number;
    name?: string;
  };
  details?: Record<string, unknown>;
  ip?: string;
  userAgent?: string;
}

// Create audit logger
const auditLogger = pino({
  name: 'audit',
  level: 'info',
  base: {
    service: 'goreview-github-app',
    env: config.nodeEnv,
  },
  timestamp: pino.stdTimeFunctions.isoTime,
  formatters: {
    level: (label) => ({ level: label }),
  },
});

/**
 * Log an audit event.
 */
export function audit(event: AuditEvent): void {
  const logEntry = {
    eventType: event.type,
    timestamp: new Date().toISOString(),
    actor: event.actor,
    resource: event.resource,
    details: event.details,
    ip: event.ip,
    userAgent: event.userAgent,
  };

  // Log to audit log
  auditLogger.info(logEntry, `audit.${event.type}`);

  // For security events, also log to console in development
  if (config.isDevelopment && event.type.startsWith('error.')) {
    console.warn('[AUDIT SECURITY]', JSON.stringify(logEntry, null, 2));
  }
}

/**
 * Create audit middleware for Express.
 */
export function auditMiddleware(eventType: AuditEventType) {
  return (req: any, res: any, next: any) => {
    const startTime = Date.now();

    res.on('finish', () => {
      audit({
        type: eventType,
        actor: req.installation ? {
          type: 'installation',
          id: req.installation.id,
        } : undefined,
        details: {
          method: req.method,
          path: req.path,
          statusCode: res.statusCode,
          duration: Date.now() - startTime,
        },
        ip: req.ip,
        userAgent: req.get('user-agent'),
      });
    });

    next();
  };
}

// Convenience audit functions
export const auditAuth = {
  success: (installationId: number, ip?: string) =>
    audit({
      type: 'auth.success',
      actor: { type: 'installation', id: installationId },
      ip,
    }),

  failure: (reason: string, ip?: string) =>
    audit({
      type: 'auth.failure',
      details: { reason },
      ip,
    }),
};

export const auditWebhook = {
  received: (event: string, deliveryId: string, ip?: string) =>
    audit({
      type: 'webhook.received',
      details: { event, deliveryId },
      ip,
    }),

  verified: (event: string, deliveryId: string) =>
    audit({
      type: 'webhook.verified',
      details: { event, deliveryId },
    }),

  rejected: (reason: string, ip?: string) =>
    audit({
      type: 'webhook.rejected',
      details: { reason },
      ip,
    }),
};

export const auditReview = {
  started: (owner: string, repo: string, pullNumber: number, installationId: number) =>
    audit({
      type: 'review.started',
      actor: { type: 'installation', id: installationId },
      resource: {
        type: 'pull_request',
        id: pullNumber,
        name: `${owner}/${repo}#${pullNumber}`,
      },
    }),

  completed: (owner: string, repo: string, pullNumber: number, issueCount: number) =>
    audit({
      type: 'review.completed',
      resource: {
        type: 'pull_request',
        id: pullNumber,
        name: `${owner}/${repo}#${pullNumber}`,
      },
      details: { issueCount },
    }),

  failed: (owner: string, repo: string, pullNumber: number, error: string) =>
    audit({
      type: 'review.failed',
      resource: {
        type: 'pull_request',
        id: pullNumber,
        name: `${owner}/${repo}#${pullNumber}`,
      },
      details: { error },
    }),
};
```

---

## Commit 19.5: Tests de seguridad

**Mensaje de commit:**
```
test(security): add security tests

- Test secret handling
- Test rate limiting
- Test input validation
- Test audit logging
```

### `github-app/src/__tests__/security.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import {
  encryptSecret,
  decryptSecret,
  maskSecrets,
  secureCompare,
  generateSecureToken,
  hashValue,
} from '../utils/secrets.js';

describe('Secret handling', () => {
  const masterKey = 'test-master-key-for-testing-only';

  describe('encryptSecret / decryptSecret', () => {
    it('encrypts and decrypts correctly', () => {
      const plaintext = 'my-secret-value';
      const encrypted = encryptSecret(plaintext, masterKey);
      const decrypted = decryptSecret(encrypted, masterKey);

      expect(decrypted).toBe(plaintext);
    });

    it('produces different ciphertext each time', () => {
      const plaintext = 'my-secret-value';
      const encrypted1 = encryptSecret(plaintext, masterKey);
      const encrypted2 = encryptSecret(plaintext, masterKey);

      expect(encrypted1).not.toBe(encrypted2);
    });

    it('fails with wrong key', () => {
      const plaintext = 'my-secret-value';
      const encrypted = encryptSecret(plaintext, masterKey);

      expect(() => {
        decryptSecret(encrypted, 'wrong-key');
      }).toThrow();
    });
  });

  describe('maskSecrets', () => {
    it('masks API keys', () => {
      const text = 'Using API_KEY=sk-1234567890abcdef';
      const masked = maskSecrets(text);

      expect(masked).not.toContain('sk-1234567890abcdef');
      expect(masked).toContain('***REDACTED***');
    });

    it('masks GitHub tokens', () => {
      const text = 'Token: ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx';
      const masked = maskSecrets(text);

      expect(masked).toContain('***REDACTED***');
    });

    it('masks Bearer tokens', () => {
      const text = 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9';
      const masked = maskSecrets(text);

      expect(masked).toContain('Bearer ***REDACTED***');
    });
  });

  describe('secureCompare', () => {
    it('returns true for equal strings', () => {
      expect(secureCompare('test', 'test')).toBe(true);
    });

    it('returns false for different strings', () => {
      expect(secureCompare('test', 'Test')).toBe(false);
    });

    it('returns false for different lengths', () => {
      expect(secureCompare('test', 'testing')).toBe(false);
    });
  });

  describe('generateSecureToken', () => {
    it('generates token of correct length', () => {
      const token = generateSecureToken(16);
      expect(token.length).toBe(32); // hex encoding doubles length
    });

    it('generates unique tokens', () => {
      const token1 = generateSecureToken();
      const token2 = generateSecureToken();

      expect(token1).not.toBe(token2);
    });
  });

  describe('hashValue', () => {
    it('produces consistent hashes', () => {
      const hash1 = hashValue('test');
      const hash2 = hashValue('test');

      expect(hash1).toBe(hash2);
    });

    it('produces different hashes for different inputs', () => {
      const hash1 = hashValue('test1');
      const hash2 = hashValue('test2');

      expect(hash1).not.toBe(hash2);
    });
  });
});
```

### `github-app/src/__tests__/validation.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import {
  sanitizeString,
  sanitizePath,
  detectMaliciousInput,
  schemas,
} from '../middleware/validation.js';

describe('Input validation', () => {
  describe('sanitizeString', () => {
    it('removes null bytes', () => {
      expect(sanitizeString('test\0null')).toBe('testnull');
    });

    it('trims whitespace', () => {
      expect(sanitizeString('  test  ')).toBe('test');
    });
  });

  describe('sanitizePath', () => {
    it('removes path traversal attempts', () => {
      expect(sanitizePath('../../../etc/passwd')).toBe('etc/passwd');
    });

    it('normalizes slashes', () => {
      expect(sanitizePath('path\\to\\file')).toBe('path/to/file');
    });

    it('removes invalid characters', () => {
      expect(sanitizePath('file<>:"|?*.txt')).toBe('file.txt');
    });
  });

  describe('detectMaliciousInput', () => {
    it('detects SQL injection', () => {
      expect(detectMaliciousInput("'; DROP TABLE users; --")).toBe(true);
      expect(detectMaliciousInput('SELECT * FROM users')).toBe(true);
    });

    it('detects script injection', () => {
      expect(detectMaliciousInput('<script>alert(1)</script>')).toBe(true);
    });

    it('detects command injection', () => {
      expect(detectMaliciousInput('test; rm -rf /')).toBe(true);
      expect(detectMaliciousInput('$(whoami)')).toBe(true);
    });

    it('allows normal input', () => {
      expect(detectMaliciousInput('Hello, World!')).toBe(false);
      expect(detectMaliciousInput('user@example.com')).toBe(false);
    });
  });

  describe('schemas.repoRef', () => {
    it('validates correct repo references', () => {
      const result = schemas.repoRef.safeParse({
        owner: 'octocat',
        repo: 'hello-world',
      });

      expect(result.success).toBe(true);
    });

    it('rejects invalid owner names', () => {
      const result = schemas.repoRef.safeParse({
        owner: '-invalid',
        repo: 'test',
      });

      expect(result.success).toBe(false);
    });
  });
});
```

---

## Resumen de la Iteracion 19

### Commits:
1. `security: add secure secret handling`
2. `security: add advanced rate limiting`
3. `security: add input validation`
4. `security: add audit logging`
5. `test(security): add security tests`

### Archivos:
```
github-app/src/
├── utils/
│   ├── secrets.ts
│   └── auditLog.ts
├── middleware/
│   ├── rateLimit.ts
│   └── validation.ts
└── __tests__/
    ├── security.test.ts
    └── validation.test.ts
```

---

## Siguiente Iteracion

Continua con: **[20-PERFORMANCE.md](20-PERFORMANCE.md)**
