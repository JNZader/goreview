export type WebhookPayload = Record<string, unknown>;
/**
 * Route webhook events to appropriate handlers.
 */
export declare function handleWebhook(event: string, payload: WebhookPayload): Promise<void>;
//# sourceMappingURL=webhookHandler.d.ts.map