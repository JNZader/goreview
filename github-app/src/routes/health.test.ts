import { describe, it, expect, vi } from 'vitest';
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

describe('Health Routes', () => {
  const app = express();
  app.use('/health', healthRouter);

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
    it('should return 200 when service is ready', async () => {
      const response = await request(app).get('/health/ready');

      expect(response.status).toBe(200);
      expect(response.body).toHaveProperty('status', 'ready');
      expect(response.body).toHaveProperty('checks');
      expect(response.body.checks).toHaveProperty('github', true);
      expect(response.body.checks).toHaveProperty('ollama', true);
    });
  });
});
