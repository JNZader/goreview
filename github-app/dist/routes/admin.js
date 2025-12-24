import { Router } from 'express';
import { jobQueue } from '../queue/jobQueue.js';
import { config } from '../config/index.js';
import { AppError } from '../middleware/errorHandler.js';
import { logger } from '../utils/logger.js';
import { secureCompare } from '../utils/secrets.js';
export const adminRouter = Router();
/**
 * Simple API key authentication for admin routes.
 */
const requireAuth = (req, _res, next) => {
    const authHeader = req.headers.authorization;
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
        throw new AppError('Authorization required', 401);
    }
    const token = authHeader.slice(7);
    // Use webhook secret as admin token for simplicity
    // Use timing-safe comparison to prevent timing attacks
    if (!secureCompare(token, config.github.webhookSecret)) {
        logger.warn({ ip: req.ip }, 'Invalid admin token');
        throw new AppError('Invalid token', 403);
    }
    next();
};
// Apply auth to all admin routes
adminRouter.use(requireAuth);
/**
 * GET /admin/stats - Queue statistics
 */
adminRouter.get('/stats', (_req, res) => {
    const stats = jobQueue.getStats();
    res.json({
        queue: stats,
        uptime: process.uptime(),
        memory: process.memoryUsage(),
    });
});
/**
 * GET /admin/jobs - List all jobs
 */
adminRouter.get('/jobs', (req, res) => {
    const status = req.query.status;
    const limitParam = typeof req.query.limit === 'string' ? req.query.limit : '50';
    const offsetParam = typeof req.query.offset === 'string' ? req.query.offset : '0';
    const limit = Math.min(parseInt(limitParam, 10) || 50, 100);
    const offset = parseInt(offsetParam, 10) || 0;
    const allJobs = jobQueue.listJobs();
    // Filter by status if provided
    const filtered = status
        ? allJobs.filter(job => job.status === status)
        : allJobs;
    // Paginate
    const paginated = filtered.slice(offset, offset + limit);
    res.json({
        jobs: paginated,
        total: filtered.length,
        limit,
        offset,
    });
});
/**
 * GET /admin/jobs/:id - Get specific job
 */
adminRouter.get('/jobs/:id', (req, res) => {
    const jobId = req.params.id;
    const job = jobQueue.getJob(jobId);
    if (!job) {
        throw new AppError('Job not found', 404);
    }
    res.json(job);
});
/**
 * POST /admin/cleanup - Cleanup old jobs
 */
adminRouter.post('/cleanup', (req, res) => {
    const maxAgeParam = typeof req.query.maxAge === 'string' ? req.query.maxAge : '3600000';
    const maxAge = parseInt(maxAgeParam, 10) || 3600000; // 1 hour default
    const removed = jobQueue.cleanup(maxAge);
    logger.info({ removed, maxAge }, 'Admin cleanup executed');
    res.json({
        removed,
        maxAge,
    });
});
/**
 * DELETE /admin/jobs/:id - Cancel a pending job
 */
adminRouter.delete('/jobs/:id', (req, res) => {
    const jobId = req.params.id;
    const success = jobQueue.cancelJob(jobId);
    if (!success) {
        throw new AppError('Job not found or not cancellable', 404);
    }
    logger.info({ jobId }, 'Job cancelled by admin');
    res.json({ cancelled: true });
});
//# sourceMappingURL=admin.js.map