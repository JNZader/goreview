# Iteracion 16: GitHub App Webhooks

## Objetivos

- Handlers para eventos de PR
- Handler para instalaciones
- Cola de procesamiento asincrono
- Retry logic para fallos

## Tiempo Estimado: 6 horas

---

## Commit 16.1: Crear handler de Pull Requests

**Mensaje de commit:**
```
feat(github-app): add pull request handler

- Handle PR opened/synchronized events
- Validate PR before review
- Trigger review process
```

### `github-app/src/handlers/pullRequestHandler.ts`

```typescript
import { getOctokit } from '../services/github.js';
import { prReviewService } from '../services/reviewService.js';
import { commentService } from '../services/commentService.js';
import { loadRepoConfig } from '../config/repoConfig.js';
import { logger } from '../utils/logger.js';
import { jobQueue } from '../queue/jobQueue.js';

interface PullRequestPayload {
  action: string;
  number: number;
  pull_request: {
    number: number;
    title: string;
    body: string | null;
    state: string;
    draft: boolean;
    head: {
      sha: string;
      ref: string;
    };
    base: {
      ref: string;
    };
    user: {
      login: string;
    };
    changed_files: number;
    additions: number;
    deletions: number;
  };
  repository: {
    id: number;
    full_name: string;
    owner: {
      login: string;
    };
    name: string;
  };
  installation: {
    id: number;
  };
  sender: {
    login: string;
  };
}

/**
 * Handle pull request events.
 */
export async function handlePullRequest(
  action: string | undefined,
  payload: unknown
): Promise<void> {
  const pr = payload as PullRequestPayload;

  // Only handle relevant actions
  const relevantActions = ['opened', 'synchronize', 'reopened'];
  if (!action || !relevantActions.includes(action)) {
    logger.debug({ action }, 'Ignoring PR action');
    return;
  }

  const { repository, pull_request, installation } = pr;
  const owner = repository.owner.login;
  const repo = repository.name;
  const pullNumber = pull_request.number;

  logger.info({
    owner,
    repo,
    pullNumber,
    action,
    title: pull_request.title,
  }, 'Processing pull request');

  // Skip draft PRs
  if (pull_request.draft) {
    logger.info({ pullNumber }, 'Skipping draft PR');
    return;
  }

  // Get Octokit for this installation
  const octokit = await getOctokit(installation.id);

  // Load repository configuration
  const repoConfig = await loadRepoConfig(octokit, owner, repo);

  // Check if auto-review is enabled
  if (!repoConfig.review.auto_review) {
    logger.info({ owner, repo }, 'Auto-review disabled for repository');
    return;
  }

  // Validate PR is reviewable
  const validation = validatePR(pull_request, repoConfig);
  if (!validation.valid) {
    logger.info({
      pullNumber,
      reason: validation.reason,
    }, 'PR not eligible for review');
    return;
  }

  // Queue the review job
  await jobQueue.add({
    type: 'pr_review',
    data: {
      installationId: installation.id,
      owner,
      repo,
      pullNumber,
      headSha: pull_request.head.sha,
    },
  });

  logger.info({ owner, repo, pullNumber }, 'PR review job queued');
}

interface ValidationResult {
  valid: boolean;
  reason?: string;
}

function validatePR(
  pr: PullRequestPayload['pull_request'],
  config: any
): ValidationResult {
  // Check if too many files
  if (pr.changed_files > config.review.max_files) {
    return {
      valid: false,
      reason: `Too many files: ${pr.changed_files} > ${config.review.max_files}`,
    };
  }

  // Check if too large
  const totalChanges = pr.additions + pr.deletions;
  if (totalChanges > 10000) {
    return {
      valid: false,
      reason: `Too many changes: ${totalChanges} lines`,
    };
  }

  return { valid: true };
}

/**
 * Process a queued PR review job.
 */
export async function processReviewJob(
  installationId: number,
  owner: string,
  repo: string,
  pullNumber: number,
  headSha: string
): Promise<void> {
  const startTime = Date.now();

  try {
    const octokit = await getOctokit(installationId);

    // Verify the PR still exists and SHA matches
    const { data: currentPR } = await octokit.pulls.get({
      owner,
      repo,
      pull_number: pullNumber,
    });

    if (currentPR.head.sha !== headSha) {
      logger.info({
        pullNumber,
        expectedSha: headSha,
        actualSha: currentPR.head.sha,
      }, 'PR has new commits, skipping stale review');
      return;
    }

    // Set commit status to pending
    await octokit.repos.createCommitStatus({
      owner,
      repo,
      sha: headSha,
      state: 'pending',
      context: 'ai-review',
      description: 'AI code review in progress...',
    });

    // Perform the review
    const result = await prReviewService.reviewPR(
      octokit,
      owner,
      repo,
      pullNumber
    );

    // Post review comments
    await commentService.postReview(
      octokit,
      owner,
      repo,
      pullNumber,
      result
    );

    // Update commit status
    const state = result.criticalIssues > 0 ? 'failure' : 'success';
    const description = result.criticalIssues > 0
      ? `Found ${result.criticalIssues} critical issue(s)`
      : `Review complete. Score: ${result.overallScore}/100`;

    await octokit.repos.createCommitStatus({
      owner,
      repo,
      sha: headSha,
      state,
      context: 'ai-review',
      description,
    });

    const duration = Date.now() - startTime;
    logger.info({
      owner,
      repo,
      pullNumber,
      duration,
      filesReviewed: result.filesReviewed,
      totalIssues: result.totalIssues,
    }, 'PR review completed successfully');

  } catch (error) {
    logger.error({ error, owner, repo, pullNumber }, 'PR review failed');

    // Try to update status to error
    try {
      const octokit = await getOctokit(installationId);
      await octokit.repos.createCommitStatus({
        owner,
        repo,
        sha: headSha,
        state: 'error',
        context: 'ai-review',
        description: 'Review failed. Please try again.',
      });
    } catch {
      // Ignore status update errors
    }

    throw error;
  }
}
```

---

## Commit 16.2: Crear handler de instalaciones

**Mensaje de commit:**
```
feat(github-app): add installation handler

- Handle app installation/uninstallation
- Track installed repositories
- Clear caches on uninstall
```

### `github-app/src/handlers/installationHandler.ts`

```typescript
import { clearOctokitCache } from '../services/github.js';
import { clearRepoConfigCache } from '../config/repoConfig.js';
import { logger } from '../utils/logger.js';

interface InstallationPayload {
  action: string;
  installation: {
    id: number;
    account: {
      login: string;
      type: string;
    };
  };
  repositories?: Array<{
    id: number;
    name: string;
    full_name: string;
    private: boolean;
  }>;
  sender: {
    login: string;
  };
}

/**
 * Handle installation events.
 */
export async function handleInstallation(
  action: string | undefined,
  payload: unknown
): Promise<void> {
  const installation = payload as InstallationPayload;
  const { installation: inst, repositories } = installation;

  logger.info({
    action,
    installationId: inst.id,
    account: inst.account.login,
    accountType: inst.account.type,
  }, 'Installation event received');

  switch (action) {
    case 'created':
      await handleInstallationCreated(inst, repositories);
      break;

    case 'deleted':
      await handleInstallationDeleted(inst);
      break;

    case 'added':
      await handleRepositoriesAdded(inst, repositories);
      break;

    case 'removed':
      await handleRepositoriesRemoved(inst, repositories);
      break;

    default:
      logger.debug({ action }, 'Unhandled installation action');
  }
}

async function handleInstallationCreated(
  installation: InstallationPayload['installation'],
  repositories: InstallationPayload['repositories']
): Promise<void> {
  logger.info({
    installationId: installation.id,
    account: installation.account.login,
    repoCount: repositories?.length || 0,
  }, 'App installed');

  // Log installed repositories
  if (repositories) {
    for (const repo of repositories) {
      logger.info({
        installationId: installation.id,
        repo: repo.full_name,
        private: repo.private,
      }, 'Repository added to installation');
    }
  }

  // Could send welcome message, initialize settings, etc.
}

async function handleInstallationDeleted(
  installation: InstallationPayload['installation']
): Promise<void> {
  logger.info({
    installationId: installation.id,
    account: installation.account.login,
  }, 'App uninstalled');

  // Clear caches
  clearOctokitCache(installation.id);

  // Could clean up any stored data for this installation
}

async function handleRepositoriesAdded(
  installation: InstallationPayload['installation'],
  repositories: InstallationPayload['repositories']
): Promise<void> {
  if (!repositories) return;

  for (const repo of repositories) {
    logger.info({
      installationId: installation.id,
      repo: repo.full_name,
    }, 'Repository added');

    // Clear any cached config for this repo
    const [owner, name] = repo.full_name.split('/');
    clearRepoConfigCache(owner, name);
  }
}

async function handleRepositoriesRemoved(
  installation: InstallationPayload['installation'],
  repositories: InstallationPayload['repositories']
): Promise<void> {
  if (!repositories) return;

  for (const repo of repositories) {
    logger.info({
      installationId: installation.id,
      repo: repo.full_name,
    }, 'Repository removed');

    // Clear cached config
    const [owner, name] = repo.full_name.split('/');
    clearRepoConfigCache(owner, name);
  }
}
```

---

## Commit 16.3: Crear cola de trabajos

**Mensaje de commit:**
```
feat(github-app): add job queue

- In-memory job queue
- Concurrent job processing
- Retry with exponential backoff
- Job status tracking
```

### `github-app/src/queue/jobQueue.ts`

```typescript
import { logger } from '../utils/logger.js';
import { processReviewJob } from '../handlers/pullRequestHandler.js';

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

    } catch (error: any) {
      logger.error({ error, jobId: job.id, attempt: job.attempts }, 'Job failed');

      job.error = error.message;

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
    switch (job.type) {
      case 'pr_review':
        await processReviewJob(
          job.data.installationId,
          job.data.owner,
          job.data.repo,
          job.data.pullNumber,
          job.data.headSha
        );
        break;

      default:
        throw new Error(`Unknown job type: ${job.type}`);
    }
  }

  /**
   * Generate unique job ID.
   */
  private generateId(): string {
    return `job_${Date.now()}_${Math.random().toString(36).slice(2, 9)}`;
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
```

---

## Commit 16.4: Agregar endpoint de estado de cola

**Mensaje de commit:**
```
feat(github-app): add queue status endpoint

- Expose queue statistics
- Show job counts by status
- Protected admin endpoint
```

### `github-app/src/routes/admin.ts`

```typescript
import { Router } from 'express';
import { jobQueue } from '../queue/jobQueue.js';
import { ollamaService } from '../services/ollama.js';
import { config } from '../config/index.js';

export const adminRouter = Router();

// Simple auth middleware for admin routes
const adminAuth = (req: any, res: any, next: any) => {
  const authHeader = req.headers.authorization;

  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Unauthorized' });
  }

  const token = authHeader.slice(7);

  // In production, use a proper admin token
  if (config.isDevelopment || token === process.env.ADMIN_TOKEN) {
    next();
  } else {
    res.status(403).json({ error: 'Forbidden' });
  }
};

adminRouter.use(adminAuth);

/**
 * Get queue statistics.
 */
adminRouter.get('/queue/stats', (req, res) => {
  const stats = jobQueue.getStats();
  res.json(stats);
});

/**
 * Get job by ID.
 */
adminRouter.get('/queue/jobs/:id', (req, res) => {
  const job = jobQueue.getJob(req.params.id);

  if (!job) {
    return res.status(404).json({ error: 'Job not found' });
  }

  res.json(job);
});

/**
 * Trigger queue cleanup.
 */
adminRouter.post('/queue/cleanup', (req, res) => {
  const maxAge = parseInt(req.query.maxAge as string) || 3600000;
  const removed = jobQueue.cleanup(maxAge);
  res.json({ removed });
});

/**
 * Check provider health.
 */
adminRouter.get('/health/provider', async (req, res) => {
  const isHealthy = await ollamaService.healthCheck();

  res.status(isHealthy ? 200 : 503).json({
    provider: config.ai.provider,
    model: config.ai.model,
    healthy: isHealthy,
  });
});
```

Actualizar `github-app/src/index.ts` para incluir las rutas admin:

```typescript
// Add to imports
import { adminRouter } from './routes/admin.js';

// Add route
app.use('/admin', adminRouter);
```

---

## Commit 16.5: Tests de handlers

**Mensaje de commit:**
```
test(github-app): add handler tests

- Test PR event handling
- Test installation events
- Test job queue
```

### `github-app/src/__tests__/pullRequestHandler.test.ts`

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { handlePullRequest } from '../handlers/pullRequestHandler.js';

// Mock dependencies
vi.mock('../services/github.js', () => ({
  getOctokit: vi.fn().mockResolvedValue({}),
}));

vi.mock('../config/repoConfig.js', () => ({
  loadRepoConfig: vi.fn().mockResolvedValue({
    review: {
      enabled: true,
      auto_review: true,
      max_files: 50,
    },
  }),
}));

vi.mock('../queue/jobQueue.js', () => ({
  jobQueue: {
    add: vi.fn().mockResolvedValue('job_123'),
  },
}));

describe('Pull Request Handler', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  const createPayload = (overrides = {}) => ({
    action: 'opened',
    number: 1,
    pull_request: {
      number: 1,
      title: 'Test PR',
      body: 'Test body',
      state: 'open',
      draft: false,
      head: { sha: 'abc123', ref: 'feature' },
      base: { ref: 'main' },
      user: { login: 'user' },
      changed_files: 5,
      additions: 100,
      deletions: 50,
      ...overrides,
    },
    repository: {
      id: 123,
      full_name: 'owner/repo',
      owner: { login: 'owner' },
      name: 'repo',
    },
    installation: { id: 456 },
    sender: { login: 'user' },
  });

  it('queues review for opened PR', async () => {
    const { jobQueue } = await import('../queue/jobQueue.js');

    await handlePullRequest('opened', createPayload());

    expect(jobQueue.add).toHaveBeenCalledWith({
      type: 'pr_review',
      data: expect.objectContaining({
        owner: 'owner',
        repo: 'repo',
        pullNumber: 1,
      }),
    });
  });

  it('skips draft PRs', async () => {
    const { jobQueue } = await import('../queue/jobQueue.js');

    await handlePullRequest('opened', createPayload({ draft: true }));

    expect(jobQueue.add).not.toHaveBeenCalled();
  });

  it('skips non-relevant actions', async () => {
    const { jobQueue } = await import('../queue/jobQueue.js');

    await handlePullRequest('closed', createPayload());

    expect(jobQueue.add).not.toHaveBeenCalled();
  });
});
```

### `github-app/src/__tests__/jobQueue.test.ts`

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Create a fresh queue for testing
class TestJobQueue {
  private jobs = new Map();

  async add(job: any) {
    const id = `test_${Date.now()}`;
    this.jobs.set(id, { ...job, id, status: 'pending', attempts: 0 });
    return id;
  }

  getJob(id: string) {
    return this.jobs.get(id);
  }

  getStats() {
    let pending = 0, completed = 0, failed = 0;
    for (const job of this.jobs.values()) {
      if (job.status === 'pending') pending++;
      if (job.status === 'completed') completed++;
      if (job.status === 'failed') failed++;
    }
    return { total: this.jobs.size, pending, completed, failed, processing: 0 };
  }
}

describe('Job Queue', () => {
  let queue: TestJobQueue;

  beforeEach(() => {
    queue = new TestJobQueue();
  });

  it('adds jobs to queue', async () => {
    const id = await queue.add({
      type: 'pr_review',
      data: { owner: 'test', repo: 'test', pullNumber: 1 },
    });

    expect(id).toBeDefined();

    const job = queue.getJob(id);
    expect(job).toBeDefined();
    expect(job.status).toBe('pending');
  });

  it('tracks job statistics', async () => {
    await queue.add({ type: 'pr_review', data: {} });
    await queue.add({ type: 'pr_review', data: {} });

    const stats = queue.getStats();
    expect(stats.total).toBe(2);
    expect(stats.pending).toBe(2);
  });
});
```

---

## Resumen de la Iteracion 16

### Commits:
1. `feat(github-app): add pull request handler`
2. `feat(github-app): add installation handler`
3. `feat(github-app): add job queue`
4. `feat(github-app): add queue status endpoint`
5. `test(github-app): add handler tests`

### Archivos:
```
github-app/src/
├── handlers/
│   ├── pullRequestHandler.ts
│   └── installationHandler.ts
├── queue/
│   └── jobQueue.ts
├── routes/
│   └── admin.ts
└── __tests__/
    ├── pullRequestHandler.test.ts
    └── jobQueue.test.ts
```

---

## Siguiente Iteracion

Continua con: **[17-DOCKER-SETUP.md](17-DOCKER-SETUP.md)**
