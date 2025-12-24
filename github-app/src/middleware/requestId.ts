/**
 * Request ID middleware for tracing and debugging
 */

import type { Request, Response, NextFunction } from 'express';
import crypto from 'node:crypto';

// Extend Express Request type using module augmentation
declare module 'express-serve-static-core' {
  interface Request {
    id: string;
  }
}

/**
 * Generates a unique request ID
 */
function generateRequestId(): string {
  return crypto.randomUUID();
}

/**
 * Middleware that adds a unique request ID to each request
 * Uses X-Request-ID header if present, otherwise generates a new one
 */
export function requestIdMiddleware(req: Request, res: Response, next: NextFunction): void {
  // Use existing header or generate new ID
  const existingId = req.headers['x-request-id'];
  const requestId = typeof existingId === 'string' ? existingId : generateRequestId();

  // Attach to request object
  req.id = requestId;

  // Add to response headers for client tracking
  res.setHeader('X-Request-ID', requestId);

  next();
}

export default requestIdMiddleware;
