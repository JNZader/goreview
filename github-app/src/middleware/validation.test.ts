/**
 * Tests for input validation middleware
 */

import { describe, it, expect, vi, beforeEach } from 'vitest';
import type { Request, Response, NextFunction } from 'express';
import { z } from 'zod';
import {
  validateRequest,
  validateBody,
  validateQuery,
  validateParams,
  validateWebhookSignature,
  webhookSignatureMiddleware,
  sanitizeString,
  sanitizeObject,
  PullRequestPayloadSchema,
  InstallationPayloadSchema,
  GitHubUserSchema,
} from './validation.js';

// Mock logger
vi.mock('../utils/logger.js', () => ({
  logger: {
    info: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));

describe('validation', () => {
  // Helper to create mock request
  const createMockRequest = (overrides: Partial<Request> = {}): Request => {
    return {
      body: {},
      query: {},
      params: {},
      headers: {},
      path: '/test',
      method: 'POST',
      ...overrides,
    } as Request;
  };

  // Helper to create mock response
  const createMockResponse = (): Response => {
    return {
      status: vi.fn().mockReturnThis(),
      json: vi.fn().mockReturnThis(),
    } as unknown as Response;
  };

  // ===========================================================================
  // Sanitization Tests
  // ===========================================================================

  describe('sanitizeString', () => {
    it('should remove control characters', () => {
      const input = 'hello\x00\x08\x7Fworld';
      const result = sanitizeString(input);

      expect(result).toBe('helloworld');
    });

    it('should remove script tags', () => {
      const input = 'before<script>alert("xss")</script>after';
      const result = sanitizeString(input);

      expect(result).toBe('beforeafter');
    });

    it('should trim whitespace', () => {
      const result = sanitizeString('  hello  ');

      expect(result).toBe('hello');
    });

    it('should handle non-string input', () => {
      expect(sanitizeString(null as unknown as string)).toBe('');
      expect(sanitizeString(undefined as unknown as string)).toBe('');
      expect(sanitizeString(123 as unknown as string)).toBe('');
    });
  });

  describe('sanitizeObject', () => {
    it('should sanitize string values', () => {
      const obj = {
        name: '  John<script>xss</script>  ',
        age: 30,
      };

      const result = sanitizeObject(obj);

      expect(result.name).toBe('John');
      expect(result.age).toBe(30);
    });

    it('should sanitize nested objects', () => {
      const obj = {
        user: {
          name: '<script>alert(1)</script>Bob',
        },
      };

      const result = sanitizeObject(obj);

      expect(result.user.name).toBe('Bob');
    });

    it('should sanitize arrays', () => {
      const obj = {
        tags: ['<script>xss</script>tag1', 'tag2'],
      };

      const result = sanitizeObject(obj);

      expect(result.tags[0]).toBe('tag1');
      expect(result.tags[1]).toBe('tag2');
    });

    it('should handle nested arrays with objects', () => {
      const obj = {
        items: [
          { name: '  item1<script></script>  ' },
          { name: 'item2' },
        ],
      };

      const result = sanitizeObject(obj);

      expect(result.items[0].name).toBe('item1');
    });
  });

  // ===========================================================================
  // Schema Validation Tests
  // ===========================================================================

  describe('GitHubUserSchema', () => {
    it('should validate valid user', () => {
      const user = {
        id: 12345,
        login: 'testuser',
      };

      const result = GitHubUserSchema.safeParse(user);

      expect(result.success).toBe(true);
    });

    it('should reject invalid user id', () => {
      const user = {
        id: -1,
        login: 'testuser',
      };

      const result = GitHubUserSchema.safeParse(user);

      expect(result.success).toBe(false);
    });

    it('should reject empty login', () => {
      const user = {
        id: 123,
        login: '',
      };

      const result = GitHubUserSchema.safeParse(user);

      expect(result.success).toBe(false);
    });
  });

  describe('PullRequestPayloadSchema', () => {
    const validPayload = {
      action: 'opened',
      number: 1,
      pull_request: {
        id: 100,
        number: 1,
        state: 'open',
        title: 'Test PR',
        draft: false,
        head: {
          ref: 'feature-branch',
          sha: 'a'.repeat(40),
        },
        base: {
          ref: 'main',
          sha: 'b'.repeat(40),
        },
        user: {
          id: 1,
          login: 'testuser',
        },
      },
      repository: {
        id: 200,
        name: 'test-repo',
        full_name: 'owner/test-repo',
        private: false,
        owner: {
          id: 1,
          login: 'owner',
        },
      },
      sender: {
        id: 1,
        login: 'testuser',
      },
    };

    it('should validate valid PR payload', () => {
      const result = PullRequestPayloadSchema.safeParse(validPayload);

      expect(result.success).toBe(true);
    });

    it('should reject invalid action', () => {
      const payload = { ...validPayload, action: 'invalid' };
      const result = PullRequestPayloadSchema.safeParse(payload);

      expect(result.success).toBe(false);
    });

    it('should reject invalid SHA format', () => {
      const payload = {
        ...validPayload,
        pull_request: {
          ...validPayload.pull_request,
          head: {
            ref: 'branch',
            sha: 'invalid-sha',
          },
        },
      };
      const result = PullRequestPayloadSchema.safeParse(payload);

      expect(result.success).toBe(false);
    });

    it('should accept all valid actions', () => {
      const actions = [
        'opened', 'closed', 'reopened', 'synchronize', 'edited',
        'ready_for_review', 'review_requested', 'labeled',
      ];

      for (const action of actions) {
        const payload = { ...validPayload, action };
        const result = PullRequestPayloadSchema.safeParse(payload);

        expect(result.success).toBe(true);
      }
    });
  });

  describe('InstallationPayloadSchema', () => {
    const validPayload = {
      action: 'created',
      installation: {
        id: 12345,
      },
      sender: {
        id: 1,
        login: 'installer',
      },
    };

    it('should validate valid installation payload', () => {
      const result = InstallationPayloadSchema.safeParse(validPayload);

      expect(result.success).toBe(true);
    });

    it('should accept all valid actions', () => {
      const actions = ['created', 'deleted', 'suspend', 'unsuspend', 'new_permissions_accepted'];

      for (const action of actions) {
        const payload = { ...validPayload, action };
        const result = InstallationPayloadSchema.safeParse(payload);

        expect(result.success).toBe(true);
      }
    });
  });

  // ===========================================================================
  // Middleware Tests
  // ===========================================================================

  describe('validateBody', () => {
    const testSchema = z.object({
      name: z.string().min(1),
      value: z.number().positive(),
    });

    it('should pass valid body', () => {
      const req = createMockRequest({
        body: { name: 'test', value: 42 },
      });
      const res = createMockResponse();
      const next = vi.fn();

      const middleware = validateBody(testSchema);
      middleware(req, res, next);

      expect(next).toHaveBeenCalled();
      expect(req.body.name).toBe('test');
    });

    it('should reject invalid body', () => {
      const req = createMockRequest({
        body: { name: '', value: -1 },
      });
      const res = createMockResponse();
      const next = vi.fn();

      const middleware = validateBody(testSchema);
      middleware(req, res, next);

      expect(next).toHaveBeenCalledWith(expect.any(Error));
    });

    it('should sanitize input before validation', () => {
      // Script tags are removed entirely, so use content outside tags
      const req = createMockRequest({
        body: { name: '  hello<script>xss</script>world  ', value: 42 },
      });
      const res = createMockResponse();
      const next = vi.fn();

      const middleware = validateBody(testSchema);
      middleware(req, res, next);

      expect(next).toHaveBeenCalled();
      expect(req.body.name).toBe('helloworld');
    });
  });

  describe('validateQuery', () => {
    const testSchema = z.object({
      page: z.string().optional(),
      limit: z.string().optional(),
    });

    it('should pass valid query', () => {
      const req = createMockRequest({
        query: { page: '1', limit: '10' },
      });
      const res = createMockResponse();
      const next = vi.fn();

      const middleware = validateQuery(testSchema);
      middleware(req, res, next);

      expect(next).toHaveBeenCalled();
    });
  });

  describe('validateParams', () => {
    const testSchema = z.object({
      id: z.string().uuid(),
    });

    it('should pass valid params', () => {
      const req = createMockRequest({
        params: { id: '550e8400-e29b-41d4-a716-446655440000' },
      });
      const res = createMockResponse();
      const next = vi.fn();

      const middleware = validateParams(testSchema);
      middleware(req, res, next);

      expect(next).toHaveBeenCalled();
    });

    it('should reject invalid UUID', () => {
      const req = createMockRequest({
        params: { id: 'not-a-uuid' },
      });
      const res = createMockResponse();
      const next = vi.fn();

      const middleware = validateParams(testSchema);
      middleware(req, res, next);

      expect(next).toHaveBeenCalledWith(expect.any(Error));
    });
  });

  // ===========================================================================
  // Webhook Signature Tests
  // ===========================================================================

  describe('validateWebhookSignature', () => {
    const secret = 'webhook-secret';
    const payload = '{"action":"opened"}';

    it('should validate correct signature', () => {
      const crypto = require('node:crypto');
      const signature = 'sha256=' + crypto
        .createHmac('sha256', secret)
        .update(payload)
        .digest('hex');

      const result = validateWebhookSignature(payload, signature, secret);

      expect(result).toBe(true);
    });

    it('should reject incorrect signature', () => {
      const result = validateWebhookSignature(payload, 'sha256=invalid', secret);

      expect(result).toBe(false);
    });

    it('should reject missing signature', () => {
      const result = validateWebhookSignature(payload, undefined, secret);

      expect(result).toBe(false);
    });

    it('should reject non-sha256 algorithm', () => {
      const result = validateWebhookSignature(payload, 'sha1=abcdef', secret);

      expect(result).toBe(false);
    });

    it('should reject malformed signature', () => {
      const result = validateWebhookSignature(payload, 'invalid-format', secret);

      expect(result).toBe(false);
    });
  });

  describe('webhookSignatureMiddleware', () => {
    const secret = 'test-secret';

    it('should reject request without raw body', () => {
      const req = createMockRequest({
        headers: { 'x-hub-signature-256': 'sha256=test' },
      });
      const res = createMockResponse();
      const next = vi.fn();

      const middleware = webhookSignatureMiddleware(secret);
      middleware(req, res, next);

      expect(next).toHaveBeenCalledWith(expect.any(Error));
    });

    it('should reject invalid signature', () => {
      const req = createMockRequest({
        headers: { 'x-hub-signature-256': 'sha256=invalid' },
      }) as Request & { rawBody: Buffer };
      req.rawBody = Buffer.from('{"test": true}');

      const res = createMockResponse();
      const next = vi.fn();

      const middleware = webhookSignatureMiddleware(secret);
      middleware(req, res, next);

      expect(next).toHaveBeenCalledWith(expect.any(Error));
    });

    it('should pass valid signature', () => {
      const crypto = require('node:crypto');
      const payload = '{"test": true}';
      const signature = 'sha256=' + crypto
        .createHmac('sha256', secret)
        .update(payload)
        .digest('hex');

      const req = createMockRequest({
        headers: { 'x-hub-signature-256': signature },
      }) as Request & { rawBody: Buffer };
      req.rawBody = Buffer.from(payload);

      const res = createMockResponse();
      const next = vi.fn();

      const middleware = webhookSignatureMiddleware(secret);
      middleware(req, res, next);

      expect(next).toHaveBeenCalled();
      expect(next).not.toHaveBeenCalledWith(expect.any(Error));
    });
  });
});
