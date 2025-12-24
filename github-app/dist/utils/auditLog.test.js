/**
 * Tests for security audit logging system
 */
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { auditLog } from './auditLog.js';
// Mock dependencies
vi.mock('./logger.js', () => ({
    logger: {
        info: vi.fn(),
        warn: vi.fn(),
        error: vi.fn(),
    },
}));
vi.mock('../middleware/rateLimit.js', () => ({
    getClientIp: vi.fn().mockReturnValue('192.168.1.100'),
}));
vi.mock('./secrets.js', () => ({
    maskSecret: vi.fn().mockImplementation((s) => s ? '****' : s),
}));
import { logger } from './logger.js';
describe('auditLog', () => {
    // Helper to create mock request
    const createMockRequest = (overrides = {}) => {
        return {
            headers: {
                'x-request-id': 'req-12345',
                'user-agent': 'Test-Agent/1.0',
            },
            path: '/api/test',
            method: 'POST',
            socket: { remoteAddress: '192.168.1.100' },
            ...overrides,
        };
    };
    beforeEach(() => {
        vi.clearAllMocks();
    });
    // ===========================================================================
    // Core Logging Tests
    // ===========================================================================
    describe('log', () => {
        it('should log audit events with correct structure', () => {
            auditLog.log('api.request', 'Test API request', {
                method: 'GET',
                path: '/test',
            });
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                audit: true,
                event: 'api.request',
                severity: 'info',
                message: 'Test API request',
                context: expect.objectContaining({
                    method: 'GET',
                    path: '/test',
                }),
            }));
        });
        it('should use correct log level based on severity', () => {
            auditLog.log('security.suspicious_activity', 'Test critical', {}, 'critical');
            expect(logger.error).toHaveBeenCalled();
        });
        it('should log warnings with warn level', () => {
            auditLog.log('rate_limit.exceeded', 'Rate limit test', {}, 'warn');
            expect(logger.warn).toHaveBeenCalled();
        });
        it('should include timestamp in ISO format', () => {
            auditLog.log('api.request', 'Test', {});
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                timestamp: expect.stringMatching(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/),
            }));
        });
    });
    // ===========================================================================
    // Authentication Events Tests
    // ===========================================================================
    describe('authentication events', () => {
        it('should log successful authentication', () => {
            const req = createMockRequest();
            auditLog.authSuccess(req, 'user-123', 'user');
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'auth.login',
                context: expect.objectContaining({
                    actorId: 'user-123',
                    actorType: 'user',
                }),
            }));
        });
        it('should log failed authentication', () => {
            const req = createMockRequest();
            auditLog.authFailed(req, 'Invalid credentials', 'unknown-user');
            expect(logger.warn).toHaveBeenCalledWith(expect.objectContaining({
                event: 'auth.failed',
                context: expect.objectContaining({
                    metadata: expect.objectContaining({
                        reason: 'Invalid credentials',
                    }),
                }),
            }));
        });
        it('should log token refresh', () => {
            const req = createMockRequest();
            auditLog.tokenRefresh(req, 'app-456');
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'auth.token_refresh',
                context: expect.objectContaining({
                    actorId: 'app-456',
                }),
            }));
        });
    });
    // ===========================================================================
    // Webhook Events Tests
    // ===========================================================================
    describe('webhook events', () => {
        it('should log webhook received', () => {
            const req = createMockRequest();
            auditLog.webhookReceived(req, 'pull_request', 'delivery-123');
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'webhook.received',
                context: expect.objectContaining({
                    metadata: expect.objectContaining({
                        eventType: 'pull_request',
                        deliveryId: 'delivery-123',
                    }),
                }),
            }));
        });
        it('should log webhook rejected', () => {
            const req = createMockRequest();
            auditLog.webhookRejected(req, 'Invalid signature');
            expect(logger.warn).toHaveBeenCalledWith(expect.objectContaining({
                event: 'webhook.rejected',
            }));
        });
        it('should log successful webhook processing', () => {
            const req = createMockRequest();
            auditLog.webhookProcessed(req, 'push', true, 150);
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'webhook.processed',
                context: expect.objectContaining({
                    duration: 150,
                    metadata: expect.objectContaining({
                        success: true,
                    }),
                }),
            }));
        });
        it('should log failed webhook processing with warning', () => {
            const req = createMockRequest();
            auditLog.webhookProcessed(req, 'push', false, 100);
            expect(logger.warn).toHaveBeenCalledWith(expect.objectContaining({
                event: 'webhook.processed',
                context: expect.objectContaining({
                    metadata: expect.objectContaining({
                        success: false,
                    }),
                }),
            }));
        });
    });
    // ===========================================================================
    // Security Events Tests
    // ===========================================================================
    describe('security events', () => {
        it('should log rate limit exceeded', () => {
            const req = createMockRequest();
            auditLog.rateLimitExceeded(req, 100, 0);
            expect(logger.warn).toHaveBeenCalledWith(expect.objectContaining({
                event: 'rate_limit.exceeded',
                context: expect.objectContaining({
                    metadata: expect.objectContaining({
                        limit: 100,
                        remaining: 0,
                    }),
                }),
            }));
        });
        it('should log IP blocked', () => {
            auditLog.ipBlocked('10.0.0.1', 'Too many failed attempts', 3600000);
            expect(logger.warn).toHaveBeenCalledWith(expect.objectContaining({
                event: 'security.ip_blocked',
                context: expect.objectContaining({
                    ip: '10.0.0.1',
                    metadata: expect.objectContaining({
                        reason: 'Too many failed attempts',
                        durationMs: 3600000,
                    }),
                }),
            }));
        });
        it('should log suspicious activity with critical severity', () => {
            const req = createMockRequest();
            auditLog.suspiciousActivity(req, 'Multiple injection attempts', {
                pattern: 'SQL injection',
                count: 5,
            });
            expect(logger.error).toHaveBeenCalledWith(expect.objectContaining({
                event: 'security.suspicious_activity',
                severity: 'critical',
            }));
        });
        it('should log invalid signature with critical severity', () => {
            const req = createMockRequest();
            auditLog.signatureInvalid(req, 'pull_request');
            expect(logger.error).toHaveBeenCalledWith(expect.objectContaining({
                event: 'security.signature_invalid',
                severity: 'critical',
            }));
        });
    });
    // ===========================================================================
    // Admin Events Tests
    // ===========================================================================
    describe('admin events', () => {
        it('should log admin access', () => {
            const req = createMockRequest();
            auditLog.adminAccess(req, '/admin/stats', 'admin-user');
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'admin.access',
                context: expect.objectContaining({
                    actorId: 'admin-user',
                    metadata: expect.objectContaining({
                        endpoint: '/admin/stats',
                    }),
                }),
            }));
        });
        it('should log admin action', () => {
            const req = createMockRequest();
            auditLog.adminAction(req, 'cancel', 'job', 'job-123');
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'admin.action',
                context: expect.objectContaining({
                    resourceType: 'job',
                    resourceId: 'job-123',
                    metadata: expect.objectContaining({
                        action: 'cancel',
                    }),
                }),
            }));
        });
    });
    // ===========================================================================
    // Installation Events Tests
    // ===========================================================================
    describe('installation events', () => {
        it('should log installation created', () => {
            auditLog.installationCreated(12345, 'test-org');
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'installation.created',
                context: expect.objectContaining({
                    resourceType: 'installation',
                    resourceId: 12345,
                }),
            }));
        });
        it('should log installation deleted', () => {
            auditLog.installationDeleted(12345, 'test-org');
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'installation.deleted',
            }));
        });
        it('should log installation suspended with warning', () => {
            auditLog.installationSuspended(12345, 'test-org');
            expect(logger.warn).toHaveBeenCalledWith(expect.objectContaining({
                event: 'installation.suspended',
            }));
        });
    });
    // ===========================================================================
    // Review Events Tests
    // ===========================================================================
    describe('review events', () => {
        it('should log review started', () => {
            auditLog.reviewStarted('owner/repo', 42, 'job-123');
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'review.started',
                context: expect.objectContaining({
                    resourceType: 'pull_request',
                    resourceId: 42,
                }),
            }));
        });
        it('should log review completed', () => {
            auditLog.reviewCompleted('owner/repo', 42, 5000, 3);
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'review.completed',
                context: expect.objectContaining({
                    duration: 5000,
                    metadata: expect.objectContaining({
                        issuesFound: 3,
                    }),
                }),
            }));
        });
        it('should log review failed with error level', () => {
            auditLog.reviewFailed('owner/repo', 42, 'Timeout');
            expect(logger.error).toHaveBeenCalledWith(expect.objectContaining({
                event: 'review.failed',
            }));
        });
    });
    // ===========================================================================
    // Job Events Tests
    // ===========================================================================
    describe('job events', () => {
        it('should log job created', () => {
            auditLog.jobCreated('job-123', 'code_review', 1);
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'job.created',
                context: expect.objectContaining({
                    resourceType: 'job',
                    resourceId: 'job-123',
                }),
            }));
        });
        it('should log job processed', () => {
            auditLog.jobProcessed('job-123', 'code_review', 3000);
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'job.processed',
                context: expect.objectContaining({
                    duration: 3000,
                }),
            }));
        });
        it('should log job failed with error level', () => {
            auditLog.jobFailed('job-123', 'code_review', 'Network error', 3);
            expect(logger.error).toHaveBeenCalledWith(expect.objectContaining({
                event: 'job.failed',
                context: expect.objectContaining({
                    metadata: expect.objectContaining({
                        error: 'Network error',
                        attempts: 3,
                    }),
                }),
            }));
        });
        it('should log job cancelled with warning', () => {
            auditLog.jobCancelled('job-123', 'User request');
            expect(logger.warn).toHaveBeenCalledWith(expect.objectContaining({
                event: 'job.cancelled',
            }));
        });
    });
    // ===========================================================================
    // API Events Tests
    // ===========================================================================
    describe('API events', () => {
        it('should log successful API request', () => {
            const req = createMockRequest();
            auditLog.apiRequest(req, 200, 50);
            expect(logger.info).toHaveBeenCalledWith(expect.objectContaining({
                event: 'api.request',
                context: expect.objectContaining({
                    statusCode: 200,
                    duration: 50,
                }),
            }));
        });
        it('should log 4xx errors with warning', () => {
            const req = createMockRequest();
            auditLog.apiRequest(req, 404, 10);
            expect(logger.warn).toHaveBeenCalledWith(expect.objectContaining({
                event: 'api.request',
            }));
        });
        it('should log API error', () => {
            const req = createMockRequest();
            const error = new Error('Something went wrong');
            auditLog.apiError(req, error, 500);
            expect(logger.error).toHaveBeenCalledWith(expect.objectContaining({
                event: 'api.error',
                context: expect.objectContaining({
                    statusCode: 500,
                    metadata: expect.objectContaining({
                        errorName: 'Error',
                        errorMessage: 'Something went wrong',
                    }),
                }),
            }));
        });
        it('should log 4xx API error with warning', () => {
            const req = createMockRequest();
            const error = new Error('Not found');
            auditLog.apiError(req, error, 404);
            expect(logger.warn).toHaveBeenCalledWith(expect.objectContaining({
                event: 'api.error',
            }));
        });
    });
});
//# sourceMappingURL=auditLog.test.js.map