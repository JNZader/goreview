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
    hsts?: boolean | {
        maxAge: number;
        includeSubDomains?: boolean;
    };
    referrerPolicy?: string;
    permittedCrossDomainPolicies?: boolean;
}
/**
 * Creates security headers middleware
 */
export declare function securityHeaders(options?: SecurityOptions): RequestHandler;
/**
 * Hide X-Powered-By header
 */
export declare function hidePoweredBy(_req: Request, res: Response, next: NextFunction): void;
export default securityHeaders;
//# sourceMappingURL=security.d.ts.map