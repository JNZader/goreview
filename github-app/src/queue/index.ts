/**
 * Queue abstraction layer.
 * Uses BullMQ with Redis when available, falls back to in-memory queue.
 */

import { logger } from '../utils/logger.js';

// Determine which queue to use based on environment
const useRedis = !!process.env.REDIS_URL;

export interface QueueJob {
  id: string;
  type: 'pr_review';
  data: {
    installationId: number;
    owner: string;
    repo: string;
    pullNumber: number;
    headSha: string;
  };
  status: 'pending' | 'processing' | 'completed' | 'failed';
  attempts: number;
  maxAttempts: number;
  createdAt: Date;
  processedAt?: Date;
  error?: string;
}

export interface QueueStats {
  total: number;
  pending: number;
  processing: number;
  completed: number;
  failed: number;
}

export interface JobQueue {
  add(jobSpec: Omit<QueueJob, 'id' | 'status' | 'attempts' | 'maxAttempts' | 'createdAt'>): Promise<string>;
  getJob(id: string): Promise<QueueJob | undefined>;
  getStats(): Promise<QueueStats>;
  listJobs(): Promise<QueueJob[]>;
  cancelJob(id: string): Promise<boolean>;
  cleanup(maxAge?: number): Promise<number>;
}

let queueInstance: JobQueue | null = null;
let queueInitialized = false;

/**
 * Initialize the job queue
 */
export async function initializeQueue(): Promise<JobQueue> {
  if (queueInitialized && queueInstance) {
    return queueInstance;
  }

  if (useRedis) {
    try {
      logger.info('Initializing BullMQ with Redis...');
      const { initializeBullQueue, bullJobQueue } = await import('./bullQueue.js');
      await initializeBullQueue();
      queueInstance = bullJobQueue as unknown as JobQueue;
      logger.info('BullMQ queue initialized successfully');
    } catch (error) {
      logger.warn({ error }, 'Failed to initialize BullMQ, falling back to in-memory queue');
      const { jobQueue } = await import('./jobQueue.js');
      queueInstance = createInMemoryAdapter(jobQueue);
    }
  } else {
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
export async function getQueue(): Promise<JobQueue> {
  if (!queueInstance) {
    return initializeQueue();
  }
  return queueInstance;
}

// In-memory queue interface (sync operations)
interface InMemoryQueue {
  add(jobSpec: { type: 'pr_review'; data: QueueJob['data'] }): Promise<string>;
  getJob(id: string): QueueJob | undefined;
  getStats(): QueueStats;
  listJobs(): QueueJob[];
  cancelJob(id: string): boolean;
  cleanup(maxAge?: number): number;
}

/**
 * Create an adapter for the in-memory queue to match the async interface
 */
function createInMemoryAdapter(memQueue: InMemoryQueue): JobQueue {
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

// Re-export types
export type { PRReviewJobData, JobResult } from './bullQueue.js';

// Legacy export for backwards compatibility
export { jobQueue } from './jobQueue.js';
