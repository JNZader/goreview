# Iteracion 14: GitHub App Configuration

## Objetivos

- Sistema de configuracion con Zod
- Variables de entorno y validacion
- Configuracion por repositorio
- Manejo de secretos

## Tiempo Estimado: 4 horas

---

## Commit 14.1: Crear esquema de configuracion con Zod

**Mensaje de commit:**
```
feat(github-app): add zod config schema

- Define environment variable schema
- Add validation with descriptive errors
- Support defaults and transformations
```

### `github-app/src/config/schema.ts`

```typescript
import { z } from 'zod';

// Duration parser (e.g., "1h", "30m", "60s")
const durationSchema = z.string().transform((val) => {
  const match = val.match(/^(\d+)(h|m|s)$/);
  if (!match) throw new Error(`Invalid duration: ${val}`);

  const [, num, unit] = match;
  const multipliers: Record<string, number> = {
    h: 3600000,
    m: 60000,
    s: 1000,
  };

  return parseInt(num) * multipliers[unit];
});

export const envSchema = z.object({
  // Server
  NODE_ENV: z.enum(['development', 'production', 'test']).default('development'),
  PORT: z.coerce.number().default(3000),
  LOG_LEVEL: z.enum(['debug', 'info', 'warn', 'error']).default('info'),

  // GitHub App
  GITHUB_APP_ID: z.coerce.number(),
  GITHUB_PRIVATE_KEY: z.string().transform((key) => {
    // Handle escaped newlines from env vars
    return key.replace(/\\n/g, '\n');
  }),
  GITHUB_WEBHOOK_SECRET: z.string().min(20),
  GITHUB_CLIENT_ID: z.string().optional(),
  GITHUB_CLIENT_SECRET: z.string().optional(),

  // AI Provider
  AI_PROVIDER: z.enum(['ollama', 'openai']).default('ollama'),
  AI_MODEL: z.string().default('qwen2.5-coder:14b'),
  OLLAMA_BASE_URL: z.string().url().default('http://localhost:11434'),
  OPENAI_API_KEY: z.string().optional(),

  // Rate Limiting
  RATE_LIMIT_RPS: z.coerce.number().default(10),
  RATE_LIMIT_BURST: z.coerce.number().default(20),

  // Cache
  CACHE_TTL: durationSchema.default('1h'),
  CACHE_MAX_ENTRIES: z.coerce.number().default(1000),

  // Review Settings
  REVIEW_MAX_FILES: z.coerce.number().default(50),
  REVIEW_MAX_DIFF_SIZE: z.coerce.number().default(500000), // 500KB
  REVIEW_TIMEOUT: durationSchema.default('5m'),
});

export type EnvConfig = z.infer<typeof envSchema>;
```

### `github-app/src/config/index.ts`

```typescript
import dotenv from 'dotenv';
import { envSchema } from './schema.js';
import { logger } from '../utils/logger.js';

// Load .env file
dotenv.config();

// Parse and validate environment
const result = envSchema.safeParse(process.env);

if (!result.success) {
  console.error('Configuration validation failed:');
  for (const issue of result.error.issues) {
    console.error(`  - ${issue.path.join('.')}: ${issue.message}`);
  }
  process.exit(1);
}

const env = result.data;

export const config = {
  // Server
  nodeEnv: env.NODE_ENV,
  port: env.PORT,
  logLevel: env.LOG_LEVEL,
  isDevelopment: env.NODE_ENV === 'development',
  isProduction: env.NODE_ENV === 'production',

  // GitHub
  github: {
    appId: env.GITHUB_APP_ID,
    privateKey: env.GITHUB_PRIVATE_KEY,
    webhookSecret: env.GITHUB_WEBHOOK_SECRET,
    clientId: env.GITHUB_CLIENT_ID,
    clientSecret: env.GITHUB_CLIENT_SECRET,
  },

  // AI
  ai: {
    provider: env.AI_PROVIDER,
    model: env.AI_MODEL,
    ollamaBaseUrl: env.OLLAMA_BASE_URL,
    openaiApiKey: env.OPENAI_API_KEY,
  },

  // Rate Limiting
  rateLimit: {
    rps: env.RATE_LIMIT_RPS,
    burst: env.RATE_LIMIT_BURST,
  },

  // Cache
  cache: {
    ttl: env.CACHE_TTL,
    maxEntries: env.CACHE_MAX_ENTRIES,
  },

  // Review
  review: {
    maxFiles: env.REVIEW_MAX_FILES,
    maxDiffSize: env.REVIEW_MAX_DIFF_SIZE,
    timeout: env.REVIEW_TIMEOUT,
  },
} as const;

export type Config = typeof config;
```

---

## Commit 14.2: Agregar configuracion por repositorio

**Mensaje de commit:**
```
feat(github-app): add per-repo configuration

- Load .goreview.yaml from repository
- Merge with default settings
- Validate repository config
```

### `github-app/src/config/repoConfig.ts`

```typescript
import { z } from 'zod';
import { Octokit } from '@octokit/rest';
import YAML from 'yaml';
import { logger } from '../utils/logger.js';

// Repository-level configuration schema
const repoConfigSchema = z.object({
  version: z.string().optional(),

  review: z.object({
    enabled: z.boolean().default(true),
    auto_review: z.boolean().default(true),
    max_files: z.number().default(50),
    ignore_patterns: z.array(z.string()).default([]),
    languages: z.array(z.string()).optional(),
  }).default({}),

  rules: z.object({
    preset: z.enum(['minimal', 'standard', 'strict']).default('standard'),
    enable: z.array(z.string()).default([]),
    disable: z.array(z.string()).default([]),
  }).default({}),

  comments: z.object({
    inline: z.boolean().default(true),
    summary: z.boolean().default(true),
    request_changes: z.boolean().default(false),
    min_severity: z.enum(['info', 'warning', 'error', 'critical']).default('warning'),
  }).default({}),

  labels: z.object({
    add_on_issues: z.boolean().default(true),
    critical: z.string().default('needs-attention'),
    reviewed: z.string().default('ai-reviewed'),
  }).default({}),
}).default({});

export type RepoConfig = z.infer<typeof repoConfigSchema>;

const CONFIG_FILE = '.goreview.yaml';
const CONFIG_CACHE = new Map<string, { config: RepoConfig; fetchedAt: number }>();
const CACHE_TTL = 5 * 60 * 1000; // 5 minutes

/**
 * Load configuration from a repository.
 */
export async function loadRepoConfig(
  octokit: Octokit,
  owner: string,
  repo: string,
  ref?: string
): Promise<RepoConfig> {
  const cacheKey = `${owner}/${repo}:${ref || 'default'}`;

  // Check cache
  const cached = CONFIG_CACHE.get(cacheKey);
  if (cached && Date.now() - cached.fetchedAt < CACHE_TTL) {
    return cached.config;
  }

  try {
    const { data } = await octokit.repos.getContent({
      owner,
      repo,
      path: CONFIG_FILE,
      ref,
    });

    if ('content' in data) {
      const content = Buffer.from(data.content, 'base64').toString('utf-8');
      const parsed = YAML.parse(content);
      const config = repoConfigSchema.parse(parsed);

      // Cache result
      CONFIG_CACHE.set(cacheKey, { config, fetchedAt: Date.now() });

      logger.debug({ owner, repo }, 'Loaded repository configuration');
      return config;
    }
  } catch (error: any) {
    if (error.status !== 404) {
      logger.warn({ error, owner, repo }, 'Failed to load repo config');
    }
  }

  // Return defaults
  const defaultConfig = repoConfigSchema.parse({});
  CONFIG_CACHE.set(cacheKey, { config: defaultConfig, fetchedAt: Date.now() });
  return defaultConfig;
}

/**
 * Clear configuration cache for a repository.
 */
export function clearRepoConfigCache(owner: string, repo: string): void {
  for (const key of CONFIG_CACHE.keys()) {
    if (key.startsWith(`${owner}/${repo}:`)) {
      CONFIG_CACHE.delete(key);
    }
  }
}
```

---

## Commit 14.3: Crear archivo .env.example

**Mensaje de commit:**
```
docs(github-app): add env example file

- Document all environment variables
- Add comments explaining each setting
- Include example values
```

### `github-app/.env.example`

```bash
# ===================================
# Server Configuration
# ===================================

# Node environment (development, production, test)
NODE_ENV=development

# Port to run the server on
PORT=3000

# Log level (debug, info, warn, error)
LOG_LEVEL=info

# ===================================
# GitHub App Configuration
# ===================================

# Your GitHub App ID (found in app settings)
GITHUB_APP_ID=123456

# Private key for the GitHub App
# Generate this in your app settings and paste the entire key
# Note: Newlines should be escaped as \n in .env files
GITHUB_PRIVATE_KEY="-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----"

# Webhook secret (min 20 characters)
# Use a strong random string, e.g.: openssl rand -hex 32
GITHUB_WEBHOOK_SECRET=your-webhook-secret-here

# Optional: OAuth credentials for user authentication
GITHUB_CLIENT_ID=
GITHUB_CLIENT_SECRET=

# ===================================
# AI Provider Configuration
# ===================================

# AI provider to use (ollama, openai)
AI_PROVIDER=ollama

# Model to use for code review
AI_MODEL=qwen2.5-coder:14b

# Ollama base URL (when using ollama provider)
OLLAMA_BASE_URL=http://localhost:11434

# OpenAI API key (required when using openai provider)
OPENAI_API_KEY=

# ===================================
# Rate Limiting
# ===================================

# Requests per second limit
RATE_LIMIT_RPS=10

# Burst limit (max concurrent requests)
RATE_LIMIT_BURST=20

# ===================================
# Cache Configuration
# ===================================

# Cache time-to-live (e.g., 1h, 30m, 60s)
CACHE_TTL=1h

# Maximum cache entries
CACHE_MAX_ENTRIES=1000

# ===================================
# Review Settings
# ===================================

# Maximum files to review per PR
REVIEW_MAX_FILES=50

# Maximum diff size in bytes (500KB default)
REVIEW_MAX_DIFF_SIZE=500000

# Review timeout (e.g., 5m, 10m)
REVIEW_TIMEOUT=5m
```

---

## Commit 14.4: Tests de configuracion

**Mensaje de commit:**
```
test(github-app): add configuration tests

- Test env schema validation
- Test repo config loading
- Test config caching
```

### `github-app/src/__tests__/config.test.ts`

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { envSchema } from '../config/schema.js';

describe('Environment Schema', () => {
  const validEnv = {
    GITHUB_APP_ID: '12345',
    GITHUB_PRIVATE_KEY: '-----BEGIN RSA PRIVATE KEY-----\\ntest\\n-----END RSA PRIVATE KEY-----',
    GITHUB_WEBHOOK_SECRET: 'a-very-long-secret-key-here',
  };

  it('validates required fields', () => {
    const result = envSchema.safeParse(validEnv);
    expect(result.success).toBe(true);
  });

  it('rejects missing required fields', () => {
    const result = envSchema.safeParse({});
    expect(result.success).toBe(false);
  });

  it('applies defaults', () => {
    const result = envSchema.safeParse(validEnv);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.PORT).toBe(3000);
      expect(result.data.NODE_ENV).toBe('development');
      expect(result.data.AI_PROVIDER).toBe('ollama');
    }
  });

  it('transforms private key newlines', () => {
    const result = envSchema.safeParse(validEnv);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.GITHUB_PRIVATE_KEY).toContain('\n');
    }
  });

  it('parses duration strings', () => {
    const envWithDuration = {
      ...validEnv,
      CACHE_TTL: '30m',
    };
    const result = envSchema.safeParse(envWithDuration);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.CACHE_TTL).toBe(30 * 60 * 1000);
    }
  });

  it('rejects invalid duration format', () => {
    const envWithBadDuration = {
      ...validEnv,
      CACHE_TTL: 'invalid',
    };
    const result = envSchema.safeParse(envWithBadDuration);
    expect(result.success).toBe(false);
  });

  it('coerces numeric strings', () => {
    const envWithStrings = {
      ...validEnv,
      PORT: '8080',
      GITHUB_APP_ID: '99999',
    };
    const result = envSchema.safeParse(envWithStrings);
    expect(result.success).toBe(true);
    if (result.success) {
      expect(result.data.PORT).toBe(8080);
      expect(result.data.GITHUB_APP_ID).toBe(99999);
    }
  });
});
```

### `github-app/src/__tests__/repoConfig.test.ts`

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { loadRepoConfig, clearRepoConfigCache } from '../config/repoConfig.js';

// Mock Octokit
const mockOctokit = {
  repos: {
    getContent: vi.fn(),
  },
};

describe('Repository Configuration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    clearRepoConfigCache('owner', 'repo');
  });

  it('returns defaults when no config file', async () => {
    mockOctokit.repos.getContent.mockRejectedValue({ status: 404 });

    const config = await loadRepoConfig(
      mockOctokit as any,
      'owner',
      'repo'
    );

    expect(config.review.enabled).toBe(true);
    expect(config.rules.preset).toBe('standard');
  });

  it('parses valid YAML config', async () => {
    const yamlContent = `
version: "1.0"
review:
  enabled: true
  max_files: 100
rules:
  preset: strict
`;

    mockOctokit.repos.getContent.mockResolvedValue({
      data: {
        content: Buffer.from(yamlContent).toString('base64'),
      },
    });

    const config = await loadRepoConfig(
      mockOctokit as any,
      'owner',
      'repo'
    );

    expect(config.review.max_files).toBe(100);
    expect(config.rules.preset).toBe('strict');
  });

  it('caches configuration', async () => {
    const yamlContent = 'version: "1.0"';

    mockOctokit.repos.getContent.mockResolvedValue({
      data: {
        content: Buffer.from(yamlContent).toString('base64'),
      },
    });

    await loadRepoConfig(mockOctokit as any, 'owner', 'repo');
    await loadRepoConfig(mockOctokit as any, 'owner', 'repo');

    // Should only call API once due to caching
    expect(mockOctokit.repos.getContent).toHaveBeenCalledTimes(1);
  });
});
```

---

## Resumen de la Iteracion 14

### Commits:
1. `feat(github-app): add zod config schema`
2. `feat(github-app): add per-repo configuration`
3. `docs(github-app): add env example file`
4. `test(github-app): add configuration tests`

### Archivos:
```
github-app/
├── .env.example
└── src/
    ├── config/
    │   ├── index.ts
    │   ├── schema.ts
    │   └── repoConfig.ts
    └── __tests__/
        ├── config.test.ts
        └── repoConfig.test.ts
```

---

## Siguiente Iteracion

Continua con: **[15-GITHUB-APP-SERVICES.md](15-GITHUB-APP-SERVICES.md)**
