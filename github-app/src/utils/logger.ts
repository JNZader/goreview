// Placeholder - will be implemented in commit 13.3
export const logger = {
  info: (...args: unknown[]) => console.log('[INFO]', ...args),
  error: (...args: unknown[]) => console.error('[ERROR]', ...args),
  warn: (...args: unknown[]) => console.warn('[WARN]', ...args),
  debug: (...args: unknown[]) => console.log('[DEBUG]', ...args),
};
