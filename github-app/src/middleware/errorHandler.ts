import { Request, Response, NextFunction } from 'express';
import { logger } from '../utils/logger.js';
import { ZodError } from 'zod';

export class AppError extends Error {
  public details?: Record<string, unknown>;

  constructor(
    public override message: string,
    public statusCode: number = 500,
    public code?: string | Record<string, unknown>
  ) {
    super(message);
    this.name = 'AppError';
    if (typeof code === 'object') {
      this.details = code;
      this.code = undefined;
    }
  }
}

export const errorHandler = (
  err: Error,
  req: Request,
  res: Response,
  _next: NextFunction
) => {
  logger.error({
    err,
    method: req.method,
    url: req.url,
  }, 'Request error');

  if (err instanceof AppError) {
    return res.status(err.statusCode).json({
      error: err.message,
      code: err.code,
      ...(err.details && { details: err.details }),
    });
  }

  if (err instanceof ZodError) {
    return res.status(400).json({
      error: 'Validation error',
      details: err.errors,
    });
  }

  return res.status(500).json({
    error: 'Internal server error',
  });
};
