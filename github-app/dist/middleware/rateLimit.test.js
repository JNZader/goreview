/**
 * Tests for advanced rate limiting middleware
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { createRateLimiter, getClientIp, defaultKeyGenerator, tokenKeyGenerator, isIpBlocked, blockIp, recordFailedAttempt, getRateLimitInfo, resetRateLimit, clearAllRateLimits, RateLimitPresets, } from './rateLimit.js';
// Mock logger
vi.mock('../utils/logger.js', () => ({
    logger: {
        info: vi.fn(),
        warn: vi.fn(),
        error: vi.fn(),
    },
}));
describe('rateLimit', () => {
    // Helper to create mock request
    const createMockRequest = (overrides = {}) => {
        return {
            headers: {},
            path: '/api/test',
            method: 'GET',
            socket: { remoteAddress: '192.168.1.1' },
            ...overrides,
        };
    };
    // Helper to create mock response
    const createMockResponse = () => {
        const res = {
            status: vi.fn().mockReturnThis(),
            json: vi.fn().mockReturnThis(),
            setHeader: vi.fn().mockReturnThis(),
        };
        return res;
    };
    beforeEach(() => {
        clearAllRateLimits();
        vi.useFakeTimers();
    });
    afterEach(() => {
        vi.useRealTimers();
        vi.clearAllMocks();
    });
    // ===========================================================================
    // IP Extraction Tests
    // ===========================================================================
    describe('getClientIp', () => {
        it('should extract IP from X-Forwarded-For header', () => {
            const req = createMockRequest({
                headers: { 'x-forwarded-for': '10.0.0.1, 192.168.1.1' },
            });
            expect(getClientIp(req)).toBe('10.0.0.1');
        });
        it('should extract IP from X-Real-IP header', () => {
            const req = createMockRequest({
                headers: { 'x-real-ip': '10.0.0.2' },
            });
            expect(getClientIp(req)).toBe('10.0.0.2');
        });
        it('should fallback to socket remote address', () => {
            const req = createMockRequest();
            expect(getClientIp(req)).toBe('192.168.1.1');
        });
        it('should handle array headers', () => {
            const req = createMockRequest({
                headers: { 'x-forwarded-for': ['10.0.0.1', '192.168.1.1'] },
            });
            expect(getClientIp(req)).toBe('10.0.0.1');
        });
    });
    // ===========================================================================
    // Key Generator Tests
    // ===========================================================================
    describe('defaultKeyGenerator', () => {
        it('should generate key from IP and path', () => {
            const req = createMockRequest({ path: '/api/test' });
            const key = defaultKeyGenerator(req);
            expect(key).toBe('192.168.1.1:/api/test');
        });
    });
    describe('tokenKeyGenerator', () => {
        it('should generate key from Bearer token hash', () => {
            const req = createMockRequest({
                headers: { authorization: 'Bearer test-token-12345' },
            });
            const key = tokenKeyGenerator(req);
            expect(key).toMatch(/^token:[a-f0-9]{16}$/);
        });
        it('should fallback to IP-based key without token', () => {
            const req = createMockRequest();
            const key = tokenKeyGenerator(req);
            expect(key).toBe('ip:192.168.1.1');
        });
    });
    // ===========================================================================
    // Rate Limiter Tests
    // ===========================================================================
    describe('createRateLimiter', () => {
        it('should allow requests within limit', () => {
            const limiter = createRateLimiter({
                maxTokens: 10,
                refillRate: 1,
                tokensPerRequest: 1,
                windowMs: 60000,
            });
            const req = createMockRequest();
            const res = createMockResponse();
            const next = vi.fn();
            limiter(req, res, next);
            expect(next).toHaveBeenCalled();
            expect(res.status).not.toHaveBeenCalled();
        });
        it('should block requests exceeding limit', () => {
            const limiter = createRateLimiter({
                maxTokens: 2,
                refillRate: 0.001, // Very slow refill
                tokensPerRequest: 1,
                windowMs: 60000,
            });
            const req = createMockRequest();
            const res = createMockResponse();
            const next = vi.fn();
            // Make 3 requests (bucket starts full with 2 tokens)
            limiter(req, res, next);
            limiter(req, res, next);
            limiter(req, res, next);
            expect(res.status).toHaveBeenCalledWith(429);
            expect(res.json).toHaveBeenCalledWith(expect.objectContaining({
                error: 'Too Many Requests',
            }));
        });
        it('should set rate limit headers', () => {
            const limiter = createRateLimiter({
                maxTokens: 100,
                refillRate: 1,
                tokensPerRequest: 1,
                windowMs: 60000,
            });
            const req = createMockRequest();
            const res = createMockResponse();
            const next = vi.fn();
            limiter(req, res, next);
            expect(res.setHeader).toHaveBeenCalledWith('X-RateLimit-Limit', '100');
            expect(res.setHeader).toHaveBeenCalledWith('X-RateLimit-Remaining', expect.any(String));
            expect(res.setHeader).toHaveBeenCalledWith('X-RateLimit-Reset', expect.any(String));
        });
        it('should refill tokens over time', () => {
            const limiter = createRateLimiter({
                maxTokens: 2,
                refillRate: 1, // 1 token per second
                tokensPerRequest: 1,
                windowMs: 60000,
            });
            const req = createMockRequest();
            const res = createMockResponse();
            const next = vi.fn();
            // Exhaust tokens
            limiter(req, res, next);
            limiter(req, res, next);
            // Advance time to refill
            vi.advanceTimersByTime(2000);
            // Should be allowed again
            limiter(req, res, next);
            expect(next).toHaveBeenCalledTimes(3);
        });
        it('should skip rate limiting when skip function returns true', () => {
            const limiter = createRateLimiter({
                maxTokens: 0, // Would normally block everything
                refillRate: 0,
                tokensPerRequest: 1,
                windowMs: 60000,
                skip: (req) => req.path === '/health',
            });
            const req = createMockRequest({ path: '/health' });
            const res = createMockResponse();
            const next = vi.fn();
            limiter(req, res, next);
            expect(next).toHaveBeenCalled();
        });
        it('should use custom handler when provided', () => {
            const customHandler = vi.fn();
            const limiter = createRateLimiter({
                maxTokens: 0,
                refillRate: 0,
                tokensPerRequest: 1,
                windowMs: 60000,
                handler: customHandler,
            });
            const req = createMockRequest();
            const res = createMockResponse();
            const next = vi.fn();
            limiter(req, res, next);
            expect(customHandler).toHaveBeenCalledWith(req, res);
            expect(res.status).not.toHaveBeenCalled();
        });
    });
    // ===========================================================================
    // IP Blocking Tests
    // ===========================================================================
    describe('IP blocking', () => {
        it('should track blocked IPs', () => {
            expect(isIpBlocked('10.0.0.1')).toBe(false);
            blockIp('10.0.0.1', 60000);
            expect(isIpBlocked('10.0.0.1')).toBe(true);
        });
        it('should unblock IP after duration', () => {
            blockIp('10.0.0.2', 5000);
            expect(isIpBlocked('10.0.0.2')).toBe(true);
            vi.advanceTimersByTime(5001);
            expect(isIpBlocked('10.0.0.2')).toBe(false);
        });
    });
    describe('recordFailedAttempt', () => {
        it('should not block IP before threshold', () => {
            const ip = '10.0.0.3';
            recordFailedAttempt(ip, 5);
            recordFailedAttempt(ip, 5);
            recordFailedAttempt(ip, 5);
            recordFailedAttempt(ip, 5);
            expect(isIpBlocked(ip)).toBe(false);
        });
        it('should block IP after threshold', () => {
            const ip = '10.0.0.4';
            for (let i = 0; i < 5; i++) {
                recordFailedAttempt(ip, 5);
            }
            expect(isIpBlocked(ip)).toBe(true);
        });
        it('should reset attempts after window', () => {
            const ip = '10.0.0.5';
            recordFailedAttempt(ip, 5, 5000);
            recordFailedAttempt(ip, 5, 5000);
            vi.advanceTimersByTime(6000);
            // Should start fresh
            const blocked = recordFailedAttempt(ip, 5, 5000);
            expect(blocked).toBe(false);
            expect(isIpBlocked(ip)).toBe(false);
        });
    });
    // ===========================================================================
    // Utility Function Tests
    // ===========================================================================
    describe('getRateLimitInfo', () => {
        it('should return full bucket for new key', () => {
            const info = getRateLimitInfo('new-key', RateLimitPresets.standard);
            expect(info.limit).toBe(100);
            expect(info.remaining).toBe(100);
        });
        it('should reflect consumed tokens', () => {
            const limiter = createRateLimiter(RateLimitPresets.standard);
            const req = createMockRequest({ path: '/test-info' });
            const res = createMockResponse();
            const next = vi.fn();
            // Make a request to consume a token
            limiter(req, res, next);
            const info = getRateLimitInfo('192.168.1.1:/test-info', RateLimitPresets.standard);
            expect(info.remaining).toBeLessThan(100);
        });
    });
    describe('resetRateLimit', () => {
        it('should clear rate limit for key', () => {
            const limiter = createRateLimiter({
                maxTokens: 1,
                refillRate: 0.001,
                tokensPerRequest: 1,
                windowMs: 60000,
            });
            const req = createMockRequest({ path: '/reset-test' });
            const res = createMockResponse();
            const next = vi.fn();
            // Exhaust limit
            limiter(req, res, next);
            limiter(req, res, next);
            // Reset
            resetRateLimit('192.168.1.1:/reset-test');
            // Should be allowed again
            limiter(req, res, next);
            expect(next).toHaveBeenCalledTimes(2); // First request + after reset
        });
    });
    // ===========================================================================
    // Presets Tests
    // ===========================================================================
    describe('RateLimitPresets', () => {
        it('should have standard preset', () => {
            expect(RateLimitPresets.standard.maxTokens).toBe(100);
        });
        it('should have strict preset with lower limits', () => {
            expect(RateLimitPresets.strict.maxTokens).toBe(10);
        });
        it('should have webhook preset with higher limits', () => {
            expect(RateLimitPresets.webhook.maxTokens).toBe(1000);
        });
        it('should have admin preset', () => {
            expect(RateLimitPresets.admin.maxTokens).toBe(30);
        });
    });
});
//# sourceMappingURL=rateLimit.test.js.map