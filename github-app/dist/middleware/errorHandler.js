import { logger } from '../utils/logger.js';
import { ZodError } from 'zod';
export class AppError extends Error {
    message;
    statusCode;
    code;
    details;
    constructor(message, statusCode = 500, code) {
        super(message);
        this.message = message;
        this.statusCode = statusCode;
        this.code = code;
        this.name = 'AppError';
        if (typeof code === 'object') {
            this.details = code;
            this.code = undefined;
        }
    }
}
export const errorHandler = (err, req, res, _next) => {
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
//# sourceMappingURL=errorHandler.js.map