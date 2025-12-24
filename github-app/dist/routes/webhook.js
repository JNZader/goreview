import { Router } from 'express';
import rateLimit from 'express-rate-limit';
import { verifyWebhookSignature } from '../utils/webhookVerify.js';
import { logger } from '../utils/logger.js';
import { AppError } from '../middleware/errorHandler.js';
import { handleWebhook } from '../handlers/webhookHandler.js';
export const webhookRouter = Router();
// Rate limiter specifically for webhook endpoint (CodeQL requires this pattern)
const webhookLimiter = rateLimit({
    windowMs: 60 * 1000, // 1 minute
    max: 1000, // 1000 requests per minute per IP
    standardHeaders: true,
    legacyHeaders: false,
    keyGenerator: (req) => req.ip ?? 'unknown',
    handler: (_req, res) => {
        res.status(429).json({
            error: 'Too Many Requests',
            message: 'Rate limit exceeded. Please try again later.',
        });
    },
});
// Signature verification middleware
const verifySignature = (req, _res, next) => {
    const signature = req.headers['x-hub-signature-256'];
    const body = req.body;
    if (!verifyWebhookSignature(body, signature)) {
        logger.warn({
            ip: req.ip,
            event: req.headers['x-github-event'],
        }, 'Invalid webhook signature');
        throw new AppError('Invalid signature', 401);
    }
    next();
};
// Main webhook endpoint with rate limiting
webhookRouter.post('/', webhookLimiter, verifySignature, async (req, res) => {
    const event = req.headers['x-github-event'];
    const deliveryId = req.headers['x-github-delivery'];
    const payload = JSON.parse(req.body.toString());
    logger.info({
        event,
        deliveryId,
        action: payload.action,
        repository: payload.repository?.full_name,
    }, 'Webhook received');
    try {
        await handleWebhook(event, payload);
        res.status(200).json({ received: true });
    }
    catch (error) {
        logger.error({ error, event, deliveryId }, 'Webhook processing failed');
        res.status(500).json({ error: 'Processing failed' });
    }
});
//# sourceMappingURL=webhook.js.map