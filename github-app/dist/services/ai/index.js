/**
 * AI Provider Factory.
 * Creates the appropriate AI provider based on configuration.
 */
import { config } from '../../config/index.js';
import { logger } from '../../utils/logger.js';
import { OllamaProvider } from './ollama.js';
import { GeminiProvider } from './gemini.js';
import { GroqProvider } from './groq.js';
// Re-export types
export * from './types.js';
/**
 * Create an AI provider based on configuration.
 */
export function createProvider() {
    const providerName = config.ai.provider;
    logger.info({ provider: providerName }, 'Creating AI provider');
    switch (providerName) {
        case 'ollama':
            return new OllamaProvider();
        case 'gemini':
            if (!config.ai.geminiApiKey) {
                throw new Error('GEMINI_API_KEY is required for Gemini provider');
            }
            return new GeminiProvider();
        case 'groq':
            if (!config.ai.groqApiKey) {
                throw new Error('GROQ_API_KEY is required for Groq provider');
            }
            return new GroqProvider();
        case 'auto':
            return createAutoProvider();
        default:
            throw new Error(`Unknown AI provider: ${providerName}`);
    }
}
/**
 * Auto-detect the best available provider.
 * Priority: Gemini -> Groq -> Ollama
 */
function createAutoProvider() {
    // Try Gemini first (best quality, free tier)
    if (config.ai.geminiApiKey) {
        logger.info('Auto-detected Gemini API key, using Gemini provider');
        return new GeminiProvider();
    }
    // Try Groq second (fastest, free tier)
    if (config.ai.groqApiKey) {
        logger.info('Auto-detected Groq API key, using Groq provider');
        return new GroqProvider();
    }
    // Fallback to Ollama
    logger.info('No cloud API keys found, using Ollama provider');
    return new OllamaProvider();
}
// Singleton instance
let providerInstance = null;
/**
 * Get the AI provider singleton.
 */
export function getProvider() {
    if (!providerInstance) {
        providerInstance = createProvider();
    }
    return providerInstance;
}
/**
 * Reset the provider (useful for testing).
 */
export function resetProvider() {
    providerInstance = null;
}
//# sourceMappingURL=index.js.map