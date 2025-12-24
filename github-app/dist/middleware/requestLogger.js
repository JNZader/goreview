import { logger } from '../utils/logger.js';
export const requestLogger = (req, res, next) => {
    const start = Date.now();
    res.on('finish', () => {
        const duration = Date.now() - start;
        logger.info({
            method: req.method,
            url: req.url,
            status: res.statusCode,
            duration,
            ip: req.ip,
            userAgent: req.get('user-agent'),
        }, 'Request completed');
    });
    next();
};
//# sourceMappingURL=requestLogger.js.map