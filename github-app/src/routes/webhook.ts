import { Router } from 'express';

export const webhookRouter = Router();

// Placeholder - will be implemented in commit 13.5
webhookRouter.post('/', (_req, res) => {
  res.status(200).json({ received: true });
});
