/**
 * BullMQ-based persistent job queue with Redis backend.
 * Provides job persistence, automatic retries, and horizontal scaling.
 */
import { Job as BullJob } from 'bullmq';
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
/**
 * BullMQ Queue Manager for PR Reviews
 */
export declare class BullQueueManager {
    private queue;
    private worker;
    private queueEvents;
    private isInitialized;
    private connection;
    constructor();
    /**
     * Start the worker to process jobs
     */
    startWorker(): Promise<void>;
    /**
     * Process a single job
     */
    private processJob;
    /**
     * Check if an error should trigger a retry
     */
    private isRetryableError;
    /**
     * Add a new PR review job
     */
    addJob(data: PRReviewJobData, options?: {
        priority?: number;
        delay?: number;
        jobId?: string;
    }): Promise<string>;
    /**
     * Calculate job priority (lower = higher priority)
     */
    private calculatePriority;
    /**
     * Generate a unique job ID that prevents duplicates
     */
    private generateJobId;
    /**
     * Get job by ID
     */
    getJob(jobId: string): Promise<BullJob<PRReviewJobData, JobResult> | undefined>;
    /**
     * Get queue statistics
     */
    getStats(): Promise<{
        waiting: number;
        active: number;
        completed: number;
        failed: number;
        delayed: number;
        paused: number;
    }>;
    /**
     * List jobs with pagination
     */
    listJobs(status: 'waiting' | 'active' | 'completed' | 'failed' | 'delayed', start?: number, end?: number): Promise<BullJob<PRReviewJobData, JobResult>[]>;
    /**
     * Cancel a job
     */
    cancelJob(jobId: string): Promise<boolean>;
    /**
     * Retry a failed job
     */
    retryJob(jobId: string): Promise<boolean>;
    /**
     * Pause the queue
     */
    pause(): Promise<void>;
    /**
     * Resume the queue
     */
    resume(): Promise<void>;
    /**
     * Clean old jobs
     */
    clean(grace?: number, type?: 'completed' | 'failed'): Promise<string[]>;
    /**
     * Close the queue and worker
     */
    close(): Promise<void>;
    /**
     * Check if Redis is connected
     */
    healthCheck(): Promise<boolean>;
}
/**
 * Get the BullMQ queue manager instance
 */
export declare function getBullQueue(): BullQueueManager;
/**
 * Initialize and start the queue worker
 */
export declare function initializeBullQueue(): Promise<BullQueueManager>;
/**
 * Compatibility layer for existing jobQueue interface
 */
export declare const bullJobQueue: {
    add(jobSpec: {
        type: "pr_review";
        data: PRReviewJobData;
    }): Promise<string>;
    getJob(id: string): Promise<{
        id: string;
        type: "pr_review";
        data: PRReviewJobData;
        status: "pending" | "processing" | "completed" | "failed";
        attempts: number;
        maxAttempts: number;
        createdAt: Date;
        processedAt: Date | undefined;
        error: string;
    } | undefined>;
    mapState(state: string): "pending" | "processing" | "completed" | "failed";
    getStats(): Promise<{
        total: number;
        pending: number;
        processing: number;
        completed: number;
        failed: number;
    }>;
    listJobs(): Promise<{
        id: string;
        type: "pr_review";
        data: PRReviewJobData;
        status: "pending";
        attempts: number;
        maxAttempts: number;
        createdAt: Date;
    }[]>;
    cancelJob(id: string): Promise<boolean>;
    cleanup(maxAge?: number): Promise<number>;
};
//# sourceMappingURL=bullQueue.d.ts.map