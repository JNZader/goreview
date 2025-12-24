import dotenv from 'dotenv';
import { envSchema } from './schema.js';
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
        geminiApiKey: env.GEMINI_API_KEY,
        groqApiKey: env.GROQ_API_KEY,
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
};
//# sourceMappingURL=index.js.map