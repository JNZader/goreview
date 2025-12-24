/**
 * Ollama AI Provider.
 * Local LLM inference.
 */

import { config } from '../../config/index.js';
import { logger } from '../../utils/logger.js';
import {
  AIProvider,
  ReviewRequest,
  ReviewResponse,
  buildReviewPrompt,
  parseReviewResponse,
} from './types.js';

export class OllamaProvider implements AIProvider {
  readonly name = 'ollama';
  private baseUrl: string;
  private model: string;

  constructor() {
    this.baseUrl = config.ai.ollamaBaseUrl;
    this.model = config.ai.model;
  }

  async review(request: ReviewRequest): Promise<ReviewResponse> {
    const startTime = Date.now();
    const prompt = buildReviewPrompt(request);

    const ollamaReq = {
      model: this.model,
      prompt,
      stream: false,
      format: 'json',
      options: {
        temperature: 0.3,
        num_predict: 4096,
      },
    };

    logger.debug({ model: this.model }, 'Calling Ollama API');

    const response = await fetch(`${this.baseUrl}/api/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(ollamaReq),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Ollama error: ${response.status} ${error}`);
    }

    const result = (await response.json()) as {
      response: string;
      eval_count?: number;
    };

    const reviewResponse = parseReviewResponse(result.response);
    reviewResponse.tokensUsed = result.eval_count;
    reviewResponse.processingTimeMs = Date.now() - startTime;

    return reviewResponse;
  }

  async generateCommitMessage(diff: string): Promise<string> {
    const prompt = `Generate a conventional commit message for this diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, chore

Diff:
${diff}

Return ONLY the commit message, nothing else.`;

    const ollamaReq = {
      model: this.model,
      prompt,
      stream: false,
    };

    const response = await fetch(`${this.baseUrl}/api/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(ollamaReq),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      throw new Error(`Ollama error: ${response.status}`);
    }

    const result = (await response.json()) as { response: string };
    return result.response.trim();
  }

  async chat(prompt: string): Promise<string> {
    const ollamaReq = {
      model: this.model,
      prompt,
      stream: false,
      options: {
        temperature: 0.7,
        num_predict: 2048,
      },
    };

    const response = await fetch(`${this.baseUrl}/api/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(ollamaReq),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      throw new Error(`Ollama error: ${response.status}`);
    }

    const result = (await response.json()) as { response: string };
    return result.response.trim();
  }

  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/tags`, {
        signal: AbortSignal.timeout(5000),
      });
      return response.ok;
    } catch {
      return false;
    }
  }
}
