import crypto from 'node:crypto';
import { config } from '../config/index.js';
/**
 * Verify the webhook signature from GitHub.
 */
export function verifyWebhookSignature(payload, signature) {
    if (!signature) {
        return false;
    }
    const sig = Buffer.from(signature);
    const body = typeof payload === 'string' ? payload : payload.toString();
    const hmac = crypto.createHmac('sha256', config.github.webhookSecret);
    const digest = Buffer.from('sha256=' + hmac.update(body).digest('hex'));
    if (sig.length !== digest.length) {
        return false;
    }
    return crypto.timingSafeEqual(digest, sig);
}
//# sourceMappingURL=webhookVerify.js.map