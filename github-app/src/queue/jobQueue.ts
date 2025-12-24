import crypto from 'node:crypto';
import { logger } from '../utils/logger.js';

export interface Job {
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

/**
 * Simple in-memory job queue with retry support.
 */
class JobQueue {
  private jobs: Map<string, Job> = new Map();
  private processing: Set<string> = new Set();
  private concurrency: number = 3;
  private maxRetries: number = 3;
  private retryDelays: number[] = [1000, 5000, 15000]; // ms

  /**
   * Add a job to the queue.
   */
  async add(jobSpec: Omit<Job, 'id' | 'status' | 'attempts' | 'maxAttempts' | 'createdAt'>): Promise<string> {
    const id = this.generateId();

    const job: Job = {
      ...jobSpec,
      id,
      status: 'pending',
      attempts: 0,
      maxAttempts: this.maxRetries,
      createdAt: new Date(),
    };

    this.jobs.set(id, job);

    logger.debug({ jobId: id, type: job.type }, 'Job added to queue');

    // Trigger processing
    this.processQueue();

    return id;
  }

  /**
   * Get job status by ID.
   */
  getJob(id: string): Job | undefined {
    return this.jobs.get(id);
  }

  /**
   * Get queue statistics.
   */
  getStats(): {
    total: number;
    pending: number;
    processing: number;
    completed: number;
    failed: number;
  } {
    let pending = 0;
    let processing = 0;
    let completed = 0;
    let failed = 0;

    for (const job of this.jobs.values()) {
      switch (job.status) {
        case 'pending': pending++; break;
        case 'processing': processing++; break;
        case 'completed': completed++; break;
        case 'failed': failed++; break;
      }
    }

    return {
      total: this.jobs.size,
      pending,
      processing,
      completed,
      failed,
    };
  }

  /**
   * List all jobs, sorted by creation date (newest first).
   */
  listJobs(): Job[] {
    return Array.from(this.jobs.values())
      .sort((a, b) => b.createdAt.getTime() - a.createdAt.getTime());
  }

  /**
   * Cancel a pending job.
   * Returns true if job was cancelled, false if not found or not cancellable.
   */
  cancelJob(id: string): boolean {
    const job = this.jobs.get(id);

    if (!job) {
      return false;
    }

    // Can only cancel pending jobs
    if (job.status !== 'pending') {
      return false;
    }

    this.jobs.delete(id);
    logger.info({ jobId: id }, 'Job cancelled');

    return true;
  }

  /**
   * Process pending jobs.
   */
  private async processQueue(): Promise<void> {
    // Check concurrency limit
    if (this.processing.size >= this.concurrency) {
      return;
    }

    // Find pending jobs
    for (const [id, job] of this.jobs) {
      if (job.status === 'pending' && !this.processing.has(id)) {
        // Start processing
        this.processing.add(id);
        this.processJob(job);

        // Check concurrency again
        if (this.processing.size >= this.concurrency) {
          break;
        }
      }
    }
  }

  /**
   * Process a single job.
   */
  private async processJob(job: Job): Promise<void> {
    job.status = 'processing';
    job.attempts++;
    job.processedAt = new Date();

    logger.info({
      jobId: job.id,
      type: job.type,
      attempt: job.attempts,
    }, 'Processing job');

    try {
      await this.executeJob(job);

      job.status = 'completed';
      logger.info({ jobId: job.id }, 'Job completed successfully');

    } catch (error: unknown) {
      const err = error as Error;
      logger.error({ error, jobId: job.id, attempt: job.attempts }, 'Job failed');

      job.error = err.message;

      if (job.attempts < job.maxAttempts) {
        // Schedule retry
        const delay = this.retryDelays[job.attempts - 1] || 30000;
        job.status = 'pending';

        logger.info({
          jobId: job.id,
          nextAttempt: job.attempts + 1,
          delay,
        }, 'Scheduling job retry');

        setTimeout(() => {
          this.processQueue();
        }, delay);

      } else {
        job.status = 'failed';
        logger.error({
          jobId: job.id,
          attempts: job.attempts,
        }, 'Job failed after max retries');
      }
    } finally {
      this.processing.delete(job.id);
      this.processQueue();
    }
  }

  /**
   * Execute job based on type.
   */
  private async executeJob(job: Job): Promise<void> {
    if (job.type === 'pr_review') {
      // Dynamic import to avoid circular dependency
      const { processReviewJob } = await import('../handlers/pullRequestHandler.js');
      await processReviewJob(
        job.data.installationId,
        job.data.owner,
        job.data.repo,
        job.data.pullNumber,
        job.data.headSha
      );
      return;
    }

    throw new Error(`Unknown job type: ${job.type}`);
  }

  /**
   * Generate unique job ID using cryptographically secure random.
   */
  private generateId(): string {
    return `job_${Date.now()}_${crypto.randomBytes(6).toString('hex')}`;
  }

  /**
   * Clean up old completed/failed jobs.
   */
  cleanup(maxAge: number = 3600000): number {
    const cutoff = Date.now() - maxAge;
    let removed = 0;

    for (const [id, job] of this.jobs) {
      if (
        (job.status === 'completed' || job.status === 'failed') &&
        job.processedAt &&
        job.processedAt.getTime() < cutoff
      ) {
        this.jobs.delete(id);
        removed++;
      }
    }

    if (removed > 0) {
      logger.debug({ removed }, 'Cleaned up old jobs');
    }

    return removed;
  }
}

export const jobQueue = new JobQueue();

// Periodic cleanup
setInterval(() => {
  jobQueue.cleanup();
}, 60000);
