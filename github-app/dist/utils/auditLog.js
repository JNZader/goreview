/**
 * Security audit logging system
 * Provides structured logging for security-relevant events
 */
import { logger } from './logger.js';
import { getClientIp } from '../middleware/rateLimit.js';
import { maskSecret } from './secrets.js';
// =============================================================================
// Audit Logger Class
// =============================================================================
class AuditLogger {
    serviceName;
    constructor(serviceName = 'github-app') {
        this.serviceName = serviceName;
    }
    /**
     * Logs an audit event
     */
    log(event, message, context = {}, severity = 'info') {
        const entry = {
            timestamp: new Date().toISOString(),
            event,
            severity,
            message,
            context: this.sanitizeContext(context),
        };
        // Log using the appropriate level
        const logMethod = severity === 'critical' ? 'error' : severity;
        logger[logMethod]({
            audit: true,
            service: this.serviceName,
            ...entry,
        });
    }
    /**
     * Extracts audit context from an Express request
     */
    extractContext(req) {
        return {
            requestId: req.headers['x-request-id'],
            ip: getClientIp(req),
            userAgent: req.headers['user-agent'],
            method: req.method,
            path: req.path,
        };
    }
    /**
     * Sanitizes context to remove sensitive data
     */
    sanitizeContext(context) {
        const sanitized = { ...context };
        // Mask potentially sensitive metadata
        if (sanitized.metadata) {
            const maskedMetadata = {};
            for (const [key, value] of Object.entries(sanitized.metadata)) {
                if (this.isSensitiveKey(key) && typeof value === 'string') {
                    maskedMetadata[key] = maskSecret(value);
                }
                else {
                    maskedMetadata[key] = value;
                }
            }
            sanitized.metadata = maskedMetadata;
        }
        return sanitized;
    }
    /**
     * Checks if a key name indicates sensitive data
     */
    isSensitiveKey(key) {
        const sensitivePatterns = ['token', 'secret', 'password', 'key', 'auth', 'credential'];
        return sensitivePatterns.some((pattern) => key.toLowerCase().includes(pattern));
    }
    // ===========================================================================
    // Authentication Events
    // ===========================================================================
    authSuccess(req, actorId, actorType = 'user') {
        this.log('auth.login', `Authentication successful for ${actorType} ${actorId}`, {
            ...this.extractContext(req),
            actorId,
            actorType,
        }, 'info');
    }
    authFailed(req, reason, attemptedId) {
        this.log('auth.failed', `Authentication failed: ${reason}`, {
            ...this.extractContext(req),
            metadata: { reason, attemptedId: attemptedId ? maskSecret(attemptedId) : undefined },
        }, 'warn');
    }
    tokenRefresh(req, actorId) {
        this.log('auth.token_refresh', `Token refreshed for ${actorId}`, {
            ...this.extractContext(req),
            actorId,
            actorType: 'app',
        }, 'info');
    }
    // ===========================================================================
    // Webhook Events
    // ===========================================================================
    webhookReceived(req, eventType, deliveryId) {
        this.log('webhook.received', `Webhook received: ${eventType}`, {
            ...this.extractContext(req),
            metadata: { eventType, deliveryId },
        }, 'info');
    }
    webhookVerified(req, eventType) {
        this.log('webhook.verified', `Webhook signature verified: ${eventType}`, {
            ...this.extractContext(req),
            metadata: { eventType },
        }, 'info');
    }
    webhookRejected(req, reason) {
        this.log('webhook.rejected', `Webhook rejected: ${reason}`, {
            ...this.extractContext(req),
            metadata: { reason },
        }, 'warn');
    }
    webhookProcessed(req, eventType, success, duration) {
        this.log('webhook.processed', `Webhook processed: ${eventType} (${success ? 'success' : 'failed'})`, {
            ...this.extractContext(req),
            duration,
            metadata: { eventType, success },
        }, success ? 'info' : 'warn');
    }
    // ===========================================================================
    // Security Events
    // ===========================================================================
    rateLimitExceeded(req, limit, remaining) {
        this.log('rate_limit.exceeded', `Rate limit exceeded`, {
            ...this.extractContext(req),
            metadata: { limit, remaining },
        }, 'warn');
    }
    ipBlocked(ip, reason, duration) {
        this.log('security.ip_blocked', `IP address blocked: ${reason}`, {
            ip,
            metadata: { reason, durationMs: duration },
        }, 'warn');
    }
    suspiciousActivity(req, description, details) {
        this.log('security.suspicious_activity', `Suspicious activity detected: ${description}`, {
            ...this.extractContext(req),
            metadata: details,
        }, 'critical');
    }
    signatureInvalid(req, webhookType) {
        this.log('security.signature_invalid', `Invalid webhook signature detected`, {
            ...this.extractContext(req),
            metadata: { webhookType },
        }, 'critical');
    }
    // ===========================================================================
    // Admin Events
    // ===========================================================================
    adminAccess(req, endpoint, actorId) {
        this.log('admin.access', `Admin endpoint accessed: ${endpoint}`, {
            ...this.extractContext(req),
            actorId,
            actorType: 'user',
            metadata: { endpoint },
        }, 'info');
    }
    adminAction(req, action, resourceType, resourceId) {
        this.log('admin.action', `Admin action: ${action} on ${resourceType}`, {
            ...this.extractContext(req),
            resourceType,
            resourceId,
            metadata: { action },
        }, 'info');
    }
    // ===========================================================================
    // Installation Events
    // ===========================================================================
    installationCreated(installationId, accountLogin) {
        this.log('installation.created', `App installed on ${accountLogin}`, {
            resourceType: 'installation',
            resourceId: installationId,
            actorType: 'user',
            metadata: { accountLogin },
        }, 'info');
    }
    installationDeleted(installationId, accountLogin) {
        this.log('installation.deleted', `App uninstalled from ${accountLogin}`, {
            resourceType: 'installation',
            resourceId: installationId,
            actorType: 'user',
            metadata: { accountLogin },
        }, 'info');
    }
    installationSuspended(installationId, accountLogin) {
        this.log('installation.suspended', `Installation suspended on ${accountLogin}`, {
            resourceType: 'installation',
            resourceId: installationId,
            actorType: 'user',
            metadata: { accountLogin },
        }, 'warn');
    }
    // ===========================================================================
    // Review Events
    // ===========================================================================
    reviewStarted(repo, prNumber, jobId) {
        this.log('review.started', `Code review started for ${repo}#${prNumber}`, {
            resourceType: 'pull_request',
            resourceId: prNumber,
            metadata: { repo, jobId },
        }, 'info');
    }
    reviewCompleted(repo, prNumber, duration, issuesFound) {
        this.log('review.completed', `Code review completed for ${repo}#${prNumber}`, {
            resourceType: 'pull_request',
            resourceId: prNumber,
            duration,
            metadata: { repo, issuesFound },
        }, 'info');
    }
    reviewFailed(repo, prNumber, error) {
        this.log('review.failed', `Code review failed for ${repo}#${prNumber}: ${error}`, {
            resourceType: 'pull_request',
            resourceId: prNumber,
            metadata: { repo, error },
        }, 'error');
    }
    // ===========================================================================
    // Job Events
    // ===========================================================================
    jobCreated(jobId, type, priority) {
        this.log('job.created', `Job created: ${type}`, {
            resourceType: 'job',
            resourceId: jobId,
            metadata: { type, priority },
        }, 'info');
    }
    jobProcessed(jobId, type, duration) {
        this.log('job.processed', `Job completed: ${type}`, {
            resourceType: 'job',
            resourceId: jobId,
            duration,
            metadata: { type },
        }, 'info');
    }
    jobFailed(jobId, type, error, attempts) {
        this.log('job.failed', `Job failed: ${type}`, {
            resourceType: 'job',
            resourceId: jobId,
            metadata: { type, error, attempts },
        }, 'error');
    }
    jobCancelled(jobId, reason) {
        this.log('job.cancelled', `Job cancelled`, {
            resourceType: 'job',
            resourceId: jobId,
            metadata: { reason },
        }, 'warn');
    }
    // ===========================================================================
    // API Events
    // ===========================================================================
    apiRequest(req, statusCode, duration) {
        this.log('api.request', `${req.method} ${req.path} - ${statusCode}`, {
            ...this.extractContext(req),
            statusCode,
            duration,
        }, statusCode >= 400 ? 'warn' : 'info');
    }
    apiError(req, error, statusCode = 500) {
        this.log('api.error', `API error: ${error.message}`, {
            ...this.extractContext(req),
            statusCode,
            metadata: {
                errorName: error.name,
                errorMessage: error.message,
            },
        }, statusCode >= 500 ? 'error' : 'warn');
    }
}
// =============================================================================
// Singleton Instance
// =============================================================================
export const auditLog = new AuditLogger();
// =============================================================================
// Exports
// =============================================================================
export default auditLog;
//# sourceMappingURL=auditLog.js.map