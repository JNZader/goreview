/**
 * Queue abstraction layer.
 * Uses BullMQ with Redis when available, falls back to in-memory queue.
 */
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
/**
 * Initialize the job queue
 */
export declare function initializeQueue(): Promise<JobQueue>;
/**
 * Get the job queue instance
 */
export declare function getQueue(): Promise<JobQueue>;
export type { PRReviewJobData, JobResult } from './bullQueue.js';
export { jobQueue } from './jobQueue.js';
//# sourceMappingURL=index.d.ts.map