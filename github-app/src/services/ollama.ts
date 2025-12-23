import { config } from '../config/index.js';
import { logger } from '../utils/logger.js';

export interface ReviewRequest {
  diff: string;
  language: string;
  filePath: string;
  context?: string;
}

export interface ReviewIssue {
  id: string;
  type: 'bug' | 'security' | 'performance' | 'style' | 'best_practice';
  severity: 'info' | 'warning' | 'error' | 'critical';
  message: string;
  suggestion?: string;
  line?: number;
  endLine?: number;
}

export interface ReviewResponse {
  issues: ReviewIssue[];
  summary: string;
  score: number;
}

/**
 * Ollama service for code review.
 */
export class OllamaService {
  private baseUrl: string;
  private model: string;

  constructor() {
    this.baseUrl = config.ai.ollamaBaseUrl;
    this.model = config.ai.model;
  }

  /**
   * Review code and return issues.
   */
  async review(request: ReviewRequest): Promise<ReviewResponse> {
    const prompt = this.buildPrompt(request);

    const response = await this.generate(prompt);

    return this.parseResponse(response);
  }

  /**
   * Generate a commit message from diff.
   */
  async generateCommitMessage(diff: string): Promise<string> {
    const prompt = `Generate a conventional commit message for this diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, chore

Diff:
${diff}

Return ONLY the commit message, nothing else.`;

    return this.generate(prompt, { format: undefined });
  }

  /**
   * Check if Ollama is available.
   */
  async healthCheck(): Promise<boolean> {
    try {
      const response = await fetch(`${this.baseUrl}/api/tags`);
      return response.ok;
    } catch {
      return false;
    }
  }

  private async generate(
    prompt: string,
    options: { format?: string } = { format: 'json' }
  ): Promise<string> {
    const body = {
      model: this.model,
      prompt,
      stream: false,
      format: options.format,
      options: {
        temperature: 0.3,
        num_predict: 4096,
      },
    };

    logger.debug({ model: this.model }, 'Calling Ollama API');

    const response = await fetch(`${this.baseUrl}/api/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      signal: AbortSignal.timeout(config.review.timeout),
    });

    if (!response.ok) {
      const error = await response.text();
      throw new Error(`Ollama error: ${response.status} ${error}`);
    }

    const result = await response.json();
    return result.response;
  }

  private buildPrompt(request: ReviewRequest): string {
    return `You are an expert code reviewer. Analyze this code and identify issues.

File: ${request.filePath}
Language: ${request.language}
${request.context ? `Context: ${request.context}` : ''}

Code changes:
${request.diff}

Return a JSON object with this structure:
{
  "issues": [
    {
      "id": "1",
      "type": "bug|security|performance|style|best_practice",
      "severity": "info|warning|error|critical",
      "message": "description of the issue",
      "suggestion": "how to fix it",
      "line": 10
    }
  ],
  "summary": "brief summary of the review",
  "score": 85
}

Only report real issues, not style nitpicks. Focus on:
- Security vulnerabilities
- Bugs and logic errors
- Performance problems
- Best practices violations

Return valid JSON only.`;
  }

  private parseResponse(response: string): ReviewResponse {
    try {
      const parsed = JSON.parse(response);

      return {
        issues: parsed.issues || [],
        summary: parsed.summary || 'No summary provided',
        score: typeof parsed.score === 'number' ? parsed.score : 70,
      };
    } catch (error) {
      logger.warn({ response }, 'Failed to parse Ollama response as JSON');

      // Fallback response
      return {
        issues: [],
        summary: response.slice(0, 200),
        score: 70,
      };
    }
  }
}

export const ollamaService = new OllamaService();
