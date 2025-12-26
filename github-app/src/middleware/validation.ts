/**
 * Input validation middleware using Zod schemas
 * Provides comprehensive validation for webhook payloads and API requests
 */

import type { Request, Response, NextFunction, RequestHandler } from 'express';
import { z } from 'zod';
import { logger } from '../utils/logger.js';
import { AppError } from './errorHandler.js';

// =============================================================================
// Base Schemas
// =============================================================================

/**
 * GitHub user schema
 */
export const GitHubUserSchema = z.object({
  id: z.number().int().positive(),
  login: z.string().min(1).max(39),
  node_id: z.string().optional(),
  avatar_url: z.string().url().optional(),
  type: z.enum(['User', 'Bot', 'Organization']).optional(),
});

/**
 * GitHub repository schema
 */
export const GitHubRepositorySchema = z.object({
  id: z.number().int().positive(),
  node_id: z.string().optional(),
  name: z.string().min(1).max(100),
  full_name: z.string().min(1).max(200),
  private: z.boolean(),
  owner: GitHubUserSchema,
  default_branch: z.string().optional(),
  language: z.string().nullable().optional(),
});

/**
 * GitHub installation schema
 */
export const GitHubInstallationSchema = z.object({
  id: z.number().int().positive(),
  account: GitHubUserSchema.optional(),
  repository_selection: z.enum(['all', 'selected']).optional(),
  permissions: z.record(z.string()).optional(),
});

// =============================================================================
// Webhook Payload Schemas
// =============================================================================

/**
 * Pull request webhook payload
 */
export const PullRequestPayloadSchema = z.object({
  action: z.enum([
    'opened',
    'closed',
    'reopened',
    'synchronize',
    'edited',
    'ready_for_review',
    'review_requested',
    'labeled',
    'unlabeled',
    'assigned',
    'unassigned',
  ]),
  number: z.number().int().positive(),
  pull_request: z.object({
    id: z.number().int().positive(),
    number: z.number().int().positive(),
    state: z.enum(['open', 'closed']),
    title: z.string().min(1).max(256),
    body: z.string().nullable().optional(),
    draft: z.boolean().optional(),
    head: z.object({
      ref: z.string().min(1).max(255),
      sha: z.string().regex(/^[a-f0-9]{40}$/),
      repo: GitHubRepositorySchema.nullable().optional(),
    }),
    base: z.object({
      ref: z.string().min(1).max(255),
      sha: z.string().regex(/^[a-f0-9]{40}$/),
      repo: GitHubRepositorySchema.nullable().optional(),
    }),
    user: GitHubUserSchema,
    merged: z.boolean().optional(),
    mergeable: z.boolean().nullable().optional(),
    changed_files: z.number().int().nonnegative().optional(),
    additions: z.number().int().nonnegative().optional(),
    deletions: z.number().int().nonnegative().optional(),
  }),
  repository: GitHubRepositorySchema,
  installation: GitHubInstallationSchema.optional(),
  sender: GitHubUserSchema,
});

/**
 * Installation webhook payload
 */
export const InstallationPayloadSchema = z.object({
  action: z.enum(['created', 'deleted', 'suspend', 'unsuspend', 'new_permissions_accepted']),
  installation: GitHubInstallationSchema,
  repositories: z
    .array(
      z.object({
        id: z.number().int().positive(),
        name: z.string().min(1).max(100),
        full_name: z.string().min(1).max(200),
        private: z.boolean(),
      })
    )
    .optional(),
  sender: GitHubUserSchema,
});

/**
 * Push webhook payload
 */
export const PushPayloadSchema = z.object({
  ref: z.string().min(1),
  before: z.string().regex(/^[a-f0-9]{40}$/),
  after: z.string().regex(/^[a-f0-9]{40}$/),
  created: z.boolean().optional(),
  deleted: z.boolean().optional(),
  forced: z.boolean().optional(),
  commits: z
    .array(
      z.object({
        id: z.string().regex(/^[a-f0-9]{40}$/),
        message: z.string(),
        author: z.object({
          name: z.string(),
          email: z.string().email(),
        }),
        added: z.array(z.string()).optional(),
        removed: z.array(z.string()).optional(),
        modified: z.array(z.string()).optional(),
      })
    )
    .optional(),
  repository: GitHubRepositorySchema,
  installation: GitHubInstallationSchema.optional(),
  sender: GitHubUserSchema,
});

/**
 * Ping webhook payload
 */
export const PingPayloadSchema = z.object({
  zen: z.string(),
  hook_id: z.number().int().positive(),
  hook: z
    .object({
      id: z.number().int().positive(),
      type: z.string(),
      events: z.array(z.string()),
      active: z.boolean(),
    })
    .optional(),
  repository: GitHubRepositorySchema.optional(),
  sender: GitHubUserSchema.optional(),
});

// =============================================================================
// API Request Schemas
// =============================================================================

/**
 * Admin job query params
 */
export const JobQuerySchema = z.object({
  status: z.enum(['pending', 'processing', 'completed', 'failed']).optional(),
  limit: z
    .string()
    .transform((val) => Number.parseInt(val, 10))
    .pipe(z.number().int().positive().max(100))
    .optional()
    .default('50'),
  offset: z
    .string()
    .transform((val) => Number.parseInt(val, 10))
    .pipe(z.number().int().nonnegative())
    .optional()
    .default('0'),
});

/**
 * Job ID param
 */
export const JobIdParamSchema = z.object({
  id: z.string().uuid(),
});

// =============================================================================
// Input Sanitization
// =============================================================================

/**
 * Sanitizes a string by removing potentially dangerous characters.
 * Note: For webhook payloads from GitHub, we trust the source.
 * This sanitization removes control characters that could cause issues in logging/display.
 */
export function sanitizeString(input: string): string {
  if (typeof input !== 'string') {
    return '';
  }

  // Remove ASCII control characters (except tab, newline, carriage return)
  // Using character code filtering instead of regex with literal control chars (SonarQube S6324)
  return input
    .split('')
    .filter((char) => {
      const code = char.codePointAt(0) ?? 0;
      // Allow tab (9), newline (10), carriage return (13), and printable chars (32-126, 128+)
      return code === 9 || code === 10 || code === 13 || (code >= 32 && code !== 127);
    })
    .join('')
    .trim();
}

/**
 * Sanitizes a single value (string, object, or other)
 */
function sanitizeValue(value: unknown): unknown {
  if (typeof value === 'string') {
    return sanitizeString(value);
  }
  if (typeof value === 'object' && value !== null) {
    return sanitizeObject(value as Record<string, unknown>);
  }
  return value;
}

/**
 * Sanitizes an object recursively
 */
export function sanitizeObject<T extends Record<string, unknown>>(obj: T): T {
  const sanitized: Record<string, unknown> = {};

  for (const [key, value] of Object.entries(obj)) {
    if (Array.isArray(value)) {
      sanitized[key] = value.map(sanitizeValue);
    } else {
      sanitized[key] = sanitizeValue(value);
    }
  }

  return sanitized as T;
}

// =============================================================================
// Validation Middleware Factory
// =============================================================================

interface ValidationOptions {
  /** Where to find the data to validate */
  source?: 'body' | 'query' | 'params';
  /** Whether to sanitize strings */
  sanitize?: boolean;
  /** Custom error message */
  errorMessage?: string;
  /** Whether to strip unknown keys */
  stripUnknown?: boolean;
}

/**
 * Creates a validation middleware for a Zod schema
 */
export function validateRequest<T extends z.ZodType>(
  schema: T,
  options: ValidationOptions = {}
): RequestHandler {
  const { source = 'body', sanitize = true, errorMessage, stripUnknown = true } = options;

  return (req: Request, _res: Response, next: NextFunction): void => {
    try {
      let data = req[source];

      // Sanitize input if enabled
      if (sanitize && typeof data === 'object' && data !== null) {
        data = sanitizeObject(data as Record<string, unknown>);
      }

      // Parse and validate with schema
      // Note: stripUnknown is handled by Zod's default behavior (strips unknown by default)
      const result = schema.safeParse(data);

      if (!result.success) {
        const errors = result.error.errors.map((e) => ({
          path: e.path.join('.'),
          message: e.message,
          code: e.code,
        }));

        logger.warn({
          event: 'validation_failed',
          source,
          errors,
          path: req.path,
          method: req.method,
        });

        throw new AppError(errorMessage ?? 'Validation failed', 400, { errors });
      }

      // Replace request data with validated/sanitized data
      (req as unknown as Record<string, unknown>)[source] = result.data;

      next();
    } catch (error) {
      if (error instanceof AppError) {
        next(error);
        return;
      }

      logger.error({
        event: 'validation_error',
        error: error instanceof Error ? error.message : 'Unknown error',
        path: req.path,
      });

      next(new AppError('Invalid request data', 400));
    }
  };
}

/**
 * Validates request body
 */
export function validateBody<T extends z.ZodType>(
  schema: T,
  options?: Omit<ValidationOptions, 'source'>
): RequestHandler {
  return validateRequest(schema, { ...options, source: 'body' });
}

/**
 * Validates query parameters
 */
export function validateQuery<T extends z.ZodType>(
  schema: T,
  options?: Omit<ValidationOptions, 'source'>
): RequestHandler {
  return validateRequest(schema, { ...options, source: 'query' });
}

/**
 * Validates route parameters
 */
export function validateParams<T extends z.ZodType>(
  schema: T,
  options?: Omit<ValidationOptions, 'source'>
): RequestHandler {
  return validateRequest(schema, { ...options, source: 'params' });
}

// =============================================================================
// Webhook Signature Validation
// =============================================================================

import crypto from 'node:crypto';

/**
 * Validates GitHub webhook signature
 */
export function validateWebhookSignature(
  payload: string | Buffer,
  signature: string | undefined,
  secret: string
): boolean {
  if (!signature) {
    return false;
  }

  const parts = signature.split('=');
  if (parts.length !== 2 || parts[0] !== 'sha256') {
    return false;
  }

  const expected = crypto.createHmac('sha256', secret).update(payload).digest('hex');

  // Timing-safe comparison
  const signatureHex = parts[1];
  if (!signatureHex) {
    return false;
  }

  try {
    return crypto.timingSafeEqual(Buffer.from(signatureHex, 'hex'), Buffer.from(expected, 'hex'));
  } catch {
    return false;
  }
}

/**
 * Middleware to validate GitHub webhook signatures
 */
export function webhookSignatureMiddleware(secret: string): RequestHandler {
  return (req: Request, _res: Response, next: NextFunction): void => {
    const signature = req.headers['x-hub-signature-256'] as string | undefined;
    const rawBody = (req as unknown as { rawBody?: Buffer }).rawBody;

    if (!rawBody) {
      logger.error({
        event: 'webhook_signature_missing_body',
        path: req.path,
      });
      next(new AppError('Raw body not available for signature verification', 500));
      return;
    }

    if (!validateWebhookSignature(rawBody, signature, secret)) {
      logger.warn({
        event: 'webhook_signature_invalid',
        path: req.path,
        hasSignature: !!signature,
      });
      next(new AppError('Invalid webhook signature', 401));
      return;
    }

    next();
  };
}

// =============================================================================
// Pre-built Validators
// =============================================================================

export const validators = {
  pullRequest: validateBody(PullRequestPayloadSchema),
  installation: validateBody(InstallationPayloadSchema),
  push: validateBody(PushPayloadSchema),
  ping: validateBody(PingPayloadSchema),
  jobQuery: validateQuery(JobQuerySchema),
  jobIdParam: validateParams(JobIdParamSchema),
};

// =============================================================================
// Type Exports
// =============================================================================

export type PullRequestPayload = z.infer<typeof PullRequestPayloadSchema>;
export type InstallationPayload = z.infer<typeof InstallationPayloadSchema>;
export type PushPayload = z.infer<typeof PushPayloadSchema>;
export type PingPayload = z.infer<typeof PingPayloadSchema>;
export type JobQuery = z.infer<typeof JobQuerySchema>;
export type JobIdParam = z.infer<typeof JobIdParamSchema>;

// =============================================================================
// Exports
// =============================================================================

export const validation = {
  validateRequest,
  validateBody,
  validateQuery,
  validateParams,
  validateWebhookSignature,
  webhookSignatureMiddleware,
  sanitizeString,
  sanitizeObject,
  validators,
  schemas: {
    GitHubUserSchema,
    GitHubRepositorySchema,
    GitHubInstallationSchema,
    PullRequestPayloadSchema,
    InstallationPayloadSchema,
    PushPayloadSchema,
    PingPayloadSchema,
    JobQuerySchema,
    JobIdParamSchema,
  },
};

export default validation;
