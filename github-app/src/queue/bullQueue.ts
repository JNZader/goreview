/**
 * BullMQ-based persistent job queue with Redis backend.
 * Provides job persistence, automatic retries, and horizontal scaling.
 */

import { Queue, Worker, Job as BullJob, QueueEvents, ConnectionOptions } from 'bullmq';
import { logger } from '../utils/logger.js';

// Job data interface
export interface PRReviewJobData {
  installationId: number;
  owner: string;
  repo: string;
  pullNumber: number;
  headSha: string;
}

export interface JobResult {
  success: boolean;
  reviewId?: string;
  error?: string;
  duration?: number;
}

// Queue names
const QUEUE_NAME = 'pr-reviews';

// Redis connection config (supports Upstash Redis and standard Redis)
function getRedisConfig(): ConnectionOptions {
  const redisUrl = process.env.REDIS_URL || 'redis://localhost:6379';

  try {
    const url = new URL(redisUrl);
    const useTLS = url.protocol === 'rediss:';

    return {
      host: url.hostname,
      port: parseInt(url.port) || 6379,
      password: url.password || undefined,
      username: url.username || undefined,
      // Upstash Redis requires TLS with specific options
      tls: useTLS ? {
        rejectUnauthorized: true,
      } : undefined,
      // Upstash compatibility: enable offline queue for serverless
      enableOfflineQueue: true,
      maxRetriesPerRequest: null, // Required for BullMQ with Upstash
    };
  } catch {
    // Fallback for simple host:port format
    return {
      host: 'localhost',
      port: 6379,
    };
  }
}

/**
 * BullMQ Queue Manager for PR Reviews
 */
export class BullQueueManager {
  private readonly queue: Queue<PRReviewJobData, JobResult>;
  private worker: Worker<PRReviewJobData, JobResult> | null = null;
  private queueEvents: QueueEvents | null = null;
  private isInitialized = false;
  private readonly connection: ConnectionOptions;

  constructor() {
    this.connection = getRedisConfig();

    this.queue = new Queue<PRReviewJobData, JobResult>(QUEUE_NAME, {
      connection: this.connection,
      defaultJobOptions: {
        attempts: 3,
        backoff: {
          type: 'exponential',
          delay: 1000, // Initial delay: 1 second
        },
        removeOnComplete: {
          age: 3600, // Keep completed jobs for 1 hour
          count: 1000, // Keep last 1000 completed jobs
        },
        removeOnFail: {
          age: 86400, // Keep failed jobs for 24 hours
        },
      },
    });

    logger.info({ queue: QUEUE_NAME }, 'BullMQ queue initialized');
  }

  /**
   * Start the worker to process jobs
   */
  async startWorker(): Promise<void> {
    if (this.worker) {
      logger.warn('Worker already started');
      return;
    }

    this.worker = new Worker<PRReviewJobData, JobResult>(
      QUEUE_NAME,
      async (job) => this.processJob(job),
      {
        connection: this.connection,
        concurrency: parseInt(process.env.QUEUE_CONCURRENCY || '3'),
        limiter: {
          max: 10, // Max 10 jobs per minute
          duration: 60000,
        },
      }
    );

    // Event handlers
    this.worker.on('completed', (job: BullJob<PRReviewJobData, JobResult>, result: JobResult) => {
      logger.info({
        jobId: job.id,
        owner: job.data.owner,
        repo: job.data.repo,
        pullNumber: job.data.pullNumber,
        duration: result.duration,
      }, 'Job completed successfully');
    });

    this.worker.on('failed', (job: BullJob<PRReviewJobData, JobResult> | undefined, error: Error) => {
      logger.error({
        jobId: job?.id,
        error: error.message,
        attempts: job?.attemptsMade,
      }, 'Job failed');
    });

    this.worker.on('error', (error: Error) => {
      logger.error({ error }, 'Worker error');
    });

    this.worker.on('stalled', (jobId: string) => {
      logger.warn({ jobId }, 'Job stalled');
    });

    // Queue events for monitoring
    this.queueEvents = new QueueEvents(QUEUE_NAME, {
      connection: this.connection,
    });

    this.queueEvents.on('waiting', ({ jobId }: { jobId: string }) => {
      logger.debug({ jobId }, 'Job waiting');
    });

    this.queueEvents.on('active', ({ jobId }: { jobId: string }) => {
      logger.debug({ jobId }, 'Job active');
    });

    this.queueEvents.on('progress', ({ jobId, data }) => {
      logger.debug({ jobId, progress: data }, 'Job progress');
    });

    this.isInitialized = true;
    logger.info({ concurrency: this.worker.opts.concurrency }, 'BullMQ worker started');
  }

  /**
   * Process a single job
   */
  private async processJob(job: BullJob<PRReviewJobData, JobResult>): Promise<JobResult> {
    const startTime = Date.now();

    logger.info({
      jobId: job.id,
      attempt: job.attemptsMade + 1,
      owner: job.data.owner,
      repo: job.data.repo,
      pullNumber: job.data.pullNumber,
    }, 'Processing PR review job');

    try {
      // Update progress
      await job.updateProgress(10);

      // Dynamic import to avoid circular dependency
      const { processReviewJob } = await import('../handlers/pullRequestHandler.js');

      await job.updateProgress(20);

      await processReviewJob(
        job.data.installationId,
        job.data.owner,
        job.data.repo,
        job.data.pullNumber,
        job.data.headSha
      );

      await job.updateProgress(100);

      const duration = Date.now() - startTime;

      return {
        success: true,
        duration,
      };

    } catch (error) {
      const err = error as Error;

      // Check if this is a retryable error
      if (this.isRetryableError(err)) {
        throw error; // Let BullMQ handle retry
      }

      // Non-retryable error
      logger.error({
        jobId: job.id,
        error: err.message,
      }, 'Non-retryable error, marking as failed');

      return {
        success: false,
        error: err.message,
        duration: Date.now() - startTime,
      };
    }
  }

  /**
   * Check if an error should trigger a retry
   */
  private isRetryableError(error: Error): boolean {
    const retryablePatterns = [
      'ECONNRESET',
      'ETIMEDOUT',
      'ECONNREFUSED',
      'rate limit',
      'too many requests',
      '503',
      '502',
      '500',
    ];

    const message = error.message.toLowerCase();
    return retryablePatterns.some(pattern => message.includes(pattern.toLowerCase()));
  }

  /**
   * Add a new PR review job
   */
  async addJob(data: PRReviewJobData, options?: {
    priority?: number;
    delay?: number;
    jobId?: string;
  }): Promise<string> {
    const job = await this.queue.add(
      'pr-review',
      data,
      {
        priority: options?.priority ?? this.calculatePriority(data),
        delay: options?.delay,
        jobId: options?.jobId ?? this.generateJobId(data),
        // Deduplicate: same PR won't be reviewed twice simultaneously
      }
    );

    logger.info({
      jobId: job.id,
      owner: data.owner,
      repo: data.repo,
      pullNumber: data.pullNumber,
    }, 'PR review job added');

    return job.id!;
  }

  /**
   * Calculate job priority (lower = higher priority)
   */
  private calculatePriority(data: PRReviewJobData): number {
    // Could be enhanced with:
    // - Smaller PRs get higher priority
    // - Certain repos get higher priority
    // - Active reviewers get higher priority
    return 10; // Default priority
  }

  /**
   * Generate a unique job ID that prevents duplicates
   */
  private generateJobId(data: PRReviewJobData): string {
    return `pr-${data.owner}-${data.repo}-${data.pullNumber}-${data.headSha}`;
  }

  /**
   * Get job by ID
   */
  async getJob(jobId: string): Promise<BullJob<PRReviewJobData, JobResult> | undefined> {
    return await this.queue.getJob(jobId);
  }

  /**
   * Get queue statistics
   */
  async getStats(): Promise<{
    waiting: number;
    active: number;
    completed: number;
    failed: number;
    delayed: number;
    paused: number;
  }> {
    const counts = await this.queue.getJobCounts();
    return {
      waiting: counts.waiting ?? 0,
      active: counts.active ?? 0,
      completed: counts.completed ?? 0,
      failed: counts.failed ?? 0,
      delayed: counts.delayed ?? 0,
      paused: counts.paused ?? 0,
    };
  }

  /**
   * List jobs with pagination
   */
  async listJobs(
    status: 'waiting' | 'active' | 'completed' | 'failed' | 'delayed',
    start = 0,
    end = 20
  ): Promise<BullJob<PRReviewJobData, JobResult>[]> {
    switch (status) {
      case 'waiting':
        return this.queue.getWaiting(start, end);
      case 'active':
        return this.queue.getActive(start, end);
      case 'completed':
        return this.queue.getCompleted(start, end);
      case 'failed':
        return this.queue.getFailed(start, end);
      case 'delayed':
        return this.queue.getDelayed(start, end);
      default:
        return [];
    }
  }

  /**
   * Cancel a job
   */
  async cancelJob(jobId: string): Promise<boolean> {
    const job = await this.queue.getJob(jobId);
    if (!job) {
      return false;
    }

    const state = await job.getState();
    if (state === 'waiting' || state === 'delayed') {
      await job.remove();
      logger.info({ jobId }, 'Job cancelled');
      return true;
    }

    return false;
  }

  /**
   * Retry a failed job
   */
  async retryJob(jobId: string): Promise<boolean> {
    const job = await this.queue.getJob(jobId);
    if (!job) {
      return false;
    }

    const state = await job.getState();
    if (state === 'failed') {
      await job.retry();
      logger.info({ jobId }, 'Job retried');
      return true;
    }

    return false;
  }

  /**
   * Pause the queue
   */
  async pause(): Promise<void> {
    await this.queue.pause();
    logger.info('Queue paused');
  }

  /**
   * Resume the queue
   */
  async resume(): Promise<void> {
    await this.queue.resume();
    logger.info('Queue resumed');
  }

  /**
   * Clean old jobs
   */
  async clean(grace: number = 3600000, type: 'completed' | 'failed' = 'completed'): Promise<string[]> {
    const removed = await this.queue.clean(grace, 1000, type);
    logger.info({ removed: removed.length, type }, 'Cleaned old jobs');
    return removed;
  }

  /**
   * Close the queue and worker
   */
  async close(): Promise<void> {
    if (this.worker) {
      await this.worker.close();
      this.worker = null;
    }

    if (this.queueEvents) {
      await this.queueEvents.close();
      this.queueEvents = null;
    }

    await this.queue.close();
    this.isInitialized = false;

    logger.info('BullMQ queue closed');
  }

  /**
   * Check if Redis is connected
   */
  async healthCheck(): Promise<boolean> {
    try {
      await this.queue.getJobCounts();
      return true;
    } catch {
      return false;
    }
  }
}

// Singleton instance
let bullQueueManager: BullQueueManager | null = null;

/**
 * Get the BullMQ queue manager instance
 */
export function getBullQueue(): BullQueueManager {
  if (!bullQueueManager) {
    bullQueueManager = new BullQueueManager();
  }
  return bullQueueManager;
}

/**
 * Initialize and start the queue worker
 */
export async function initializeBullQueue(): Promise<BullQueueManager> {
  const queue = getBullQueue();
  await queue.startWorker();
  return queue;
}

/**
 * Compatibility layer for existing jobQueue interface
 */
export const bullJobQueue = {
  async add(jobSpec: {
    type: 'pr_review';
    data: PRReviewJobData;
  }): Promise<string> {
    return getBullQueue().addJob(jobSpec.data);
  },

  async getJob(id: string) {
    const job = await getBullQueue().getJob(id);
    if (!job) return undefined;

    const state = await job.getState();
    return {
      id: job.id!,
      type: 'pr_review' as const,
      data: job.data,
      status: this.mapState(state),
      attempts: job.attemptsMade,
      maxAttempts: job.opts.attempts || 3,
      createdAt: new Date(job.timestamp),
      processedAt: job.processedOn ? new Date(job.processedOn) : undefined,
      error: job.failedReason,
    };
  },

  mapState(state: string): 'pending' | 'processing' | 'completed' | 'failed' {
    switch (state) {
      case 'waiting':
      case 'delayed':
        return 'pending';
      case 'active':
        return 'processing';
      case 'completed':
        return 'completed';
      case 'failed':
        return 'failed';
      default:
        return 'pending';
    }
  },

  async getStats() {
    const stats = await getBullQueue().getStats();
    return {
      total: stats.waiting + stats.active + stats.completed + stats.failed,
      pending: stats.waiting + stats.delayed,
      processing: stats.active,
      completed: stats.completed,
      failed: stats.failed,
    };
  },

  async listJobs() {
    const queue = getBullQueue();
    const [waiting, active, completed, failed] = await Promise.all([
      queue.listJobs('waiting', 0, 50),
      queue.listJobs('active', 0, 50),
      queue.listJobs('completed', 0, 20),
      queue.listJobs('failed', 0, 20),
    ]);

    const allJobs = [...waiting, ...active, ...completed, ...failed];

    return allJobs.map(job => ({
      id: job.id!,
      type: 'pr_review' as const,
      data: job.data,
      status: 'pending' as const, // Will be properly mapped
      attempts: job.attemptsMade,
      maxAttempts: job.opts.attempts || 3,
      createdAt: new Date(job.timestamp),
    }));
  },

  async cancelJob(id: string) {
    return getBullQueue().cancelJob(id);
  },

  async cleanup(maxAge = 3600000) {
    const removed = await getBullQueue().clean(maxAge, 'completed');
    return removed.length;
  },
};
