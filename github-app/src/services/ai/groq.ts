/**
 * Groq AI Provider.
 * Uses OpenAI-compatible API format.
 */

import { config } from '../../config/index.js';
import { logger } from '../../utils/logger.js';
import {
  AIProvider,
  ReviewRequest,
  ReviewResponse,
  buildReviewPrompt,
  buildCommitPrompt,
  parseReviewResponse,
} from './types.js';

export class GroqProvider implements AIProvider {
  readonly name = 'groq';
  private readonly baseUrl: string;
  private readonly model: string;
  private readonly apiKey: string;

  constructor() {
    this.apiKey = config.ai.groqApiKey || '';
    this.baseUrl = 'https://api.groq.com/openai/v1';
    this.model = config.ai.model || 'llama-3.3-70b-versatile';
  }

  async review(request: ReviewRequest): Promise<ReviewResponse> {
    const startTime = Date.now();
    const prompt = buildReviewPrompt(request);

    const groqReq = {
      model: this.model,
      messages: [
        { role: 'system', content: 'You are an expert code reviewer. Return valid JSON only.' },
        { role: 'user', content: prompt },
      ],
      temperature: 0.3,
      max_tokens: 4096,
      response_format: { type: 'json_object' },
    };

    logger.debug({ model: this.model }, 'Calling Groq API');

    const response = await fetch(`${this.baseUrl}/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify(groqReq),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Groq error: ${response.status} ${error}`);
    }

    const result = (await response.json()) as {
      choices?: Array<{
        message?: { content?: string };
      }>;
      usage?: {
        total_tokens?: number;
      };
      error?: {
        message: string;
        type: string;
      };
    };

    if (result.error) {
      throw new Error(`Groq error: ${result.error.message}`);
    }

    let reviewResponse: ReviewResponse;
    if (result.choices && result.choices[0]?.message?.content) {
      reviewResponse = parseReviewResponse(result.choices[0].message.content);
    } else {
      reviewResponse = { issues: [], summary: 'No response from Groq', score: 70 };
    }

    reviewResponse.tokensUsed = result.usage?.total_tokens;
    reviewResponse.processingTimeMs = Date.now() - startTime;

    return reviewResponse;
  }

  async generateCommitMessage(diff: string): Promise<string> {
    const prompt = buildCommitPrompt(diff);

    const groqReq = {
      model: this.model,
      messages: [{ role: 'user', content: prompt }],
    };

    const response = await fetch(`${this.baseUrl}/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify(groqReq),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      throw new Error(`Groq error: ${response.status}`);
    }

    const result = (await response.json()) as {
      choices?: Array<{
        message?: { content?: string };
      }>;
    };

    if (result.choices && result.choices[0]?.message?.content) {
      return result.choices[0].message.content.trim();
    }

    throw new Error('No response from Groq');
  }

  async chat(prompt: string): Promise<string> {
    const groqReq = {
      model: this.model,
      messages: [
        { role: 'system', content: 'You are a helpful code review assistant. Be concise and helpful.' },
        { role: 'user', content: prompt },
      ],
      temperature: 0.7,
      max_tokens: 2048,
    };

    const response = await fetch(`${this.baseUrl}/chat/completions`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${this.apiKey}`,
      },
      body: JSON.stringify(groqReq),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      throw new Error(`Groq error: ${response.status}`);
    }

    const result = (await response.json()) as {
      choices?: Array<{
        message?: { content?: string };
      }>;
    };

    if (result.choices && result.choices[0]?.message?.content) {
      return result.choices[0].message.content.trim();
    }

    throw new Error('No response from Groq');
  }

  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/models`, {
        headers: {
          Authorization: `Bearer ${this.apiKey}`,
        },
        signal: AbortSignal.timeout(5000),
      });
      return response.ok;
    } catch {
      return false;
    }
  }
}
