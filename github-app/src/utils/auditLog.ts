/**
 * Security audit logging system
 * Provides structured logging for security-relevant events
 */

import type { Request } from 'express';
import { logger } from './logger.js';
import { getClientIp } from '../middleware/rateLimit.js';
import { maskSecret } from './secrets.js';

// =============================================================================
// Types
// =============================================================================

export type AuditEventType =
  | 'auth.login'
  | 'auth.logout'
  | 'auth.failed'
  | 'auth.token_refresh'
  | 'webhook.received'
  | 'webhook.verified'
  | 'webhook.rejected'
  | 'webhook.processed'
  | 'api.request'
  | 'api.error'
  | 'rate_limit.exceeded'
  | 'rate_limit.blocked'
  | 'security.ip_blocked'
  | 'security.suspicious_activity'
  | 'security.signature_invalid'
  | 'admin.access'
  | 'admin.action'
  | 'installation.created'
  | 'installation.deleted'
  | 'installation.suspended'
  | 'review.started'
  | 'review.completed'
  | 'review.failed'
  | 'job.created'
  | 'job.processed'
  | 'job.failed'
  | 'job.cancelled';

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

// =============================================================================
// Audit Logger Class
// =============================================================================

class AuditLogger {
  private readonly serviceName: string;

  constructor(serviceName: string = 'github-app') {
    this.serviceName = serviceName;
  }

  /**
   * Logs an audit event
   */
  log(
    event: AuditEventType,
    message: string,
    context: AuditContext = {},
    severity: AuditSeverity = 'info'
  ): void {
    const entry: AuditEntry = {
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
  extractContext(req: Request): AuditContext {
    return {
      requestId: req.headers['x-request-id'] as string,
      ip: getClientIp(req),
      userAgent: req.headers['user-agent'],
      method: req.method,
      path: req.path,
    };
  }

  /**
   * Sanitizes context to remove sensitive data
   */
  private sanitizeContext(context: AuditContext): AuditContext {
    const sanitized = { ...context };

    // Mask potentially sensitive metadata
    if (sanitized.metadata) {
      const maskedMetadata: Record<string, unknown> = {};
      for (const [key, value] of Object.entries(sanitized.metadata)) {
        if (this.isSensitiveKey(key) && typeof value === 'string') {
          maskedMetadata[key] = maskSecret(value);
        } else {
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
  private isSensitiveKey(key: string): boolean {
    const sensitivePatterns = ['token', 'secret', 'password', 'key', 'auth', 'credential'];
    return sensitivePatterns.some((pattern) => key.toLowerCase().includes(pattern));
  }

  // ===========================================================================
  // Authentication Events
  // ===========================================================================

  authSuccess(req: Request, actorId: string | number, actorType: 'user' | 'app' = 'user'): void {
    this.log(
      'auth.login',
      `Authentication successful for ${actorType} ${actorId}`,
      {
        ...this.extractContext(req),
        actorId,
        actorType,
      },
      'info'
    );
  }

  authFailed(req: Request, reason: string, attemptedId?: string): void {
    this.log(
      'auth.failed',
      `Authentication failed: ${reason}`,
      {
        ...this.extractContext(req),
        metadata: { reason, attemptedId: attemptedId ? maskSecret(attemptedId) : undefined },
      },
      'warn'
    );
  }

  tokenRefresh(req: Request, actorId: string | number): void {
    this.log(
      'auth.token_refresh',
      `Token refreshed for ${actorId}`,
      {
        ...this.extractContext(req),
        actorId,
        actorType: 'app',
      },
      'info'
    );
  }

  // ===========================================================================
  // Webhook Events
  // ===========================================================================

  webhookReceived(req: Request, eventType: string, deliveryId?: string): void {
    this.log(
      'webhook.received',
      `Webhook received: ${eventType}`,
      {
        ...this.extractContext(req),
        metadata: { eventType, deliveryId },
      },
      'info'
    );
  }

  webhookVerified(req: Request, eventType: string): void {
    this.log(
      'webhook.verified',
      `Webhook signature verified: ${eventType}`,
      {
        ...this.extractContext(req),
        metadata: { eventType },
      },
      'info'
    );
  }

  webhookRejected(req: Request, reason: string): void {
    this.log(
      'webhook.rejected',
      `Webhook rejected: ${reason}`,
      {
        ...this.extractContext(req),
        metadata: { reason },
      },
      'warn'
    );
  }

  webhookProcessed(
    req: Request,
    eventType: string,
    success: boolean,
    duration: number
  ): void {
    this.log(
      'webhook.processed',
      `Webhook processed: ${eventType} (${success ? 'success' : 'failed'})`,
      {
        ...this.extractContext(req),
        duration,
        metadata: { eventType, success },
      },
      success ? 'info' : 'warn'
    );
  }

  // ===========================================================================
  // Security Events
  // ===========================================================================

  rateLimitExceeded(req: Request, limit: number, remaining: number): void {
    this.log(
      'rate_limit.exceeded',
      `Rate limit exceeded`,
      {
        ...this.extractContext(req),
        metadata: { limit, remaining },
      },
      'warn'
    );
  }

  ipBlocked(ip: string, reason: string, duration?: number): void {
    this.log(
      'security.ip_blocked',
      `IP address blocked: ${reason}`,
      {
        ip,
        metadata: { reason, durationMs: duration },
      },
      'warn'
    );
  }

  suspiciousActivity(req: Request, description: string, details?: Record<string, unknown>): void {
    this.log(
      'security.suspicious_activity',
      `Suspicious activity detected: ${description}`,
      {
        ...this.extractContext(req),
        metadata: details,
      },
      'critical'
    );
  }

  signatureInvalid(req: Request, webhookType?: string): void {
    this.log(
      'security.signature_invalid',
      `Invalid webhook signature detected`,
      {
        ...this.extractContext(req),
        metadata: { webhookType },
      },
      'critical'
    );
  }

  // ===========================================================================
  // Admin Events
  // ===========================================================================

  adminAccess(req: Request, endpoint: string, actorId?: string): void {
    this.log(
      'admin.access',
      `Admin endpoint accessed: ${endpoint}`,
      {
        ...this.extractContext(req),
        actorId,
        actorType: 'user',
        metadata: { endpoint },
      },
      'info'
    );
  }

  adminAction(
    req: Request,
    action: string,
    resourceType: string,
    resourceId: string | number
  ): void {
    this.log(
      'admin.action',
      `Admin action: ${action} on ${resourceType}`,
      {
        ...this.extractContext(req),
        resourceType,
        resourceId,
        metadata: { action },
      },
      'info'
    );
  }

  // ===========================================================================
  // Installation Events
  // ===========================================================================

  installationCreated(installationId: number, accountLogin: string): void {
    this.log(
      'installation.created',
      `App installed on ${accountLogin}`,
      {
        resourceType: 'installation',
        resourceId: installationId,
        actorType: 'user',
        metadata: { accountLogin },
      },
      'info'
    );
  }

  installationDeleted(installationId: number, accountLogin: string): void {
    this.log(
      'installation.deleted',
      `App uninstalled from ${accountLogin}`,
      {
        resourceType: 'installation',
        resourceId: installationId,
        actorType: 'user',
        metadata: { accountLogin },
      },
      'info'
    );
  }

  installationSuspended(installationId: number, accountLogin: string): void {
    this.log(
      'installation.suspended',
      `Installation suspended on ${accountLogin}`,
      {
        resourceType: 'installation',
        resourceId: installationId,
        actorType: 'user',
        metadata: { accountLogin },
      },
      'warn'
    );
  }

  // ===========================================================================
  // Review Events
  // ===========================================================================

  reviewStarted(
    repo: string,
    prNumber: number,
    jobId: string
  ): void {
    this.log(
      'review.started',
      `Code review started for ${repo}#${prNumber}`,
      {
        resourceType: 'pull_request',
        resourceId: prNumber,
        metadata: { repo, jobId },
      },
      'info'
    );
  }

  reviewCompleted(
    repo: string,
    prNumber: number,
    duration: number,
    issuesFound: number
  ): void {
    this.log(
      'review.completed',
      `Code review completed for ${repo}#${prNumber}`,
      {
        resourceType: 'pull_request',
        resourceId: prNumber,
        duration,
        metadata: { repo, issuesFound },
      },
      'info'
    );
  }

  reviewFailed(
    repo: string,
    prNumber: number,
    error: string
  ): void {
    this.log(
      'review.failed',
      `Code review failed for ${repo}#${prNumber}: ${error}`,
      {
        resourceType: 'pull_request',
        resourceId: prNumber,
        metadata: { repo, error },
      },
      'error'
    );
  }

  // ===========================================================================
  // Job Events
  // ===========================================================================

  jobCreated(jobId: string, type: string, priority: number): void {
    this.log(
      'job.created',
      `Job created: ${type}`,
      {
        resourceType: 'job',
        resourceId: jobId,
        metadata: { type, priority },
      },
      'info'
    );
  }

  jobProcessed(jobId: string, type: string, duration: number): void {
    this.log(
      'job.processed',
      `Job completed: ${type}`,
      {
        resourceType: 'job',
        resourceId: jobId,
        duration,
        metadata: { type },
      },
      'info'
    );
  }

  jobFailed(jobId: string, type: string, error: string, attempts: number): void {
    this.log(
      'job.failed',
      `Job failed: ${type}`,
      {
        resourceType: 'job',
        resourceId: jobId,
        metadata: { type, error, attempts },
      },
      'error'
    );
  }

  jobCancelled(jobId: string, reason?: string): void {
    this.log(
      'job.cancelled',
      `Job cancelled`,
      {
        resourceType: 'job',
        resourceId: jobId,
        metadata: { reason },
      },
      'warn'
    );
  }

  // ===========================================================================
  // API Events
  // ===========================================================================

  apiRequest(req: Request, statusCode: number, duration: number): void {
    this.log(
      'api.request',
      `${req.method} ${req.path} - ${statusCode}`,
      {
        ...this.extractContext(req),
        statusCode,
        duration,
      },
      statusCode >= 400 ? 'warn' : 'info'
    );
  }

  apiError(req: Request, error: Error, statusCode: number = 500): void {
    this.log(
      'api.error',
      `API error: ${error.message}`,
      {
        ...this.extractContext(req),
        statusCode,
        metadata: {
          errorName: error.name,
          errorMessage: error.message,
        },
      },
      statusCode >= 500 ? 'error' : 'warn'
    );
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
