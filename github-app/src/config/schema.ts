import { z } from 'zod';

// Duration parser (e.g., "1h", "30m", "60s")
// Using character class [hms] instead of alternation for SonarQube S6035
const DURATION_REGEX = /^(\d+)([hms])$/;

const durationSchema = z.string().superRefine((val, ctx) => {
  const match = DURATION_REGEX.exec(val);
  if (!match) {
    ctx.addIssue({
      code: z.ZodIssueCode.custom,
      message: `Invalid duration format: ${val}. Expected format like "1h", "30m", or "60s"`,
    });
  }
}).transform((val) => {
  const match = DURATION_REGEX.exec(val);
  if (!match) return 0; // Won't reach here due to superRefine

  const num = match[1] ?? '0';
  const unit = match[2] ?? 's';
  const multipliers: Record<string, number> = {
    h: 3600000,
    m: 60000,
    s: 1000,
  };

  return Number.parseInt(num, 10) * (multipliers[unit] ?? 1000);
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
    return key.replaceAll(String.raw`\n`, '\n');
  }),
  GITHUB_WEBHOOK_SECRET: z.string().min(20),
  GITHUB_CLIENT_ID: z.string().optional(),
  GITHUB_CLIENT_SECRET: z.string().optional(),

  // AI Provider
  AI_PROVIDER: z.enum(['ollama', 'openai', 'gemini', 'groq', 'auto']).default('auto'),
  AI_MODEL: z.string().default(''),
  OLLAMA_BASE_URL: z.string().url().default('http://localhost:11434'),
  OPENAI_API_KEY: z.string().optional(),
  GEMINI_API_KEY: z.string().optional(),
  GROQ_API_KEY: z.string().optional(),

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
