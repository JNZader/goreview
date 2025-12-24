import { describe, it, expect, vi, beforeEach } from 'vitest';
import express from 'express';
import request from 'supertest';
import { healthRouter } from './health.js';
// Mock the config
vi.mock('../config/index.js', () => ({
    config: {
        github: {
            appId: '12345',
            privateKey: 'test-key',
            webhookSecret: 'test-secret',
        },
        port: 3000,
        nodeEnv: 'test',
        logLevel: 'info',
    },
}));
// Mock the AI provider
const mockHealthCheck = vi.fn();
vi.mock('../services/ai/index.js', () => ({
    getProvider: () => ({
        healthCheck: mockHealthCheck,
    }),
}));
// Mock the logger
vi.mock('../utils/logger.js', () => ({
    logger: {
        warn: vi.fn(),
        error: vi.fn(),
        info: vi.fn(),
    },
}));
describe('Health Routes', () => {
    const app = express();
    app.use('/health', healthRouter);
    beforeEach(() => {
        vi.clearAllMocks();
    });
    describe('GET /health', () => {
        it('should return 200 with status ok and metadata', async () => {
            const response = await request(app).get('/health');
            expect(response.status).toBe(200);
            expect(response.body).toHaveProperty('status', 'ok');
            expect(response.body).toHaveProperty('timestamp');
            expect(response.body).toHaveProperty('version');
        });
        it('should return valid ISO timestamp', async () => {
            const response = await request(app).get('/health');
            const timestamp = new Date(response.body.timestamp);
            expect(timestamp.toISOString()).toBe(response.body.timestamp);
        });
    });
    describe('GET /health/ready', () => {
        it('should return 200 when all services are healthy', async () => {
            mockHealthCheck.mockResolvedValue(true);
            const response = await request(app).get('/health/ready');
            expect(response.status).toBe(200);
            expect(response.body).toHaveProperty('status', 'ready');
            expect(response.body).toHaveProperty('checks');
            expect(response.body.checks).toHaveProperty('server', true);
            expect(response.body.checks).toHaveProperty('ai_provider', true);
        });
        it('should return 503 when AI provider is unhealthy', async () => {
            mockHealthCheck.mockResolvedValue(false);
            const response = await request(app).get('/health/ready');
            expect(response.status).toBe(503);
            expect(response.body).toHaveProperty('status', 'degraded');
            expect(response.body.checks).toHaveProperty('server', true);
            expect(response.body.checks).toHaveProperty('ai_provider', false);
        });
        it('should return 503 when AI provider throws error', async () => {
            mockHealthCheck.mockRejectedValue(new Error('Connection failed'));
            const response = await request(app).get('/health/ready');
            expect(response.status).toBe(503);
            expect(response.body).toHaveProperty('status', 'degraded');
            expect(response.body.checks).toHaveProperty('ai_provider', false);
        });
    });
});
//# sourceMappingURL=health.test.js.map