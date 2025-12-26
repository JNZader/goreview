# =============================================================================
# GitHub App Dockerfile
# Multi-stage build for minimal image
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Dependencies
# -----------------------------------------------------------------------------
FROM node:20-alpine AS deps

WORKDIR /app

# Copy package files (explicit, no glob patterns for security - S6470)
COPY package.json ./
COPY package-lock.json ./

# Install all dependencies (including dev for build)
RUN npm ci

# -----------------------------------------------------------------------------
# Stage 2: Builder
# -----------------------------------------------------------------------------
FROM node:20-alpine AS builder

WORKDIR /app

# Copy dependencies
COPY --from=deps /app/node_modules ./node_modules

# Copy only required source files (explicit, no recursive copy for security - S6470)
COPY src/ ./src/
COPY package.json tsconfig.json ./

# Build TypeScript
RUN npm run build

# Remove dev dependencies
RUN npm prune --production

# -----------------------------------------------------------------------------
# Stage 3: Runtime
# -----------------------------------------------------------------------------
FROM node:20-alpine AS runtime

# Upgrade packages, install dumb-init, and create non-root user
RUN apk upgrade --no-cache \
    && apk add --no-cache dumb-init \
    && addgroup -g 1001 -S nodejs \
    && adduser -S nodejs -u 1001

WORKDIR /app

# Copy built application (no --chown for read-only security - S6504)
# Files owned by root, app runs as non-root with read-only access
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./

# Set environment
ENV NODE_ENV=production
ENV PORT=3000

# Switch to non-root user
USER nodejs

# Expose port
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

# Start with dumb-init
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["node", "dist/index.js"]

# Labels
LABEL org.opencontainers.image.title="GoReview GitHub App"
LABEL org.opencontainers.image.description="GitHub App for AI code reviews"
LABEL org.opencontainers.image.source="https://github.com/JNZader/goreview"
