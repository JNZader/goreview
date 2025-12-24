import express from 'express';
import { config } from './config/index.js';
import { logger } from './utils/logger.js';
import { errorHandler } from './middleware/errorHandler.js';
import { requestLogger } from './middleware/requestLogger.js';
import { requestIdMiddleware } from './middleware/requestId.js';
import { securityHeaders } from './middleware/security.js';
import { webhookRouter } from './routes/webhook.js';
import { healthRouter } from './routes/health.js';
import { adminRouter } from './routes/admin.js';
import { initializeQueue } from './queue/index.js';

const app = express();

// Disable X-Powered-By header to prevent fingerprinting
app.disable('x-powered-by');

// Trust proxy for accurate IP logging
app.set('trust proxy', 1);

// Security headers
app.use(securityHeaders());

// Request ID tracing
app.use(requestIdMiddleware);

// Raw body for webhook signature verification (with size limit)
app.use('/webhook', express.raw({ type: 'application/json', limit: '1mb' }));

// JSON parsing for other routes (with size limit)
app.use(express.json({ limit: '500kb' }));

// Request logging
app.use(requestLogger);

// Routes
app.use('/health', healthRouter);
app.use('/webhook', webhookRouter);
app.use('/admin', adminRouter);

// Error handling
app.use(errorHandler);

// Initialize queue and start server
const startServer = async () => {
  try {
    await initializeQueue();
  } catch (error) {
    logger.warn({ error }, 'Queue initialization failed, continuing without persistent queue');
  }

  const server = app.listen(config.port, () => {
    logger.info({ port: config.port }, 'Server started');
  });

  return server;
};

const serverPromise = startServer();

// Graceful shutdown
const shutdown = async () => {
  logger.info('Shutting down gracefully...');
  const server = await serverPromise;
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
