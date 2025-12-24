/**
 * Security audit logging system
 * Provides structured logging for security-relevant events
 */
import type { Request } from 'express';
export type AuditEventType = 'auth.login' | 'auth.logout' | 'auth.failed' | 'auth.token_refresh' | 'webhook.received' | 'webhook.verified' | 'webhook.rejected' | 'webhook.processed' | 'api.request' | 'api.error' | 'rate_limit.exceeded' | 'rate_limit.blocked' | 'security.ip_blocked' | 'security.suspicious_activity' | 'security.signature_invalid' | 'admin.access' | 'admin.action' | 'installation.created' | 'installation.deleted' | 'installation.suspended' | 'review.started' | 'review.completed' | 'review.failed' | 'job.created' | 'job.processed' | 'job.failed' | 'job.cancelled';
export type AuditSeverity = 'info' | 'warn' | 'error' | 'critical';
export interface AuditContext {
    /** Request ID for correlation */
    requestId?: string;
    /** Client IP address */
    ip?: string;
    /** User agent string */
    userAgent?: string;
    /** Authenticated user/app ID */
    actorId?: string | number;
    /** Actor type (user, app, system) */
    actorType?: 'user' | 'app' | 'system';
    /** Target resource type */
    resourceType?: string;
    /** Target resource ID */
    resourceId?: string | number;
    /** HTTP method */
    method?: string;
    /** Request path */
    path?: string;
    /** Response status code */
    statusCode?: number;
    /** Request duration in ms */
    duration?: number;
    /** Additional metadata */
    metadata?: Record<string, unknown>;
}
export interface AuditEntry {
    /** Event timestamp */
    timestamp: string;
    /** Event type */
    event: AuditEventType;
    /** Event severity */
    severity: AuditSeverity;
    /** Human-readable message */
    message: string;
    /** Event context */
    context: AuditContext;
}
declare class AuditLogger {
    private readonly serviceName;
    constructor(serviceName?: string);
    /**
     * Logs an audit event
     */
    log(event: AuditEventType, message: string, context?: AuditContext, severity?: AuditSeverity): void;
    /**
     * Extracts audit context from an Express request
     */
    extractContext(req: Request): AuditContext;
    /**
     * Sanitizes context to remove sensitive data
     */
    private sanitizeContext;
    /**
     * Checks if a key name indicates sensitive data
     */
    private isSensitiveKey;
    authSuccess(req: Request, actorId: string | number, actorType?: 'user' | 'app'): void;
    authFailed(req: Request, reason: string, attemptedId?: string): void;
    tokenRefresh(req: Request, actorId: string | number): void;
    webhookReceived(req: Request, eventType: string, deliveryId?: string): void;
    webhookVerified(req: Request, eventType: string): void;
    webhookRejected(req: Request, reason: string): void;
    webhookProcessed(req: Request, eventType: string, success: boolean, duration: number): void;
    rateLimitExceeded(req: Request, limit: number, remaining: number): void;
    ipBlocked(ip: string, reason: string, duration?: number): void;
    suspiciousActivity(req: Request, description: string, details?: Record<string, unknown>): void;
    signatureInvalid(req: Request, webhookType?: string): void;
    adminAccess(req: Request, endpoint: string, actorId?: string): void;
    adminAction(req: Request, action: string, resourceType: string, resourceId: string | number): void;
    installationCreated(installationId: number, accountLogin: string): void;
    installationDeleted(installationId: number, accountLogin: string): void;
    installationSuspended(installationId: number, accountLogin: string): void;
    reviewStarted(repo: string, prNumber: number, jobId: string): void;
    reviewCompleted(repo: string, prNumber: number, duration: number, issuesFound: number): void;
    reviewFailed(repo: string, prNumber: number, error: string): void;
    jobCreated(jobId: string, type: string, priority: number): void;
    jobProcessed(jobId: string, type: string, duration: number): void;
    jobFailed(jobId: string, type: string, error: string, attempts: number): void;
    jobCancelled(jobId: string, reason?: string): void;
    apiRequest(req: Request, statusCode: number, duration: number): void;
    apiError(req: Request, error: Error, statusCode?: number): void;
}
export declare const auditLog: AuditLogger;
export default auditLog;
//# sourceMappingURL=auditLog.d.ts.map