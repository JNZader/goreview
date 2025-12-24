/**
 * Request ID middleware for tracing and debugging
 */
import crypto from 'node:crypto';
/**
 * Generates a unique request ID
 */
function generateRequestId() {
    return crypto.randomUUID();
}
/**
 * Middleware that adds a unique request ID to each request
 * Uses X-Request-ID header if present, otherwise generates a new one
 */
export function requestIdMiddleware(req, res, next) {
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
//# sourceMappingURL=requestId.js.map