# Iteracion 13: GitHub App Setup

## Objetivos

- Crear proyecto Node.js/TypeScript
- Configurar Express server
- Setup de autenticacion GitHub App
- Estructura base del proyecto

## Tiempo Estimado: 6 horas

---

## Commit 13.1: Inicializar proyecto Node.js

**Mensaje de commit:**
```
feat(github-app): initialize node.js project

- Create package.json with dependencies
- Add TypeScript configuration
- Setup project structure
```

### `github-app/package.json`

```json
{
  "name": "@ai-toolkit/github-app",
  "version": "1.0.0",
  "description": "GitHub App for AI-powered code reviews",
  "main": "dist/index.js",
  "scripts": {
    "build": "tsc",
    "start": "node dist/index.js",
    "dev": "tsx watch src/index.ts",
    "test": "vitest",
    "test:coverage": "vitest run --coverage",
    "lint": "eslint src --ext .ts",
    "lint:fix": "eslint src --ext .ts --fix",
    "typecheck": "tsc --noEmit"
  },
  "dependencies": {
    "@octokit/rest": "^20.0.2",
    "@octokit/webhooks": "^12.0.10",
    "@octokit/auth-app": "^6.0.3",
    "express": "^4.18.2",
    "zod": "^3.22.4",
    "pino": "^8.17.2",
    "pino-pretty": "^10.3.1",
    "dotenv": "^16.3.1",
    "jsonwebtoken": "^9.0.2"
  },
  "devDependencies": {
    "@types/express": "^4.17.21",
    "@types/node": "^20.10.6",
    "@types/jsonwebtoken": "^9.0.5",
    "typescript": "^5.3.3",
    "tsx": "^4.7.0",
    "vitest": "^1.1.3",
    "@vitest/coverage-v8": "^1.1.3",
    "eslint": "^8.56.0",
    "@typescript-eslint/eslint-plugin": "^6.17.0",
    "@typescript-eslint/parser": "^6.17.0"
  },
  "engines": {
    "node": ">=20.0.0"
  }
}
```

### `github-app/tsconfig.json`

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "lib": ["ES2022"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "resolveJsonModule": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedIndexedAccess": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

### Estructura de directorios:

```
github-app/
├── src/
│   ├── index.ts
│   ├── config/
│   ├── routes/
│   ├── services/
│   ├── handlers/
│   └── utils/
├── package.json
└── tsconfig.json
```

---

## Commit 13.2: Crear servidor Express base

**Mensaje de commit:**
```
feat(github-app): add express server

- Create main server file
- Setup middleware
- Add health check endpoint
- Configure error handling
```

### `github-app/src/index.ts`

```typescript
import express from 'express';
import { config } from './config/index.js';
import { logger } from './utils/logger.js';
import { errorHandler } from './middleware/errorHandler.js';
import { requestLogger } from './middleware/requestLogger.js';
import { webhookRouter } from './routes/webhook.js';
import { healthRouter } from './routes/health.js';

const app = express();

// Trust proxy for accurate IP logging
app.set('trust proxy', 1);

// Raw body for webhook signature verification
app.use('/webhook', express.raw({ type: 'application/json' }));

// JSON parsing for other routes
app.use(express.json());

// Request logging
app.use(requestLogger);

// Routes
app.use('/health', healthRouter);
app.use('/webhook', webhookRouter);

// Error handling
app.use(errorHandler);

// Start server
const server = app.listen(config.port, () => {
  logger.info({ port: config.port }, 'Server started');
});

// Graceful shutdown
const shutdown = async () => {
  logger.info('Shutting down gracefully...');
  server.close(() => {
    logger.info('Server closed');
    process.exit(0);
  });

  // Force close after 10s
  setTimeout(() => {
    logger.error('Forced shutdown');
    process.exit(1);
  }, 10000);
};

process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);

export { app };
```

### `github-app/src/routes/health.ts`

```typescript
import { Router } from 'express';

export const healthRouter = Router();

healthRouter.get('/', (req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    version: process.env.npm_package_version || '1.0.0',
  });
});

healthRouter.get('/ready', async (req, res) => {
  // Check dependencies are ready
  const checks = {
    github: true, // Will be implemented later
    ollama: true, // Will be implemented later
  };

  const allHealthy = Object.values(checks).every(Boolean);

  res.status(allHealthy ? 200 : 503).json({
    status: allHealthy ? 'ready' : 'not_ready',
    checks,
  });
});
```

---

## Commit 13.3: Agregar sistema de logging

**Mensaje de commit:**
```
feat(github-app): add pino logger

- Configure structured logging
- Add request logging middleware
- Support log levels from env
```

### `github-app/src/utils/logger.ts`

```typescript
import pino from 'pino';
import { config } from '../config/index.js';

export const logger = pino({
  level: config.logLevel,
  transport: config.isDevelopment
    ? {
        target: 'pino-pretty',
        options: {
          colorize: true,
          translateTime: 'SYS:standard',
          ignore: 'pid,hostname',
        },
      }
    : undefined,
  base: {
    env: config.nodeEnv,
  },
  redact: ['req.headers.authorization', 'req.headers["x-hub-signature-256"]'],
});

export type Logger = typeof logger;
```

### `github-app/src/middleware/requestLogger.ts`

```typescript
import { Request, Response, NextFunction } from 'express';
import { logger } from '../utils/logger.js';

export const requestLogger = (req: Request, res: Response, next: NextFunction) => {
  const start = Date.now();

  res.on('finish', () => {
    const duration = Date.now() - start;

    logger.info({
      method: req.method,
      url: req.url,
      status: res.statusCode,
      duration,
      ip: req.ip,
      userAgent: req.get('user-agent'),
    }, 'Request completed');
  });

  next();
};
```

### `github-app/src/middleware/errorHandler.ts`

```typescript
import { Request, Response, NextFunction } from 'express';
import { logger } from '../utils/logger.js';
import { ZodError } from 'zod';

export class AppError extends Error {
  constructor(
    public message: string,
    public statusCode: number = 500,
    public code?: string
  ) {
    super(message);
    this.name = 'AppError';
  }
}

export const errorHandler = (
  err: Error,
  req: Request,
  res: Response,
  next: NextFunction
) => {
  logger.error({
    err,
    method: req.method,
    url: req.url,
  }, 'Request error');

  if (err instanceof AppError) {
    return res.status(err.statusCode).json({
      error: err.message,
      code: err.code,
    });
  }

  if (err instanceof ZodError) {
    return res.status(400).json({
      error: 'Validation error',
      details: err.errors,
    });
  }

  res.status(500).json({
    error: 'Internal server error',
  });
};
```

---

## Commit 13.4: Configurar autenticacion GitHub App

**Mensaje de commit:**
```
feat(github-app): add github authentication

- Setup @octokit/auth-app
- JWT generation for app auth
- Installation token management
- Webhook signature verification
```

### `github-app/src/services/github.ts`

```typescript
import { Octokit } from '@octokit/rest';
import { createAppAuth } from '@octokit/auth-app';
import { config } from '../config/index.js';
import { logger } from '../utils/logger.js';

// Octokit instance cache per installation
const octokitCache = new Map<number, Octokit>();

/**
 * Get an authenticated Octokit instance for an installation.
 */
export async function getOctokit(installationId: number): Promise<Octokit> {
  if (octokitCache.has(installationId)) {
    return octokitCache.get(installationId)!;
  }

  const octokit = new Octokit({
    authStrategy: createAppAuth,
    auth: {
      appId: config.github.appId,
      privateKey: config.github.privateKey,
      installationId,
    },
    log: {
      debug: (msg) => logger.debug(msg),
      info: (msg) => logger.info(msg),
      warn: (msg) => logger.warn(msg),
      error: (msg) => logger.error(msg),
    },
  });

  octokitCache.set(installationId, octokit);
  return octokit;
}

/**
 * Clear cached Octokit instance for an installation.
 */
export function clearOctokitCache(installationId: number): void {
  octokitCache.delete(installationId);
}

/**
 * Get the app's Octokit instance (not installation-specific).
 */
export function getAppOctokit(): Octokit {
  return new Octokit({
    authStrategy: createAppAuth,
    auth: {
      appId: config.github.appId,
      privateKey: config.github.privateKey,
    },
  });
}
```

### `github-app/src/utils/webhookVerify.ts`

```typescript
import crypto from 'crypto';
import { config } from '../config/index.js';

/**
 * Verify the webhook signature from GitHub.
 */
export function verifyWebhookSignature(
  payload: Buffer | string,
  signature: string | undefined
): boolean {
  if (!signature) {
    return false;
  }

  const sig = Buffer.from(signature);
  const body = typeof payload === 'string' ? payload : payload.toString();

  const hmac = crypto.createHmac('sha256', config.github.webhookSecret);
  const digest = Buffer.from('sha256=' + hmac.update(body).digest('hex'));

  if (sig.length !== digest.length) {
    return false;
  }

  return crypto.timingSafeEqual(digest, sig);
}
```

---

## Commit 13.5: Crear rutas de webhook

**Mensaje de commit:**
```
feat(github-app): add webhook routes

- Setup webhook endpoint
- Verify signature middleware
- Route to appropriate handlers
```

### `github-app/src/routes/webhook.ts`

```typescript
import { Router, Request, Response, NextFunction } from 'express';
import { verifyWebhookSignature } from '../utils/webhookVerify.js';
import { logger } from '../utils/logger.js';
import { AppError } from '../middleware/errorHandler.js';
import { handleWebhook } from '../handlers/webhookHandler.js';

export const webhookRouter = Router();

// Signature verification middleware
const verifySignature = (req: Request, res: Response, next: NextFunction) => {
  const signature = req.headers['x-hub-signature-256'] as string | undefined;
  const body = req.body as Buffer;

  if (!verifyWebhookSignature(body, signature)) {
    logger.warn({
      ip: req.ip,
      event: req.headers['x-github-event'],
    }, 'Invalid webhook signature');
    throw new AppError('Invalid signature', 401);
  }

  next();
};

// Main webhook endpoint
webhookRouter.post('/', verifySignature, async (req: Request, res: Response) => {
  const event = req.headers['x-github-event'] as string;
  const deliveryId = req.headers['x-github-delivery'] as string;
  const payload = JSON.parse((req.body as Buffer).toString());

  logger.info({
    event,
    deliveryId,
    action: payload.action,
    repository: payload.repository?.full_name,
  }, 'Webhook received');

  try {
    await handleWebhook(event, payload);
    res.status(200).json({ received: true });
  } catch (error) {
    logger.error({ error, event, deliveryId }, 'Webhook processing failed');
    res.status(500).json({ error: 'Processing failed' });
  }
});
```

### `github-app/src/handlers/webhookHandler.ts`

```typescript
import { logger } from '../utils/logger.js';
import { handlePullRequest } from './pullRequestHandler.js';
import { handleInstallation } from './installationHandler.js';

export type WebhookPayload = Record<string, unknown>;

/**
 * Route webhook events to appropriate handlers.
 */
export async function handleWebhook(
  event: string,
  payload: WebhookPayload
): Promise<void> {
  const action = payload.action as string | undefined;

  logger.debug({ event, action }, 'Processing webhook');

  switch (event) {
    case 'pull_request':
      await handlePullRequest(action, payload);
      break;

    case 'pull_request_review_comment':
      // Handle review comments
      break;

    case 'installation':
    case 'installation_repositories':
      await handleInstallation(action, payload);
      break;

    case 'ping':
      logger.info('Ping event received');
      break;

    default:
      logger.debug({ event }, 'Unhandled event type');
  }
}
```

---

## Commit 13.6: Tests del servidor

**Mensaje de commit:**
```
test(github-app): add server tests

- Test health endpoints
- Test webhook signature verification
- Test error handling
```

### `github-app/src/__tests__/health.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import request from 'supertest';
import { app } from '../index.js';

describe('Health endpoints', () => {
  it('GET /health returns ok status', async () => {
    const response = await request(app).get('/health');

    expect(response.status).toBe(200);
    expect(response.body.status).toBe('ok');
    expect(response.body.timestamp).toBeDefined();
  });

  it('GET /health/ready returns ready status', async () => {
    const response = await request(app).get('/health/ready');

    expect(response.status).toBe(200);
    expect(response.body.status).toBe('ready');
    expect(response.body.checks).toBeDefined();
  });
});
```

### `github-app/src/__tests__/webhookVerify.test.ts`

```typescript
import { describe, it, expect, vi } from 'vitest';
import crypto from 'crypto';
import { verifyWebhookSignature } from '../utils/webhookVerify.js';

// Mock config
vi.mock('../config/index.js', () => ({
  config: {
    github: {
      webhookSecret: 'test-secret',
    },
  },
}));

describe('Webhook signature verification', () => {
  const secret = 'test-secret';

  function generateSignature(payload: string): string {
    const hmac = crypto.createHmac('sha256', secret);
    return 'sha256=' + hmac.update(payload).digest('hex');
  }

  it('accepts valid signature', () => {
    const payload = '{"test": "data"}';
    const signature = generateSignature(payload);

    expect(verifyWebhookSignature(payload, signature)).toBe(true);
  });

  it('rejects invalid signature', () => {
    const payload = '{"test": "data"}';

    expect(verifyWebhookSignature(payload, 'sha256=invalid')).toBe(false);
  });

  it('rejects missing signature', () => {
    const payload = '{"test": "data"}';

    expect(verifyWebhookSignature(payload, undefined)).toBe(false);
  });

  it('handles buffer payload', () => {
    const payload = Buffer.from('{"test": "data"}');
    const signature = generateSignature(payload.toString());

    expect(verifyWebhookSignature(payload, signature)).toBe(true);
  });
});
```

---

## Resumen de la Iteracion 13

### Commits:
1. `feat(github-app): initialize node.js project`
2. `feat(github-app): add express server`
3. `feat(github-app): add pino logger`
4. `feat(github-app): add github authentication`
5. `feat(github-app): add webhook routes`
6. `test(github-app): add server tests`

### Archivos:
```
github-app/
├── package.json
├── tsconfig.json
├── src/
│   ├── index.ts
│   ├── config/
│   │   └── index.ts
│   ├── routes/
│   │   ├── health.ts
│   │   └── webhook.ts
│   ├── handlers/
│   │   └── webhookHandler.ts
│   ├── services/
│   │   └── github.ts
│   ├── middleware/
│   │   ├── errorHandler.ts
│   │   └── requestLogger.ts
│   ├── utils/
│   │   ├── logger.ts
│   │   └── webhookVerify.ts
│   └── __tests__/
│       ├── health.test.ts
│       └── webhookVerify.test.ts
└── .env.example
```

---

## Siguiente Iteracion

Continua con: **[14-GITHUB-APP-CONFIG.md](14-GITHUB-APP-CONFIG.md)**
