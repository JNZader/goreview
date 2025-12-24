/**
 * Request ID middleware for tracing and debugging
 */
import type { Request, Response, NextFunction } from 'express';
declare module 'express-serve-static-core' {
    interface Request {
        id: string;
    }
}
/**
 * Middleware that adds a unique request ID to each request
 * Uses X-Request-ID header if present, otherwise generates a new one
 */
export declare function requestIdMiddleware(req: Request, res: Response, next: NextFunction): void;
export default requestIdMiddleware;
//# sourceMappingURL=requestId.d.ts.map