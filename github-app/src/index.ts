import express from 'express';
import { config } from './config/index.js';
import { logger } from './utils/logger.js';
import { errorHandler } from './middleware/errorHandler.js';
import { requestLogger } from './middleware/requestLogger.js';
import { webhookRouter } from './routes/webhook.js';
import { healthRouter } from './routes/health.js';
import { adminRouter } from './routes/admin.js';

const app = express();

// Trust proxy for accurate IP logging
app.set('trust proxy', 1);

// Raw body for webhook signature verification
app.use('/webhook', express.raw({ type: 'application/json' }));

// JSON parsing for other routes
app.use(express.json());

// Request logging
app.use(requestLogger);

// Routes
app.use('/health', healthRouter);
app.use('/webhook', webhookRouter);
app.use('/admin', adminRouter);

// Error handling
app.use(errorHandler);

// Start server
const server = app.listen(config.port, () => {
  logger.info({ port: config.port }, 'Server started');
});

// Graceful shutdown
const shutdown = async () => {
  logger.info('Shutting down gracefully...');
  server.close(() => {
    logger.info('Server closed');
    process.exit(0);
  });

  // Force close after 10s
  setTimeout(() => {
    logger.error('Forced shutdown');
    process.exit(1);
  }, 10000);
};

process.on('SIGTERM', shutdown);
process.on('SIGINT', shutdown);

export { app };
