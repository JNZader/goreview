import { describe, it, expect, vi, beforeEach } from 'vitest';
import crypto from 'crypto';
import { verifyWebhookSignature } from './webhookVerify.js';
// Mock the config module
vi.mock('../config/index.js', () => ({
    config: {
        github: {
            webhookSecret: 'test-secret',
        },
    },
}));
describe('verifyWebhookSignature', () => {
    const secret = 'test-secret';
    function createSignature(payload) {
        const hmac = crypto.createHmac('sha256', secret);
        return 'sha256=' + hmac.update(payload).digest('hex');
    }
    beforeEach(() => {
        vi.clearAllMocks();
    });
    it('should return false when signature is undefined', () => {
        const payload = Buffer.from('test payload');
        const result = verifyWebhookSignature(payload, undefined);
        expect(result).toBe(false);
    });
    it('should return false when signature is empty string', () => {
        const payload = Buffer.from('test payload');
        const result = verifyWebhookSignature(payload, '');
        expect(result).toBe(false);
    });
    it('should return true for valid signature with Buffer payload', () => {
        const payload = '{"action":"opened"}';
        const signature = createSignature(payload);
        const result = verifyWebhookSignature(Buffer.from(payload), signature);
        expect(result).toBe(true);
    });
    it('should return true for valid signature with string payload', () => {
        const payload = '{"action":"opened"}';
        const signature = createSignature(payload);
        const result = verifyWebhookSignature(payload, signature);
        expect(result).toBe(true);
    });
    it('should return false for invalid signature', () => {
        const payload = Buffer.from('test payload');
        const invalidSignature = 'sha256=' + 'a'.repeat(64);
        const result = verifyWebhookSignature(payload, invalidSignature);
        expect(result).toBe(false);
    });
    it('should return false for signature with wrong length', () => {
        const payload = Buffer.from('test payload');
        const shortSignature = 'sha256=abc123';
        const result = verifyWebhookSignature(payload, shortSignature);
        expect(result).toBe(false);
    });
    it('should return false for tampered payload', () => {
        const originalPayload = '{"action":"opened"}';
        const signature = createSignature(originalPayload);
        const tamperedPayload = '{"action":"closed"}';
        const result = verifyWebhookSignature(Buffer.from(tamperedPayload), signature);
        expect(result).toBe(false);
    });
});
//# sourceMappingURL=webhookVerify.test.js.map