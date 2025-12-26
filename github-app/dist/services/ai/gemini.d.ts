/**
 * Gemini AI Provider.
 * Uses Google Generative AI API.
 */
import { AIProvider, ReviewRequest, ReviewResponse } from './types.js';
export declare class GeminiProvider implements AIProvider {
    readonly name = "gemini";
    private readonly baseUrl;
    private readonly model;
    private readonly apiKey;
    constructor();
    review(request: ReviewRequest): Promise<ReviewResponse>;
    generateCommitMessage(diff: string): Promise<string>;
    chat(prompt: string): Promise<string>;
    healthCheck(): Promise<boolean>;
}
//# sourceMappingURL=gemini.d.ts.map