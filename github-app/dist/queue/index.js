/**
 * Queue abstraction layer.
 * Uses BullMQ with Redis when available, falls back to in-memory queue.
 */
import { logger } from '../utils/logger.js';
// Determine which queue to use based on environment
const useRedis = !!process.env.REDIS_URL;
let queueInstance = null;
let queueInitialized = false;
/**
 * Initialize the job queue
 */
export async function initializeQueue() {
    if (queueInitialized && queueInstance) {
        return queueInstance;
    }
    if (useRedis) {
        try {
            logger.info('Initializing BullMQ with Redis...');
            const { initializeBullQueue, bullJobQueue } = await import('./bullQueue.js');
            await initializeBullQueue();
            queueInstance = bullJobQueue;
            logger.info('BullMQ queue initialized successfully');
        }
        catch (error) {
            logger.warn({ error }, 'Failed to initialize BullMQ, falling back to in-memory queue');
            const { jobQueue } = await import('./jobQueue.js');
            queueInstance = createInMemoryAdapter(jobQueue);
        }
    }
    else {
        logger.info('No REDIS_URL set, using in-memory queue');
        const { jobQueue } = await import('./jobQueue.js');
        queueInstance = createInMemoryAdapter(jobQueue);
    }
    queueInitialized = true;
    return queueInstance;
}
/**
 * Get the job queue instance
 */
export async function getQueue() {
    if (!queueInstance) {
        return initializeQueue();
    }
    return queueInstance;
}
/**
 * Create an adapter for the in-memory queue to match the async interface
 */
function createInMemoryAdapter(memQueue) {
    return {
        async add(jobSpec) {
            return memQueue.add({ type: jobSpec.type, data: jobSpec.data });
        },
        async getJob(id) {
            return memQueue.getJob(id);
        },
        async getStats() {
            return memQueue.getStats();
        },
        async listJobs() {
            return memQueue.listJobs();
        },
        async cancelJob(id) {
            return memQueue.cancelJob(id);
        },
        async cleanup(maxAge = 3600000) {
            return memQueue.cleanup(maxAge);
        },
    };
}
// Legacy export for backwards compatibility
export { jobQueue } from './jobQueue.js';
//# sourceMappingURL=index.js.map