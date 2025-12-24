/**
 * AI Provider Factory.
 * Creates the appropriate AI provider based on configuration.
 */
import { AIProvider } from './types.js';
export * from './types.js';
/**
 * Create an AI provider based on configuration.
 */
export declare function createProvider(): AIProvider;
/**
 * Get the AI provider singleton.
 */
export declare function getProvider(): AIProvider;
/**
 * Reset the provider (useful for testing).
 */
export declare function resetProvider(): void;
//# sourceMappingURL=index.d.ts.map