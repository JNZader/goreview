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
declare class JobQueue {
    private jobs;
    private processing;
    private concurrency;
    private maxRetries;
    private retryDelays;
    /**
     * Add a job to the queue.
     */
    add(jobSpec: Omit<Job, 'id' | 'status' | 'attempts' | 'maxAttempts' | 'createdAt'>): Promise<string>;
    /**
     * Get job status by ID.
     */
    getJob(id: string): Job | undefined;
    /**
     * Get queue statistics.
     */
    getStats(): {
        total: number;
        pending: number;
        processing: number;
        completed: number;
        failed: number;
    };
    /**
     * List all jobs, sorted by creation date (newest first).
     */
    listJobs(): Job[];
    /**
     * Cancel a pending job.
     * Returns true if job was cancelled, false if not found or not cancellable.
     */
    cancelJob(id: string): boolean;
    /**
     * Process pending jobs.
     */
    private processQueue;
    /**
     * Process a single job.
     */
    private processJob;
    /**
     * Execute job based on type.
     */
    private executeJob;
    /**
     * Generate unique job ID using cryptographically secure random.
     */
    private generateId;
    /**
     * Clean up old completed/failed jobs.
     */
    cleanup(maxAge?: number): number;
}
export declare const jobQueue: JobQueue;
export {};
//# sourceMappingURL=jobQueue.d.ts.map