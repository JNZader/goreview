/**
 * Advanced rate limiting middleware
 * Implements token bucket algorithm with per-endpoint and per-client limits
 */
import type { Request, Response, NextFunction, RequestHandler } from 'express';
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
export interface RateLimitInfo {
    limit: number;
    remaining: number;
    reset: number;
}
/**
 * Extracts client IP from request, considering proxies
 */
export declare function getClientIp(req: Request): string;
/**
 * Default key generator: IP + path
 */
export declare function defaultKeyGenerator(req: Request): string;
/**
 * Key generator for API tokens (from Authorization header)
 */
export declare function tokenKeyGenerator(req: Request): string;
export declare const RateLimitPresets: {
    /** Standard API endpoint: 100 requests per minute */
    readonly standard: {
        readonly maxTokens: 100;
        readonly refillRate: number;
        readonly tokensPerRequest: 1;
        readonly windowMs: 60000;
    };
    /** Strict limit for sensitive endpoints: 10 requests per minute */
    readonly strict: {
        readonly maxTokens: 10;
        readonly refillRate: number;
        readonly tokensPerRequest: 1;
        readonly windowMs: 60000;
    };
    /** Webhook endpoints: 1000 requests per minute */
    readonly webhook: {
        readonly maxTokens: 1000;
        readonly refillRate: number;
        readonly tokensPerRequest: 1;
        readonly windowMs: 60000;
    };
    /** Admin endpoints: 30 requests per minute */
    readonly admin: {
        readonly maxTokens: 30;
        readonly refillRate: number;
        readonly tokensPerRequest: 1;
        readonly windowMs: 60000;
    };
    /** Burst-friendly: high initial limit with slow refill */
    readonly burst: {
        readonly maxTokens: 50;
        readonly refillRate: number;
        readonly tokensPerRequest: 1;
        readonly windowMs: 60000;
    };
};
/**
 * Creates a rate limiting middleware
 */
export declare function createRateLimiter(config?: Partial<RateLimitConfig>): RequestHandler;
/**
 * Rate limiter for webhook endpoints
 */
export declare const webhookRateLimiter: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
/**
 * Rate limiter for admin endpoints
 */
export declare const adminRateLimiter: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
/**
 * Rate limiter for health check endpoints (very permissive)
 */
export declare const healthRateLimiter: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
/**
 * Strict rate limiter for authentication endpoints
 */
export declare const authRateLimiter: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
/**
 * Checks if an IP is blocked
 */
export declare function isIpBlocked(ip: string): boolean;
/**
 * Blocks an IP address
 */
export declare function blockIp(ip: string, durationMs?: number): void;
/**
 * Records a failed attempt and potentially blocks IP
 */
export declare function recordFailedAttempt(ip: string, maxAttempts?: number, windowMs?: number): boolean;
/**
 * Middleware to block requests from blocked IPs
 */
export declare function ipBlockMiddleware(req: Request, res: Response, next: NextFunction): void;
/**
 * Gets current rate limit info for a key
 */
export declare function getRateLimitInfo(key: string, config?: RateLimitConfig): RateLimitInfo;
/**
 * Resets rate limit for a specific key
 */
export declare function resetRateLimit(key: string): void;
/**
 * Clears all rate limit data
 */
export declare function clearAllRateLimits(): void;
/**
 * Gets the number of tracked clients
 */
export declare function getTrackedClientCount(): number;
export declare const rateLimit: {
    createRateLimiter: typeof createRateLimiter;
    webhookRateLimiter: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    adminRateLimiter: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    healthRateLimiter: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    authRateLimiter: RequestHandler<import("express-serve-static-core").ParamsDictionary, any, any, import("qs").ParsedQs, Record<string, any>>;
    ipBlockMiddleware: typeof ipBlockMiddleware;
    getClientIp: typeof getClientIp;
    defaultKeyGenerator: typeof defaultKeyGenerator;
    tokenKeyGenerator: typeof tokenKeyGenerator;
    isIpBlocked: typeof isIpBlocked;
    blockIp: typeof blockIp;
    recordFailedAttempt: typeof recordFailedAttempt;
    getRateLimitInfo: typeof getRateLimitInfo;
    resetRateLimit: typeof resetRateLimit;
    clearAllRateLimits: typeof clearAllRateLimits;
    getTrackedClientCount: typeof getTrackedClientCount;
    RateLimitPresets: {
        /** Standard API endpoint: 100 requests per minute */
        readonly standard: {
            readonly maxTokens: 100;
            readonly refillRate: number;
            readonly tokensPerRequest: 1;
            readonly windowMs: 60000;
        };
        /** Strict limit for sensitive endpoints: 10 requests per minute */
        readonly strict: {
            readonly maxTokens: 10;
            readonly refillRate: number;
            readonly tokensPerRequest: 1;
            readonly windowMs: 60000;
        };
        /** Webhook endpoints: 1000 requests per minute */
        readonly webhook: {
            readonly maxTokens: 1000;
            readonly refillRate: number;
            readonly tokensPerRequest: 1;
            readonly windowMs: 60000;
        };
        /** Admin endpoints: 30 requests per minute */
        readonly admin: {
            readonly maxTokens: 30;
            readonly refillRate: number;
            readonly tokensPerRequest: 1;
            readonly windowMs: 60000;
        };
        /** Burst-friendly: high initial limit with slow refill */
        readonly burst: {
            readonly maxTokens: 50;
            readonly refillRate: number;
            readonly tokensPerRequest: 1;
            readonly windowMs: 60000;
        };
    };
};
export default rateLimit;
//# sourceMappingURL=rateLimit.d.ts.map