/**
 * Handler for interactive comments.
 * Responds to @goreview mentions in issue/PR comments.
 */

import type { IssueCommentEvent } from '@octokit/webhooks-types';
import { getOctokit } from '../services/github.js';
import { logger } from '../utils/logger.js';
import { getProvider } from '../services/ai/index.js';

// Bot username patterns to detect mentions
const BOT_MENTION_PATTERNS = [
  /@goreview\b/i,
  /@go-review\b/i,
  /@gorev\b/i,
];

// Command patterns
const COMMANDS = {
  explain: /\b(?:explain|why|what)\b.*(?:error|issue|problem|bug)/i,
  suggest: /\b(?:suggest|recommend|how\s+(?:to|should)|fix)\b/i,
  review: /\b(?:review|analyze|check)\b/i,
  help: /\b(?:help|commands|what\s+can\s+you)\b/i,
};

interface ParsedCommand {
  type: 'explain' | 'suggest' | 'review' | 'help' | 'general';
  context: string;
  originalComment: string;
}

/**
 * Handle issue comment events
 */
export async function handleIssueComment(event: IssueCommentEvent): Promise<void> {
  // Only handle new comments
  if (event.action !== 'created') {
    return;
  }

  const { comment, issue, repository, installation } = event;

  // Check if this is a mention of the bot
  if (!isBotMention(comment.body)) {
    return;
  }

  logger.info({
    repo: repository.full_name,
    issue: issue.number,
    author: comment.user.login,
  }, 'Bot mentioned in comment');

  // Don't respond to our own comments
  if (comment.user.type === 'Bot') {
    return;
  }

  // Parse the command
  const command = parseCommand(comment.body);

  // Get installation client
  const installationId = installation?.id;
  if (!installationId) {
    logger.error('No installation ID in event');
    return;
  }

  try {
    const octokit = await getOctokit(installationId);

    // Generate response
    const response = await generateResponse(command, {
      repo: repository.full_name,
      issueNumber: issue.number,
      isPullRequest: 'pull_request' in issue,
      author: comment.user.login,
    });

    // Post reply
    await octokit.issues.createComment({
      owner: repository.owner.login,
      repo: repository.name,
      issue_number: issue.number,
      body: response,
    });

    logger.info({
      repo: repository.full_name,
      issue: issue.number,
      commandType: command.type,
    }, 'Responded to bot mention');

  } catch (error) {
    logger.error({ error, repo: repository.full_name }, 'Failed to respond to mention');

    // Try to post error message
    try {
      const octokit = await getOctokit(installationId);
      await octokit.issues.createComment({
        owner: repository.owner.login,
        repo: repository.name,
        issue_number: issue.number,
        body: `Sorry, I encountered an error processing your request. Please try again later.`,
      });
    } catch {
      // Ignore error posting error message
    }
  }
}

/**
 * Check if the comment mentions the bot
 */
function isBotMention(body: string): boolean {
  return BOT_MENTION_PATTERNS.some(pattern => pattern.test(body));
}

/**
 * Parse the command from the comment
 */
function parseCommand(body: string): ParsedCommand {
  // Remove the @mention
  const cleanBody = body.replaceAll(/@goreview\b/gi, '').trim();

  // Detect command type
  let type: ParsedCommand['type'] = 'general';

  if (COMMANDS.help.test(cleanBody)) {
    type = 'help';
  } else if (COMMANDS.explain.test(cleanBody)) {
    type = 'explain';
  } else if (COMMANDS.suggest.test(cleanBody)) {
    type = 'suggest';
  } else if (COMMANDS.review.test(cleanBody)) {
    type = 'review';
  }

  return {
    type,
    context: cleanBody,
    originalComment: body,
  };
}

/**
 * Generate response based on command
 */
async function generateResponse(
  command: ParsedCommand,
  context: {
    repo: string;
    issueNumber: number;
    isPullRequest: boolean;
    author: string;
  }
): Promise<string> {
  // Help command - no AI needed
  if (command.type === 'help') {
    return generateHelpMessage();
  }

  // Use AI to generate response
  try {
    const provider = getProvider();

    const prompt = buildPrompt(command, context);
    const response = await provider.chat(prompt);

    return formatResponse(response, command.type);
  } catch (error) {
    logger.warn({ error }, 'AI response failed, using fallback');
    return generateFallbackResponse(command.type);
  }
}

/**
 * Build prompt for AI
 */
function buildPrompt(
  command: ParsedCommand,
  context: {
    repo: string;
    issueNumber: number;
    isPullRequest: boolean;
    author: string;
  }
): string {
  const baseContext = `
You are GoReview, an AI code review assistant.
You're responding to a comment in ${context.isPullRequest ? 'a pull request' : 'an issue'} on GitHub.
Repository: ${context.repo}
User asking: @${context.author}

The user's question/request:
${command.context}

Please provide a helpful, concise response. Use markdown formatting.
`;

  switch (command.type) {
    case 'explain':
      return `${baseContext}
The user is asking for an explanation about a code issue or error.
Explain the problem clearly and suggest how to understand it better.`;

    case 'suggest':
      return `${baseContext}
The user is asking for suggestions or recommendations.
Provide actionable suggestions with code examples if relevant.`;

    case 'review':
      return `${baseContext}
The user is asking for a code review or analysis.
Focus on the specific area they mentioned and provide constructive feedback.`;

    default:
      return `${baseContext}
Provide a helpful response to the user's question or comment.`;
  }
}

/**
 * Format AI response
 */
function formatResponse(aiResponse: string, commandType: string): string {
  // Add a header
  const header = getResponseHeader(commandType);

  return `${header}\n\n${aiResponse}`;
}

/**
 * Get response header emoji/icon
 */
function getResponseHeader(commandType: string): string {
  switch (commandType) {
    case 'explain':
      return '### Explanation';
    case 'suggest':
      return '### Suggestions';
    case 'review':
      return '### Review';
    default:
      return '### Response';
  }
}

/**
 * Generate help message
 */
function generateHelpMessage(): string {
  return `### GoReview Help

I'm an AI-powered code review assistant. Here's what I can do:

**Commands:**
- \`@goreview explain [issue]\` - Explain an error or code issue
- \`@goreview suggest [topic]\` - Get suggestions for improvements
- \`@goreview review [code/area]\` - Request a focused code review
- \`@goreview help\` - Show this help message

**Examples:**
- \`@goreview explain why this causes a null pointer exception\`
- \`@goreview suggest how to improve error handling here\`
- \`@goreview review the authentication logic\`

I'll automatically review pull requests when they're opened or updated.

---
*Powered by GoReview*`;
}

/**
 * Generate fallback response when AI fails
 */
function generateFallbackResponse(commandType: string): string {
  switch (commandType) {
    case 'explain':
      return `I'd be happy to explain, but I'm having trouble processing your request right now.
Please try again in a moment, or check the code comments and documentation for more context.`;

    case 'suggest':
      return `I'd like to help with suggestions, but I'm experiencing technical difficulties.
Please try again shortly, or refer to the project's contributing guidelines.`;

    case 'review':
      return `I'm unable to complete the review at the moment due to a technical issue.
Please try again later, or wait for the automatic review on your next push.`;

    default:
      return `Thanks for reaching out! I'm having a temporary issue processing your request.
Please try again in a few moments. Type \`@goreview help\` to see available commands.`;
  }
}

/**
 * Handle pull request review comment events
 * Allows responding to @mentions in review comments (on specific lines)
 */
export async function handlePullRequestReviewComment(event: {
  action: string;
  comment: {
    id: number;
    body: string;
    user: { login: string; type: string };
    diff_hunk?: string;
    path?: string;
    line?: number;
  };
  pull_request: {
    number: number;
  };
  repository: {
    full_name: string;
    owner: { login: string };
    name: string;
  };
  installation?: { id: number };
}): Promise<void> {
  if (event.action !== 'created') {
    return;
  }

  const { comment, pull_request, repository, installation } = event;

  if (!isBotMention(comment.body)) {
    return;
  }

  if (comment.user.type === 'Bot') {
    return;
  }

  logger.info({
    repo: repository.full_name,
    pr: pull_request.number,
    file: comment.path,
    line: comment.line,
  }, 'Bot mentioned in review comment');

  const installationId = installation?.id;
  if (!installationId) {
    logger.error('No installation ID');
    return;
  }

  try {
    const octokit = await getOctokit(installationId);

    // Enhanced context with code snippet
    const command = parseCommand(comment.body);
    const codeContext = comment.diff_hunk
      ? `\n\nCode context:\n\`\`\`\n${comment.diff_hunk}\n\`\`\`\nFile: ${comment.path}, Line: ${comment.line}`
      : '';

    const enhancedCommand: ParsedCommand = {
      ...command,
      context: command.context + codeContext,
    };

    const response = await generateResponse(enhancedCommand, {
      repo: repository.full_name,
      issueNumber: pull_request.number,
      isPullRequest: true,
      author: comment.user.login,
    });

    // Reply to the review comment
    await octokit.pulls.createReplyForReviewComment({
      owner: repository.owner.login,
      repo: repository.name,
      pull_number: pull_request.number,
      comment_id: comment.id,
      body: response,
    });

    logger.info({
      repo: repository.full_name,
      pr: pull_request.number,
    }, 'Responded to review comment mention');

  } catch (error) {
    logger.error({ error }, 'Failed to respond to review comment');
  }
}
