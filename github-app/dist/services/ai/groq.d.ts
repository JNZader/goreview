/**
 * Groq AI Provider.
 * Uses OpenAI-compatible API format.
 */
import { AIProvider, ReviewRequest, ReviewResponse } from './types.js';
export declare class GroqProvider implements AIProvider {
    readonly name = "groq";
    private baseUrl;
    private model;
    private apiKey;
    constructor();
    review(request: ReviewRequest): Promise<ReviewResponse>;
    generateCommitMessage(diff: string): Promise<string>;
    chat(prompt: string): Promise<string>;
    healthCheck(): Promise<boolean>;
}
//# sourceMappingURL=groq.d.ts.map