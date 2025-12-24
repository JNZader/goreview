/**
 * Security middleware for Express
 * Implements security headers without external dependencies
 */

import type { Request, Response, NextFunction, RequestHandler } from 'express';

export interface SecurityOptions {
  /** Enable/disable specific headers */
  contentSecurityPolicy?: boolean;
  xssFilter?: boolean;
  noSniff?: boolean;
  frameGuard?: boolean | 'deny' | 'sameorigin';
  hsts?: boolean | { maxAge: number; includeSubDomains?: boolean };
  referrerPolicy?: string;
  permittedCrossDomainPolicies?: boolean;
}

const defaultOptions: SecurityOptions = {
  contentSecurityPolicy: true,
  xssFilter: true,
  noSniff: true,
  frameGuard: 'deny',
  hsts: { maxAge: 31536000, includeSubDomains: true },
  referrerPolicy: 'strict-origin-when-cross-origin',
  permittedCrossDomainPolicies: true,
};

/**
 * Creates security headers middleware
 */
export function securityHeaders(options: SecurityOptions = {}): RequestHandler {
  const opts = { ...defaultOptions, ...options };

  return (req: Request, res: Response, next: NextFunction): void => {
    // X-Content-Type-Options
    if (opts.noSniff) {
      res.setHeader('X-Content-Type-Options', 'nosniff');
    }

    // X-Frame-Options
    if (opts.frameGuard) {
      const value = opts.frameGuard === true ? 'DENY' : opts.frameGuard.toUpperCase();
      res.setHeader('X-Frame-Options', value);
    }

    // X-XSS-Protection (legacy but still useful)
    if (opts.xssFilter) {
      res.setHeader('X-XSS-Protection', '1; mode=block');
    }

    // Strict-Transport-Security
    if (opts.hsts) {
      const hstsOpts = opts.hsts === true
        ? { maxAge: 31536000, includeSubDomains: true }
        : opts.hsts;
      let value = `max-age=${hstsOpts.maxAge}`;
      if (hstsOpts.includeSubDomains) {
        value += '; includeSubDomains';
      }
      res.setHeader('Strict-Transport-Security', value);
    }

    // Referrer-Policy
    if (opts.referrerPolicy) {
      res.setHeader('Referrer-Policy', opts.referrerPolicy);
    }

    // X-Permitted-Cross-Domain-Policies
    if (opts.permittedCrossDomainPolicies) {
      res.setHeader('X-Permitted-Cross-Domain-Policies', 'none');
    }

    // Content-Security-Policy (basic policy for APIs)
    if (opts.contentSecurityPolicy) {
      res.setHeader('Content-Security-Policy', "default-src 'none'; frame-ancestors 'none'");
    }

    // Remove X-Powered-By header
    res.removeHeader('X-Powered-By');

    next();
  };
}

/**
 * Hide X-Powered-By header
 */
export function hidePoweredBy(_req: Request, res: Response, next: NextFunction): void {
  res.removeHeader('X-Powered-By');
  next();
}

export default securityHeaders;
