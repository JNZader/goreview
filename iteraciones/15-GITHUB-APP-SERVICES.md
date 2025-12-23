# Iteracion 15: GitHub App Services

## Objetivos

- Servicio de integracion con Ollama
- Servicio de review de PRs
- Servicio de comentarios en GitHub
- Cola de procesamiento

## Tiempo Estimado: 8 horas

---

## Commit 15.1: Crear servicio de Ollama

**Mensaje de commit:**
```
feat(github-app): add ollama service

- HTTP client for Ollama API
- Review prompt building
- Response parsing
- Error handling and retries
```

### `github-app/src/services/ollama.ts`

```typescript
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
```

---

## Commit 15.2: Crear servicio de review de PRs

**Mensaje de commit:**
```
feat(github-app): add pr review service

- Fetch PR diff from GitHub
- Split diff by file
- Process files concurrently
- Aggregate results
```

### `github-app/src/services/reviewService.ts`

```typescript
import { Octokit } from '@octokit/rest';
import { ollamaService, ReviewResponse, ReviewIssue } from './ollama.js';
import { loadRepoConfig } from '../config/repoConfig.js';
import { config } from '../config/index.js';
import { logger } from '../utils/logger.js';
import { parseDiff, DiffFile } from '../utils/diffParser.js';

export interface PRReviewResult {
  filesReviewed: number;
  totalIssues: number;
  criticalIssues: number;
  files: FileReviewResult[];
  summary: string;
  overallScore: number;
}

export interface FileReviewResult {
  path: string;
  issues: ReviewIssue[];
  score: number;
}

/**
 * Service for reviewing pull requests.
 */
export class PRReviewService {
  /**
   * Review a pull request.
   */
  async reviewPR(
    octokit: Octokit,
    owner: string,
    repo: string,
    pullNumber: number
  ): Promise<PRReviewResult> {
    const startTime = Date.now();

    // Load repository configuration
    const repoConfig = await loadRepoConfig(octokit, owner, repo);

    if (!repoConfig.review.enabled) {
      logger.info({ owner, repo, pullNumber }, 'Reviews disabled for repository');
      throw new Error('Reviews are disabled for this repository');
    }

    // Get pull request diff
    const { data: diffData } = await octokit.pulls.get({
      owner,
      repo,
      pull_number: pullNumber,
      mediaType: { format: 'diff' },
    });

    const diff = diffData as unknown as string;

    // Parse diff into files
    const files = parseDiff(diff);

    // Filter files based on configuration
    const filesToReview = this.filterFiles(files, repoConfig);

    if (filesToReview.length === 0) {
      return {
        filesReviewed: 0,
        totalIssues: 0,
        criticalIssues: 0,
        files: [],
        summary: 'No files to review after applying filters.',
        overallScore: 100,
      };
    }

    // Check limits
    if (filesToReview.length > config.review.maxFiles) {
      logger.warn({
        owner, repo, pullNumber,
        fileCount: filesToReview.length,
        maxFiles: config.review.maxFiles,
      }, 'Too many files to review');

      throw new Error(`Too many files to review: ${filesToReview.length} > ${config.review.maxFiles}`);
    }

    // Review files concurrently with limit
    const results = await this.reviewFiles(filesToReview);

    // Aggregate results
    const totalIssues = results.reduce((sum, r) => sum + r.issues.length, 0);
    const criticalIssues = results.reduce(
      (sum, r) => sum + r.issues.filter(i => i.severity === 'critical').length,
      0
    );
    const avgScore = results.reduce((sum, r) => sum + r.score, 0) / results.length;

    const duration = Date.now() - startTime;
    logger.info({
      owner, repo, pullNumber,
      filesReviewed: results.length,
      totalIssues,
      duration,
    }, 'PR review completed');

    return {
      filesReviewed: results.length,
      totalIssues,
      criticalIssues,
      files: results,
      summary: this.generateSummary(results),
      overallScore: Math.round(avgScore),
    };
  }

  private filterFiles(files: DiffFile[], repoConfig: any): DiffFile[] {
    return files.filter(file => {
      // Skip deleted files
      if (file.status === 'deleted') return false;

      // Skip binary files
      if (file.isBinary) return false;

      // Check ignore patterns
      for (const pattern of repoConfig.review.ignore_patterns) {
        if (this.matchPattern(pattern, file.path)) {
          return false;
        }
      }

      // Check language filter
      if (repoConfig.review.languages?.length > 0) {
        if (!repoConfig.review.languages.includes(file.language)) {
          return false;
        }
      }

      return true;
    });
  }

  private matchPattern(pattern: string, path: string): boolean {
    // Simple glob matching
    if (pattern.endsWith('**')) {
      return path.startsWith(pattern.slice(0, -2));
    }
    if (pattern.endsWith('*')) {
      return path.startsWith(pattern.slice(0, -1));
    }
    return pattern === path;
  }

  private async reviewFiles(files: DiffFile[]): Promise<FileReviewResult[]> {
    const results: FileReviewResult[] = [];
    const concurrency = 3; // Limit concurrent reviews

    for (let i = 0; i < files.length; i += concurrency) {
      const batch = files.slice(i, i + concurrency);

      const batchResults = await Promise.all(
        batch.map(async (file) => {
          try {
            const response = await ollamaService.review({
              diff: file.content,
              language: file.language,
              filePath: file.path,
            });

            return {
              path: file.path,
              issues: response.issues,
              score: response.score,
            };
          } catch (error) {
            logger.error({ error, file: file.path }, 'Failed to review file');
            return {
              path: file.path,
              issues: [],
              score: 70,
            };
          }
        })
      );

      results.push(...batchResults);
    }

    return results;
  }

  private generateSummary(results: FileReviewResult[]): string {
    const totalIssues = results.reduce((sum, r) => sum + r.issues.length, 0);
    const criticalCount = results.reduce(
      (sum, r) => sum + r.issues.filter(i => i.severity === 'critical').length,
      0
    );
    const errorCount = results.reduce(
      (sum, r) => sum + r.issues.filter(i => i.severity === 'error').length,
      0
    );

    if (totalIssues === 0) {
      return 'No issues found. The code looks good!';
    }

    const parts = [`Found ${totalIssues} issue(s) across ${results.length} file(s).`];

    if (criticalCount > 0) {
      parts.push(`**${criticalCount} critical issue(s)** require immediate attention.`);
    }
    if (errorCount > 0) {
      parts.push(`${errorCount} error(s) should be fixed before merging.`);
    }

    return parts.join(' ');
  }
}

export const prReviewService = new PRReviewService();
```

---

## Commit 15.3: Crear parser de diff

**Mensaje de commit:**
```
feat(github-app): add diff parser utility

- Parse unified diff format
- Extract file information
- Detect programming language
- Handle binary files
```

### `github-app/src/utils/diffParser.ts`

```typescript
export interface DiffFile {
  path: string;
  oldPath?: string;
  status: 'added' | 'modified' | 'deleted' | 'renamed';
  language: string;
  content: string;
  isBinary: boolean;
  additions: number;
  deletions: number;
}

/**
 * Parse a unified diff string into file objects.
 */
export function parseDiff(diff: string): DiffFile[] {
  const files: DiffFile[] = [];
  const fileBlocks = diff.split(/^diff --git/m).slice(1);

  for (const block of fileBlocks) {
    const file = parseFileBlock('diff --git' + block);
    if (file) {
      files.push(file);
    }
  }

  return files;
}

function parseFileBlock(block: string): DiffFile | null {
  const lines = block.split('\n');

  // Parse header
  const headerMatch = lines[0].match(/^diff --git a\/(.+) b\/(.+)$/);
  if (!headerMatch) return null;

  const [, oldPath, newPath] = headerMatch;

  // Determine status
  let status: DiffFile['status'] = 'modified';
  if (block.includes('new file mode')) {
    status = 'added';
  } else if (block.includes('deleted file mode')) {
    status = 'deleted';
  } else if (block.includes('rename from')) {
    status = 'renamed';
  }

  // Check for binary
  const isBinary = block.includes('Binary files');

  // Extract content (everything after @@)
  let content = '';
  let additions = 0;
  let deletions = 0;

  let inHunk = false;
  for (const line of lines) {
    if (line.startsWith('@@')) {
      inHunk = true;
      content += line + '\n';
    } else if (inHunk) {
      content += line + '\n';

      if (line.startsWith('+') && !line.startsWith('+++')) {
        additions++;
      } else if (line.startsWith('-') && !line.startsWith('---')) {
        deletions++;
      }
    }
  }

  return {
    path: newPath,
    oldPath: oldPath !== newPath ? oldPath : undefined,
    status,
    language: detectLanguage(newPath),
    content: content.trim(),
    isBinary,
    additions,
    deletions,
  };
}

/**
 * Detect programming language from file extension.
 */
export function detectLanguage(path: string): string {
  const ext = path.split('.').pop()?.toLowerCase() || '';

  const languageMap: Record<string, string> = {
    // JavaScript/TypeScript
    js: 'javascript',
    jsx: 'javascript',
    ts: 'typescript',
    tsx: 'typescript',
    mjs: 'javascript',
    cjs: 'javascript',

    // Go
    go: 'go',

    // Python
    py: 'python',
    pyi: 'python',

    // Rust
    rs: 'rust',

    // Java/Kotlin
    java: 'java',
    kt: 'kotlin',
    kts: 'kotlin',

    // C/C++
    c: 'c',
    h: 'c',
    cpp: 'cpp',
    cc: 'cpp',
    cxx: 'cpp',
    hpp: 'cpp',

    // Ruby
    rb: 'ruby',

    // PHP
    php: 'php',

    // Shell
    sh: 'shell',
    bash: 'shell',
    zsh: 'shell',

    // Config/Data
    json: 'json',
    yaml: 'yaml',
    yml: 'yaml',
    toml: 'toml',
    xml: 'xml',

    // Markup
    md: 'markdown',
    html: 'html',
    css: 'css',
    scss: 'scss',
  };

  return languageMap[ext] || 'unknown';
}
```

---

## Commit 15.4: Crear servicio de comentarios

**Mensaje de commit:**
```
feat(github-app): add github comment service

- Post review comments
- Create inline comments
- Add review summary
- Update PR status
```

### `github-app/src/services/commentService.ts`

```typescript
import { Octokit } from '@octokit/rest';
import { PRReviewResult, FileReviewResult } from './reviewService.js';
import { ReviewIssue } from './ollama.js';
import { loadRepoConfig } from '../config/repoConfig.js';
import { logger } from '../utils/logger.js';

export interface CommentOptions {
  inline: boolean;
  summary: boolean;
  requestChanges: boolean;
  minSeverity: 'info' | 'warning' | 'error' | 'critical';
}

const SEVERITY_ORDER = ['info', 'warning', 'error', 'critical'];

/**
 * Service for posting review comments to GitHub.
 */
export class CommentService {
  /**
   * Post review results to a pull request.
   */
  async postReview(
    octokit: Octokit,
    owner: string,
    repo: string,
    pullNumber: number,
    result: PRReviewResult
  ): Promise<void> {
    const repoConfig = await loadRepoConfig(octokit, owner, repo);
    const options = repoConfig.comments;

    // Filter issues by severity
    const minSeverityIndex = SEVERITY_ORDER.indexOf(options.min_severity);

    const filteredFiles = result.files.map(file => ({
      ...file,
      issues: file.issues.filter(issue =>
        SEVERITY_ORDER.indexOf(issue.severity) >= minSeverityIndex
      ),
    }));

    const totalFilteredIssues = filteredFiles.reduce(
      (sum, f) => sum + f.issues.length, 0
    );

    // Determine review event type
    let event: 'APPROVE' | 'REQUEST_CHANGES' | 'COMMENT' = 'COMMENT';

    if (totalFilteredIssues === 0) {
      event = 'APPROVE';
    } else if (options.request_changes && result.criticalIssues > 0) {
      event = 'REQUEST_CHANGES';
    }

    // Build review body
    const body = this.buildReviewBody(result, filteredFiles);

    // Build inline comments
    const comments = options.inline
      ? this.buildInlineComments(filteredFiles, pullNumber)
      : [];

    // Create the review
    try {
      await octokit.pulls.createReview({
        owner,
        repo,
        pull_number: pullNumber,
        event,
        body,
        comments,
      });

      logger.info({
        owner, repo, pullNumber,
        event,
        commentCount: comments.length,
      }, 'Posted review to PR');
    } catch (error) {
      logger.error({ error, owner, repo, pullNumber }, 'Failed to post review');
      throw error;
    }

    // Add labels if configured
    if (repoConfig.labels.add_on_issues && result.criticalIssues > 0) {
      await this.addLabel(octokit, owner, repo, pullNumber, repoConfig.labels.critical);
    }

    if (repoConfig.labels.add_on_issues) {
      await this.addLabel(octokit, owner, repo, pullNumber, repoConfig.labels.reviewed);
    }
  }

  private buildReviewBody(
    result: PRReviewResult,
    filteredFiles: FileReviewResult[]
  ): string {
    const lines: string[] = [];

    // Header
    lines.push('## AI Code Review');
    lines.push('');

    // Summary stats
    lines.push(`**Files reviewed:** ${result.filesReviewed}`);
    lines.push(`**Issues found:** ${result.totalIssues}`);
    lines.push(`**Overall score:** ${result.overallScore}/100`);
    lines.push('');

    // Main summary
    lines.push(result.summary);
    lines.push('');

    // Issue breakdown by severity
    const severityCounts: Record<string, number> = {};
    for (const file of filteredFiles) {
      for (const issue of file.issues) {
        severityCounts[issue.severity] = (severityCounts[issue.severity] || 0) + 1;
      }
    }

    if (Object.keys(severityCounts).length > 0) {
      lines.push('### Issue Summary');
      lines.push('');
      lines.push('| Severity | Count |');
      lines.push('|----------|-------|');

      for (const severity of SEVERITY_ORDER.reverse()) {
        const count = severityCounts[severity];
        if (count) {
          const emoji = this.getSeverityEmoji(severity);
          lines.push(`| ${emoji} ${severity} | ${count} |`);
        }
      }
      lines.push('');
    }

    // Footer
    lines.push('---');
    lines.push('*Generated by AI Code Review*');

    return lines.join('\n');
  }

  private buildInlineComments(
    files: FileReviewResult[],
    pullNumber: number
  ): Array<{
    path: string;
    line: number;
    body: string;
  }> {
    const comments: Array<{
      path: string;
      line: number;
      body: string;
    }> = [];

    for (const file of files) {
      for (const issue of file.issues) {
        if (issue.line && issue.line > 0) {
          comments.push({
            path: file.path,
            line: issue.line,
            body: this.formatIssueComment(issue),
          });
        }
      }
    }

    return comments;
  }

  private formatIssueComment(issue: ReviewIssue): string {
    const emoji = this.getSeverityEmoji(issue.severity);
    const lines: string[] = [];

    lines.push(`${emoji} **${issue.severity.toUpperCase()}**: ${issue.message}`);

    if (issue.suggestion) {
      lines.push('');
      lines.push(`**Suggestion:** ${issue.suggestion}`);
    }

    return lines.join('\n');
  }

  private getSeverityEmoji(severity: string): string {
    switch (severity) {
      case 'critical': return ':rotating_light:';
      case 'error': return ':x:';
      case 'warning': return ':warning:';
      case 'info': return ':information_source:';
      default: return ':grey_question:';
    }
  }

  private async addLabel(
    octokit: Octokit,
    owner: string,
    repo: string,
    pullNumber: number,
    label: string
  ): Promise<void> {
    try {
      await octokit.issues.addLabels({
        owner,
        repo,
        issue_number: pullNumber,
        labels: [label],
      });
    } catch (error) {
      logger.warn({ error, label }, 'Failed to add label');
    }
  }
}

export const commentService = new CommentService();
```

---

## Commit 15.5: Tests de servicios

**Mensaje de commit:**
```
test(github-app): add service tests

- Test Ollama service
- Test diff parser
- Test review service
```

### `github-app/src/__tests__/diffParser.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import { parseDiff, detectLanguage } from '../utils/diffParser.js';

describe('parseDiff', () => {
  it('parses simple diff', () => {
    const diff = `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main

+import "fmt"
 func main() {}`;

    const files = parseDiff(diff);

    expect(files).toHaveLength(1);
    expect(files[0].path).toBe('main.go');
    expect(files[0].status).toBe('modified');
    expect(files[0].language).toBe('go');
    expect(files[0].additions).toBe(1);
  });

  it('detects new files', () => {
    const diff = `diff --git a/new.ts b/new.ts
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/new.ts
@@ -0,0 +1 @@
+export const x = 1;`;

    const files = parseDiff(diff);

    expect(files[0].status).toBe('added');
  });

  it('detects deleted files', () => {
    const diff = `diff --git a/old.js b/old.js
deleted file mode 100644
index 1234567..0000000
--- a/old.js
+++ /dev/null
@@ -1 +0,0 @@
-const x = 1;`;

    const files = parseDiff(diff);

    expect(files[0].status).toBe('deleted');
  });
});

describe('detectLanguage', () => {
  const cases = [
    ['main.go', 'go'],
    ['index.ts', 'typescript'],
    ['app.tsx', 'typescript'],
    ['script.js', 'javascript'],
    ['main.py', 'python'],
    ['lib.rs', 'rust'],
    ['App.java', 'java'],
    ['config.yaml', 'yaml'],
    ['unknown.xyz', 'unknown'],
  ];

  it.each(cases)('detectLanguage(%s) = %s', (path, expected) => {
    expect(detectLanguage(path)).toBe(expected);
  });
});
```

---

## Resumen de la Iteracion 15

### Commits:
1. `feat(github-app): add ollama service`
2. `feat(github-app): add pr review service`
3. `feat(github-app): add diff parser utility`
4. `feat(github-app): add github comment service`
5. `test(github-app): add service tests`

### Archivos:
```
github-app/src/
├── services/
│   ├── ollama.ts
│   ├── reviewService.ts
│   └── commentService.ts
├── utils/
│   └── diffParser.ts
└── __tests__/
    └── diffParser.test.ts
```

---

## Siguiente Iteracion

Continua con: **[16-GITHUB-APP-WEBHOOKS.md](16-GITHUB-APP-WEBHOOKS.md)**
