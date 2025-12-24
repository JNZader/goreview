import { z } from 'zod';
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
const CONFIG_FILE = '.goreview.yaml';
const CONFIG_CACHE = new Map();
const CACHE_TTL = 5 * 60 * 1000; // 5 minutes
/**
 * Load configuration from a repository.
 */
export async function loadRepoConfig(octokit, owner, repo, ref) {
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
    }
    catch (error) {
        const err = error;
        if (err.status !== 404) {
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
export function clearRepoConfigCache(owner, repo) {
    for (const key of CONFIG_CACHE.keys()) {
        if (key.startsWith(`${owner}/${repo}:`)) {
            CONFIG_CACHE.delete(key);
        }
    }
}
//# sourceMappingURL=repoConfig.js.map