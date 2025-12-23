import pino from 'pino';
import { config } from '../config/index.js';

export const logger = pino({
  level: config.logLevel,
  transport: config.isDevelopment
    ? {
        target: 'pino-pretty',
        options: {
          colorize: true,
          translateTime: 'SYS:standard',
          ignore: 'pid,hostname',
        },
      }
    : undefined,
  base: {
    env: config.nodeEnv,
  },
  redact: ['req.headers.authorization', 'req.headers["x-hub-signature-256"]'],
});

export type Logger = typeof logger;
