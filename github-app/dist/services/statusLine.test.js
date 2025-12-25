import { describe, it, expect, vi } from 'vitest';
// Mock the logger first
vi.mock('../utils/logger.js', () => ({
    logger: {
        debug: vi.fn(),
        info: vi.fn(),
        warn: vi.fn(),
        error: vi.fn(),
    },
}));
// Mock the config
vi.mock('../config/index.js', () => ({
    config: {
        nodeEnv: 'test',
        port: 3000,
        logLevel: 'info',
        isDevelopment: false,
        isProduction: false,
        github: {
            appId: 12345,
            privateKey: 'test-key',
            webhookSecret: 'test-secret',
        },
        ai: {
            provider: 'ollama',
            model: 'qwen2.5-coder:14b',
            ollamaBaseUrl: 'http://localhost:11434',
        },
        rateLimit: { rps: 10, burst: 20 },
        cache: { ttl: 3600000, maxEntries: 1000 },
        review: { maxFiles: 50, maxDiffSize: 500000, timeout: 300000 },
    },
}));
import { generateStatusLine, generateStatusBlock, parseExistingStatus, embedStatusData, calculateReviewRound, compareIssues, } from './statusLine.js';
describe('generateStatusLine', () => {
    const baseStatus = {
        totalIssues: 5,
        criticalIssues: 1,
        resolvedIssues: 0,
        persistentIssues: 0,
        score: 75,
        lastReviewAt: new Date().toISOString(),
        reviewRound: 1,
    };
    it('generates status with score and issues', () => {
        const line = generateStatusLine(baseStatus);
        expect(line).toContain('Score: 75/100');
        expect(line).toContain('Issues: 5');
        expect(line).toContain('Critical: 1');
        expect(line).toContain('Round: 1');
    });
    it('shows progress for multiple rounds', () => {
        const status = {
            ...baseStatus,
            reviewRound: 2,
            resolvedIssues: 3,
            persistentIssues: 2,
        };
        const line = generateStatusLine(status);
        expect(line).toContain('Progress: 3/5 resolved (60%)');
        expect(line).toContain('Round: 2');
    });
    it('hides critical when zero', () => {
        const status = {
            ...baseStatus,
            criticalIssues: 0,
        };
        const line = generateStatusLine(status);
        expect(line).not.toContain('Critical:');
    });
    it('respects showEmoji option', () => {
        const line = generateStatusLine(baseStatus, { showEmoji: false });
        expect(line).not.toMatch(/[\u{1F3C6}\u{2705}\u{26A0}\u{FE0F}\u{1F534}]/u);
        expect(line).toContain('Score: 75/100');
    });
    it('shows trophy for high scores', () => {
        const status = { ...baseStatus, score: 95 };
        const line = generateStatusLine(status);
        expect(line).toContain('\u{1F3C6}'); // Trophy emoji
    });
    it('shows warning for medium scores', () => {
        const status = { ...baseStatus, score: 65 };
        const line = generateStatusLine(status);
        expect(line).toContain('\u{26A0}'); // Warning emoji
    });
});
describe('generateStatusBlock', () => {
    const baseStatus = {
        totalIssues: 3,
        criticalIssues: 0,
        resolvedIssues: 0,
        persistentIssues: 0,
        score: 85,
        lastReviewAt: new Date().toISOString(),
        reviewRound: 1,
    };
    it('includes header and status line', () => {
        const block = generateStatusBlock(baseStatus);
        expect(block).toContain('## GoReview Status');
        expect(block).toContain('> ');
        expect(block).toContain('Score: 85/100');
    });
    it('shows progress section for multiple rounds', () => {
        const status = {
            ...baseStatus,
            reviewRound: 2,
            resolvedIssues: 2,
            persistentIssues: 1,
            totalIssues: 4,
        };
        const block = generateStatusBlock(status);
        expect(block).toContain('### Progress Since Last Review');
        expect(block).toContain('2 issue(s) resolved');
        expect(block).toContain('1 issue(s) still pending');
        expect(block).toContain('3 new issue(s) found');
    });
    it('shows critical issues warning', () => {
        const status = {
            ...baseStatus,
            criticalIssues: 2,
        };
        const block = generateStatusBlock(status);
        expect(block).toContain('2 critical issue(s) require immediate attention');
    });
    it('shows inactivity warning after 48 hours', () => {
        const oldDate = new Date();
        oldDate.setHours(oldDate.getHours() - 50);
        const status = {
            ...baseStatus,
            lastReviewAt: oldDate.toISOString(),
        };
        const block = generateStatusBlock(status);
        expect(block).toContain('Warning:');
        expect(block).toContain('pending issues');
    });
    it('shows critical inactivity warning after 72 hours', () => {
        const oldDate = new Date();
        oldDate.setHours(oldDate.getHours() - 75);
        const status = {
            ...baseStatus,
            lastReviewAt: oldDate.toISOString(),
        };
        const block = generateStatusBlock(status);
        expect(block).toContain('Warning:');
        expect(block).toContain('inactive');
        expect(block).toContain('close the PR');
    });
});
describe('parseExistingStatus', () => {
    it('parses embedded status data', () => {
        const status = {
            totalIssues: 5,
            criticalIssues: 1,
            resolvedIssues: 2,
            persistentIssues: 3,
            score: 80,
            lastReviewAt: '2024-01-01T00:00:00Z',
            reviewRound: 2,
        };
        const comment = `## GoReview Status\n\n${embedStatusData(status)}\n\nSome other content`;
        const parsed = parseExistingStatus(comment);
        expect(parsed).toEqual(status);
    });
    it('returns null for comments without status', () => {
        const comment = 'Just a regular comment without status data';
        const parsed = parseExistingStatus(comment);
        expect(parsed).toBeNull();
    });
    it('returns null for invalid JSON', () => {
        const comment = '<!--goreview-status:invalid json-->';
        const parsed = parseExistingStatus(comment);
        expect(parsed).toBeNull();
    });
});
describe('embedStatusData', () => {
    it('creates HTML comment with JSON', () => {
        const status = {
            totalIssues: 3,
            criticalIssues: 0,
            resolvedIssues: 0,
            persistentIssues: 0,
            score: 90,
            lastReviewAt: '2024-01-01T00:00:00Z',
            reviewRound: 1,
        };
        const embedded = embedStatusData(status);
        expect(embedded).toMatch(/^<!--goreview-status:.*-->$/);
        expect(embedded).toContain('"totalIssues":3');
        expect(embedded).toContain('"score":90');
    });
});
describe('calculateReviewRound', () => {
    it('returns 1 for null previous status', () => {
        expect(calculateReviewRound(null)).toBe(1);
    });
    it('returns 1 for status without reviewRound', () => {
        expect(calculateReviewRound({ totalIssues: 5 })).toBe(1);
    });
    it('increments previous round', () => {
        expect(calculateReviewRound({ reviewRound: 1 })).toBe(2);
        expect(calculateReviewRound({ reviewRound: 5 })).toBe(6);
    });
});
describe('compareIssues', () => {
    it('identifies resolved issues', () => {
        const previous = ['issue-1', 'issue-2', 'issue-3'];
        const current = ['issue-2'];
        const result = compareIssues(current, previous);
        expect(result.resolved).toBe(2);
        expect(result.persistent).toBe(1);
    });
    it('handles all resolved', () => {
        const previous = ['issue-1', 'issue-2'];
        const current = [];
        const result = compareIssues(current, previous);
        expect(result.resolved).toBe(2);
        expect(result.persistent).toBe(0);
    });
    it('handles no previous issues', () => {
        const previous = [];
        const current = ['issue-1', 'issue-2'];
        const result = compareIssues(current, previous);
        expect(result.resolved).toBe(0);
        expect(result.persistent).toBe(0);
    });
    it('handles all persistent', () => {
        const previous = ['issue-1', 'issue-2'];
        const current = ['issue-1', 'issue-2', 'issue-3'];
        const result = compareIssues(current, previous);
        expect(result.resolved).toBe(0);
        expect(result.persistent).toBe(2);
    });
});
//# sourceMappingURL=statusLine.test.js.map