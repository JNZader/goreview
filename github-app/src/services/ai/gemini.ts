/**
 * Gemini AI Provider.
 * Uses Google Generative AI API.
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

export class GeminiProvider implements AIProvider {
  readonly name = 'gemini';
  private readonly baseUrl: string;
  private readonly model: string;
  private readonly apiKey: string;

  constructor() {
    this.apiKey = config.ai.geminiApiKey || '';
    this.baseUrl = 'https://generativelanguage.googleapis.com/v1beta';
    this.model = config.ai.model || 'gemini-2.0-flash';
  }

  async review(request: ReviewRequest): Promise<ReviewResponse> {
    const startTime = Date.now();
    const prompt = buildReviewPrompt(request);

    const geminiReq = {
      contents: [
        {
          parts: [{ text: prompt }],
        },
      ],
      generationConfig: {
        temperature: 0.3,
        maxOutputTokens: 4096,
        responseMimeType: 'application/json',
      },
    };

    const url = `${this.baseUrl}/models/${this.model}:generateContent?key=${this.apiKey}`;

    logger.debug({ model: this.model }, 'Calling Gemini API');

    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(geminiReq),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Gemini error: ${response.status} ${error}`);
    }

    const result = (await response.json()) as {
      candidates?: Array<{
        content?: {
          parts?: Array<{ text?: string }>;
        };
      }>;
      usageMetadata?: {
        totalTokenCount?: number;
      };
      error?: {
        message: string;
        code: number;
      };
    };

    if (result.error) {
      throw new Error(`Gemini error ${result.error.code}: ${result.error.message}`);
    }

    let reviewResponse: ReviewResponse;
    const text = result.candidates?.[0]?.content?.parts?.[0]?.text;
    if (text) {
      reviewResponse = parseReviewResponse(text);
    } else {
      reviewResponse = { issues: [], summary: 'No response from Gemini', score: 70 };
    }

    reviewResponse.tokensUsed = result.usageMetadata?.totalTokenCount;
    reviewResponse.processingTimeMs = Date.now() - startTime;

    return reviewResponse;
  }

  async generateCommitMessage(diff: string): Promise<string> {
    const prompt = buildCommitPrompt(diff);

    const geminiReq = {
      contents: [{ parts: [{ text: prompt }] }],
    };

    const url = `${this.baseUrl}/models/${this.model}:generateContent?key=${this.apiKey}`;

    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(geminiReq),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      throw new Error(`Gemini error: ${response.status}`);
    }

    const result = (await response.json()) as {
      candidates?: Array<{
        content?: {
          parts?: Array<{ text?: string }>;
        };
      }>;
    };

    const commitText = result.candidates?.[0]?.content?.parts?.[0]?.text;
    if (commitText) {
      return commitText.trim();
    }

    throw new Error('No response from Gemini');
  }

  async chat(prompt: string): Promise<string> {
    const geminiReq = {
      contents: [{ parts: [{ text: prompt }] }],
      generationConfig: {
        temperature: 0.7,
        maxOutputTokens: 2048,
      },
    };

    const url = `${this.baseUrl}/models/${this.model}:generateContent?key=${this.apiKey}`;

    const response = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(geminiReq),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      throw new Error(`Gemini error: ${response.status}`);
    }

    const result = (await response.json()) as {
      candidates?: Array<{
        content?: {
          parts?: Array<{ text?: string }>;
        };
      }>;
    };

    const chatText = result.candidates?.[0]?.content?.parts?.[0]?.text;
    if (chatText) {
      return chatText.trim();
    }

    throw new Error('No response from Gemini');
  }

  async healthCheck(): Promise<boolean> {
    try {
      const url = `${this.baseUrl}/models/${this.model}?key=${this.apiKey}`;
      const response = await fetch(url, {
        signal: AbortSignal.timeout(5000),
      });
      return response.ok;
    } catch {
      return false;
    }
  }
}
