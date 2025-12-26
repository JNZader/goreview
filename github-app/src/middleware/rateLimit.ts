/**
 * Advanced rate limiting middleware
 * Implements token bucket algorithm with per-endpoint and per-client limits
 */

import type { Request, Response, NextFunction, RequestHandler } from 'express';
import crypto from 'node:crypto';
import { logger } from '../utils/logger.js';

// =============================================================================
// Types
// =============================================================================

interface TokenBucket {
  tokens: number;
  lastRefill: number;
}

interface RateLimitConfig {
  /** Maximum tokens in bucket */
  maxTokens: number;
  /** Tokens to refill per second */
  refillRate: number;
  /** Tokens consumed per request */
  tokensPerRequest: number;
  /** Time window for rate limit headers (ms) */
  windowMs: number;
  /** Custom key generator */
  keyGenerator?: (req: Request) => string;
  /** Skip rate limiting for certain requests */
  skip?: (req: Request) => boolean;
  /** Custom handler when rate limited */
  handler?: (req: Request, res: Response) => void;
}

interface RateLimitStore {
  get(key: string): TokenBucket | undefined;
  set(key: string, bucket: TokenBucket): void;
  delete(key: string): void;
  clear(): void;
}

export interface RateLimitInfo {
  limit: number;
  remaining: number;
  reset: number;
}

// =============================================================================
// In-Memory Store
// =============================================================================

class MemoryStore implements RateLimitStore {
  private readonly buckets: Map<string, TokenBucket> = new Map();
  private cleanupInterval: NodeJS.Timeout | null = null;

  constructor(cleanupIntervalMs: number = 60000) {
    // Periodic cleanup of stale buckets
    this.cleanupInterval = setInterval(() => {
      this.cleanup();
    }, cleanupIntervalMs);

    // Allow process to exit even if interval is running
    if (this.cleanupInterval.unref) {
      this.cleanupInterval.unref();
    }
  }

  get(key: string): TokenBucket | undefined {
    return this.buckets.get(key);
  }

  set(key: string, bucket: TokenBucket): void {
    this.buckets.set(key, bucket);
  }

  delete(key: string): void {
    this.buckets.delete(key);
  }

  clear(): void {
    this.buckets.clear();
  }

  private cleanup(): void {
    const now = Date.now();
    const staleThreshold = 5 * 60 * 1000; // 5 minutes

    for (const [key, bucket] of this.buckets.entries()) {
      if (now - bucket.lastRefill > staleThreshold) {
        this.buckets.delete(key);
      }
    }
  }

  destroy(): void {
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
      this.cleanupInterval = null;
    }
    this.clear();
  }

  get size(): number {
    return this.buckets.size;
  }
}

// =============================================================================
// Token Bucket Algorithm
// =============================================================================

function refillBucket(bucket: TokenBucket, config: RateLimitConfig): TokenBucket {
  const now = Date.now();
  const timePassed = (now - bucket.lastRefill) / 1000; // seconds
  const tokensToAdd = timePassed * config.refillRate;
  const newTokens = Math.min(config.maxTokens, bucket.tokens + tokensToAdd);

  return {
    tokens: newTokens,
    lastRefill: now,
  };
}

function consumeToken(bucket: TokenBucket, config: RateLimitConfig): { allowed: boolean; bucket: TokenBucket } {
  const refilled = refillBucket(bucket, config);

  if (refilled.tokens >= config.tokensPerRequest) {
    return {
      allowed: true,
      bucket: {
        tokens: refilled.tokens - config.tokensPerRequest,
        lastRefill: refilled.lastRefill,
      },
    };
  }

  return {
    allowed: false,
    bucket: refilled,
  };
}

// =============================================================================
// Default Key Generators
// =============================================================================

/**
 * Extracts client IP from request, considering proxies
 */
export function getClientIp(req: Request): string {
  // Check X-Forwarded-For header (set by proxies/load balancers)
  const forwarded = req.headers['x-forwarded-for'];
  if (forwarded) {
    const firstForwarded = Array.isArray(forwarded) ? forwarded[0] : forwarded.split(',')[0];
    if (firstForwarded) {
      return firstForwarded.trim();
    }
  }

  // Check X-Real-IP header (set by nginx)
  const realIp = req.headers['x-real-ip'];
  if (realIp) {
    const ip = Array.isArray(realIp) ? realIp[0] : realIp;
    if (ip) {
      return ip;
    }
  }

  // Fall back to socket remote address
  return req.socket?.remoteAddress ?? 'unknown';
}

/**
 * Default key generator: IP + path
 */
export function defaultKeyGenerator(req: Request): string {
  const ip = getClientIp(req);
  const path = req.path;
  return `${ip}:${path}`;
}

/**
 * Key generator for API tokens (from Authorization header)
 */
export function tokenKeyGenerator(req: Request): string {
  const auth = req.headers.authorization;
  if (auth?.startsWith('Bearer ')) {
    const token = auth.slice(7);
    // Hash the token to avoid storing sensitive data
    const hash = crypto.createHash('sha256').update(token).digest('hex').slice(0, 16);
    return `token:${hash}`;
  }
  // Fall back to IP-based limiting
  return `ip:${getClientIp(req)}`;
}

// =============================================================================
// Preset Configurations
// =============================================================================

export const RateLimitPresets = {
  /** Standard API endpoint: 100 requests per minute */
  standard: {
    maxTokens: 100,
    refillRate: 100 / 60, // ~1.67 tokens/sec
    tokensPerRequest: 1,
    windowMs: 60000,
  },

  /** Strict limit for sensitive endpoints: 10 requests per minute */
  strict: {
    maxTokens: 10,
    refillRate: 10 / 60, // ~0.17 tokens/sec
    tokensPerRequest: 1,
    windowMs: 60000,
  },

  /** Webhook endpoints: 1000 requests per minute */
  webhook: {
    maxTokens: 1000,
    refillRate: 1000 / 60, // ~16.67 tokens/sec
    tokensPerRequest: 1,
    windowMs: 60000,
  },

  /** Admin endpoints: 30 requests per minute */
  admin: {
    maxTokens: 30,
    refillRate: 30 / 60, // 0.5 tokens/sec
    tokensPerRequest: 1,
    windowMs: 60000,
  },

  /** Burst-friendly: high initial limit with slow refill */
  burst: {
    maxTokens: 50,
    refillRate: 10 / 60, // ~0.17 tokens/sec
    tokensPerRequest: 1,
    windowMs: 60000,
  },
} as const;

// =============================================================================
// Rate Limit Middleware Factory
// =============================================================================

const globalStore = new MemoryStore();

/**
 * Creates a rate limiting middleware
 */
export function createRateLimiter(config: Partial<RateLimitConfig> = {}): RequestHandler {
  const fullConfig: RateLimitConfig = {
    maxTokens: 100,
    refillRate: 100 / 60,
    tokensPerRequest: 1,
    windowMs: 60000,
    keyGenerator: defaultKeyGenerator,
    ...config,
  };

  const store = globalStore;

  return (req: Request, res: Response, next: NextFunction): void => {
    // Check if request should skip rate limiting
    if (fullConfig.skip?.(req)) {
      next();
      return;
    }

    const key = fullConfig.keyGenerator!(req);
    const bucket = store.get(key) ?? {
      tokens: fullConfig.maxTokens,
      lastRefill: Date.now(),
    };

    const result = consumeToken(bucket, fullConfig);
    store.set(key, result.bucket);

    // Calculate rate limit info for headers
    const resetTime = Math.ceil(
      (fullConfig.tokensPerRequest - result.bucket.tokens) / fullConfig.refillRate * 1000
    );
    const remaining = Math.max(0, Math.floor(result.bucket.tokens));

    // Set rate limit headers
    res.setHeader('X-RateLimit-Limit', fullConfig.maxTokens.toString());
    res.setHeader('X-RateLimit-Remaining', remaining.toString());
    res.setHeader('X-RateLimit-Reset', Math.ceil(Date.now() / 1000 + resetTime / 1000).toString());

    if (!result.allowed) {
      const retryAfter = Math.ceil(resetTime / 1000);
      res.setHeader('Retry-After', retryAfter.toString());

      logger.warn({
        event: 'rate_limit_exceeded',
        key,
        ip: getClientIp(req),
        path: req.path,
        method: req.method,
        retryAfter,
      });

      if (fullConfig.handler) {
        fullConfig.handler(req, res);
        return;
      }

      res.status(429).json({
        error: 'Too Many Requests',
        message: 'Rate limit exceeded. Please try again later.',
        retryAfter,
      });
      return;
    }

    next();
  };
}

// =============================================================================
// Specialized Middleware
// =============================================================================

/**
 * Rate limiter for webhook endpoints
 */
export const webhookRateLimiter = createRateLimiter({
  ...RateLimitPresets.webhook,
  keyGenerator: (req) => `webhook:${getClientIp(req)}`,
});

/**
 * Rate limiter for admin endpoints
 */
export const adminRateLimiter = createRateLimiter({
  ...RateLimitPresets.admin,
  keyGenerator: tokenKeyGenerator,
});

/**
 * Rate limiter for health check endpoints (very permissive)
 */
export const healthRateLimiter = createRateLimiter({
  maxTokens: 1000,
  refillRate: 1000 / 60,
  tokensPerRequest: 1,
  windowMs: 60000,
  keyGenerator: (req) => `health:${getClientIp(req)}`,
});

/**
 * Strict rate limiter for authentication endpoints
 */
export const authRateLimiter = createRateLimiter({
  ...RateLimitPresets.strict,
  keyGenerator: (req) => `auth:${getClientIp(req)}`,
});

// =============================================================================
// IP-Based Blocking
// =============================================================================

const blockedIps = new Set<string>();
const ipAttempts = new Map<string, { count: number; firstAttempt: number }>();

/**
 * Checks if an IP is blocked
 */
export function isIpBlocked(ip: string): boolean {
  return blockedIps.has(ip);
}

/**
 * Blocks an IP address
 */
export function blockIp(ip: string, durationMs: number = 3600000): void {
  blockedIps.add(ip);
  logger.warn({ event: 'ip_blocked', ip, durationMs });

  setTimeout(() => {
    blockedIps.delete(ip);
    logger.info({ event: 'ip_unblocked', ip });
  }, durationMs);
}

/**
 * Records a failed attempt and potentially blocks IP
 */
export function recordFailedAttempt(ip: string, maxAttempts: number = 5, windowMs: number = 900000): boolean {
  const now = Date.now();
  const attempts = ipAttempts.get(ip);

  if (!attempts || now - attempts.firstAttempt > windowMs) {
    ipAttempts.set(ip, { count: 1, firstAttempt: now });
    return false;
  }

  attempts.count++;

  if (attempts.count >= maxAttempts) {
    blockIp(ip);
    ipAttempts.delete(ip);
    return true;
  }

  return false;
}

/**
 * Middleware to block requests from blocked IPs
 */
export function ipBlockMiddleware(req: Request, res: Response, next: NextFunction): void {
  const ip = getClientIp(req);

  if (isIpBlocked(ip)) {
    logger.warn({
      event: 'blocked_ip_request',
      ip,
      path: req.path,
      method: req.method,
    });

    res.status(403).json({
      error: 'Forbidden',
      message: 'Your IP has been temporarily blocked due to suspicious activity.',
    });
    return;
  }

  next();
}

// =============================================================================
// Utility Functions
// =============================================================================

/**
 * Gets current rate limit info for a key
 */
export function getRateLimitInfo(key: string, config: RateLimitConfig = RateLimitPresets.standard): RateLimitInfo {
  const bucket = globalStore.get(key);

  if (!bucket) {
    return {
      limit: config.maxTokens,
      remaining: config.maxTokens,
      reset: Math.ceil(Date.now() / 1000),
    };
  }

  const refilled = refillBucket(bucket, config);
  const remaining = Math.max(0, Math.floor(refilled.tokens));
  const resetTime = Math.ceil(
    (config.tokensPerRequest - refilled.tokens) / config.refillRate * 1000
  );

  return {
    limit: config.maxTokens,
    remaining,
    reset: Math.ceil(Date.now() / 1000 + resetTime / 1000),
  };
}

/**
 * Resets rate limit for a specific key
 */
export function resetRateLimit(key: string): void {
  globalStore.delete(key);
}

/**
 * Clears all rate limit data
 */
export function clearAllRateLimits(): void {
  globalStore.clear();
}

/**
 * Gets the number of tracked clients
 */
export function getTrackedClientCount(): number {
  return globalStore.size;
}

// =============================================================================
// Exports
// =============================================================================

export const rateLimit = {
  createRateLimiter,
  webhookRateLimiter,
  adminRateLimiter,
  healthRateLimiter,
  authRateLimiter,
  ipBlockMiddleware,
  getClientIp,
  defaultKeyGenerator,
  tokenKeyGenerator,
  isIpBlocked,
  blockIp,
  recordFailedAttempt,
  getRateLimitInfo,
  resetRateLimit,
  clearAllRateLimits,
  getTrackedClientCount,
  RateLimitPresets,
};

export default rateLimit;
