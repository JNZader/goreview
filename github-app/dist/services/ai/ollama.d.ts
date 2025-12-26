/**
 * Ollama AI Provider.
 * Local LLM inference.
 */
import { AIProvider, ReviewRequest, ReviewResponse } from './types.js';
export declare class OllamaProvider implements AIProvider {
    readonly name = "ollama";
    private readonly baseUrl;
    private readonly model;
    constructor();
    review(request: ReviewRequest): Promise<ReviewResponse>;
    generateCommitMessage(diff: string): Promise<string>;
    chat(prompt: string): Promise<string>;
    healthCheck(): Promise<boolean>;
}
//# sourceMappingURL=ollama.d.ts.map