import { Router } from 'express';
import { getProvider } from '../services/ai/index.js';
import { logger } from '../utils/logger.js';
export const healthRouter = Router();
healthRouter.get('/', (_req, res) => {
    res.json({
        status: 'ok',
        timestamp: new Date().toISOString(),
        version: process.env.npm_package_version || '1.0.0',
        uptime: process.uptime(),
    });
});
healthRouter.get('/ready', async (req, res) => {
    const checks = {
        server: true,
        ai_provider: false,
    };
    // Check AI provider
    try {
        const provider = getProvider();
        checks.ai_provider = await provider.healthCheck();
    }
    catch (error) {
        logger.warn({ error, requestId: req.id }, 'AI provider health check failed');
        checks.ai_provider = false;
    }
    const allHealthy = Object.values(checks).every(Boolean);
    const status = allHealthy ? 'ready' : 'degraded';
    res.status(allHealthy ? 200 : 503).json({
        status,
        timestamp: new Date().toISOString(),
        checks,
    });
});
//# sourceMappingURL=health.js.map