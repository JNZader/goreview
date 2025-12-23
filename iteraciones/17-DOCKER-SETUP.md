# Iteracion 17: Docker Setup

## Objetivos

- Dockerfile para GoReview CLI
- Dockerfile para GitHub App
- Docker Compose para desarrollo
- Compose con Ollama incluido

## Tiempo Estimado: 4 horas

---

## Commit 17.1: Crear Dockerfile para GoReview

**Mensaje de commit:**
```
build(docker): add goreview dockerfile

- Multi-stage build
- Minimal Alpine image
- Non-root user
- Health check
```

### `goreview/Dockerfile`

```dockerfile
# =============================================================================
# GoReview CLI Dockerfile
# Multi-stage build for minimal image size
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Builder
# -----------------------------------------------------------------------------
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build arguments
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s \
        -X main.Version=${VERSION} \
        -X main.Commit=${COMMIT} \
        -X main.BuildDate=${BUILD_DATE}" \
    -o /build/goreview \
    ./cmd/goreview

# -----------------------------------------------------------------------------
# Stage 2: Runtime
# -----------------------------------------------------------------------------
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1000 goreview \
    && adduser -u 1000 -G goreview -s /bin/sh -D goreview

# Copy binary from builder
COPY --from=builder /build/goreview /usr/local/bin/goreview

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set working directory
WORKDIR /workspace

# Switch to non-root user
USER goreview

# Default command
ENTRYPOINT ["/usr/local/bin/goreview"]
CMD ["--help"]

# Labels
LABEL org.opencontainers.image.title="GoReview CLI"
LABEL org.opencontainers.image.description="AI-powered code review tool"
LABEL org.opencontainers.image.source="https://github.com/TU-USUARIO/ai-toolkit"
```

### `goreview/.dockerignore`

```
# Git
.git
.gitignore

# Build artifacts
bin/
dist/
*.exe
*.test
*.out

# IDE
.idea/
.vscode/
*.swp
*.swo

# Test and coverage
coverage/
*.cover
*.prof

# Documentation
docs/
*.md
!README.md

# Examples
examples/

# CI
.github/
```

---

## Commit 17.2: Crear Dockerfile para GitHub App

**Mensaje de commit:**
```
build(docker): add github-app dockerfile

- Multi-stage Node.js build
- Production dependencies only
- Non-root execution
- Health check configured
```

### `github-app/Dockerfile`

```dockerfile
# =============================================================================
# GitHub App Dockerfile
# Multi-stage build for minimal image
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: Dependencies
# -----------------------------------------------------------------------------
FROM node:20-alpine AS deps

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install all dependencies (including dev for build)
RUN npm ci

# -----------------------------------------------------------------------------
# Stage 2: Builder
# -----------------------------------------------------------------------------
FROM node:20-alpine AS builder

WORKDIR /app

# Copy dependencies
COPY --from=deps /app/node_modules ./node_modules

# Copy source files
COPY . .

# Build TypeScript
RUN npm run build

# Remove dev dependencies
RUN npm prune --production

# -----------------------------------------------------------------------------
# Stage 3: Runtime
# -----------------------------------------------------------------------------
FROM node:20-alpine AS runtime

# Install dumb-init for proper signal handling
RUN apk add --no-cache dumb-init

# Create non-root user
RUN addgroup -g 1001 -S nodejs \
    && adduser -S nodejs -u 1001

WORKDIR /app

# Copy built application
COPY --from=builder --chown=nodejs:nodejs /app/dist ./dist
COPY --from=builder --chown=nodejs:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=nodejs:nodejs /app/package.json ./

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
```

### `github-app/.dockerignore`

```
# Dependencies
node_modules/

# Build output
dist/

# Git
.git
.gitignore

# IDE
.idea/
.vscode/
*.swp

# Environment
.env
.env.local
.env*.local

# Tests
coverage/
*.test.ts
__tests__/

# Documentation
*.md
!README.md

# TypeScript config (not needed at runtime)
tsconfig.json

# CI
.github/
```

---

## Commit 17.3: Crear Docker Compose para desarrollo

**Mensaje de commit:**
```
build(docker): add docker-compose for development

- GitHub App service
- Ollama service
- Shared network
- Volume mounts for development
```

### `docker-compose.yml`

```yaml
# =============================================================================
# AI Toolkit - Docker Compose
# Development environment with all services
# =============================================================================

version: '3.8'

services:
  # ---------------------------------------------------------------------------
  # GitHub App - Webhook server
  # ---------------------------------------------------------------------------
  github-app:
    build:
      context: ./github-app
      target: builder  # Use builder stage for dev (has all deps)
    ports:
      - "3000:3000"
    environment:
      - NODE_ENV=development
      - PORT=3000
      - LOG_LEVEL=debug
      - OLLAMA_BASE_URL=http://ollama:11434
    env_file:
      - ./github-app/.env
    volumes:
      # Mount source for hot reload
      - ./github-app/src:/app/src:ro
      - ./github-app/package.json:/app/package.json:ro
    depends_on:
      ollama:
        condition: service_healthy
    networks:
      - ai-toolkit
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:3000/health"]
      interval: 10s
      timeout: 5s
      retries: 3
    restart: unless-stopped

  # ---------------------------------------------------------------------------
  # Ollama - Local LLM
  # ---------------------------------------------------------------------------
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama-data:/root/.ollama
    environment:
      - OLLAMA_HOST=0.0.0.0
    networks:
      - ai-toolkit
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:11434/api/tags"]
      interval: 30s
      timeout: 10s
      retries: 5
      start_period: 30s
    restart: unless-stopped
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: all
              capabilities: [gpu]

  # ---------------------------------------------------------------------------
  # Ngrok - Tunnel for webhook development
  # ---------------------------------------------------------------------------
  ngrok:
    image: ngrok/ngrok:latest
    command: http github-app:3000 --log stdout
    ports:
      - "4040:4040"  # Ngrok dashboard
    environment:
      - NGROK_AUTHTOKEN=${NGROK_AUTHTOKEN:-}
    depends_on:
      - github-app
    networks:
      - ai-toolkit
    profiles:
      - tunnel  # Only start with: docker compose --profile tunnel up

networks:
  ai-toolkit:
    driver: bridge

volumes:
  ollama-data:
    driver: local
```

---

## Commit 17.4: Agregar scripts de desarrollo

**Mensaje de commit:**
```
build(docker): add docker helper scripts

- Pull required models
- Development startup
- Production build
```

### `scripts/docker-dev.sh`

```bash
#!/bin/bash
# =============================================================================
# Development environment startup script
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_DIR"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Starting AI Toolkit development environment...${NC}"

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo -e "${RED}Error: Docker daemon is not running${NC}"
    exit 1
fi

# Check .env file
if [ ! -f "github-app/.env" ]; then
    echo -e "${YELLOW}Warning: github-app/.env not found${NC}"
    echo "Creating from example..."
    cp github-app/.env.example github-app/.env
    echo -e "${YELLOW}Please edit github-app/.env with your configuration${NC}"
fi

# Start services
echo -e "${GREEN}Starting services...${NC}"
docker compose up -d

# Wait for Ollama
echo -e "${GREEN}Waiting for Ollama to be ready...${NC}"
timeout 60 bash -c 'until curl -s http://localhost:11434/api/tags > /dev/null; do sleep 2; done'

# Pull model if not exists
MODEL="qwen2.5-coder:14b"
echo -e "${GREEN}Checking for model: $MODEL${NC}"

if ! docker compose exec -T ollama ollama list | grep -q "$MODEL"; then
    echo -e "${YELLOW}Pulling model $MODEL (this may take a while)...${NC}"
    docker compose exec -T ollama ollama pull "$MODEL"
fi

echo -e "${GREEN}Development environment ready!${NC}"
echo ""
echo "Services:"
echo "  - GitHub App: http://localhost:3000"
echo "  - Ollama API: http://localhost:11434"
echo ""
echo "To start ngrok tunnel:"
echo "  docker compose --profile tunnel up ngrok"
echo ""
echo "To view logs:"
echo "  docker compose logs -f"
echo ""
echo "To stop:"
echo "  docker compose down"
```

### `scripts/docker-pull-models.sh`

```bash
#!/bin/bash
# =============================================================================
# Pull required Ollama models
# =============================================================================

set -e

# Models to pull
MODELS=(
    "qwen2.5-coder:14b"
    "codellama:13b"
)

echo "Pulling Ollama models..."

for model in "${MODELS[@]}"; do
    echo "Pulling $model..."

    if docker compose exec -T ollama ollama list | grep -q "$model"; then
        echo "  $model already exists, skipping"
    else
        docker compose exec -T ollama ollama pull "$model"
        echo "  $model pulled successfully"
    fi
done

echo "All models ready!"
```

### `Makefile` (actualizar en raiz)

```makefile
# =============================================================================
# AI Toolkit Makefile
# =============================================================================

.PHONY: all build test lint clean docker-build docker-dev docker-stop help

# Default target
all: build

# -----------------------------------------------------------------------------
# Build targets
# -----------------------------------------------------------------------------

build: build-cli build-app

build-cli:
	cd goreview && go build -o ../bin/goreview ./cmd/goreview

build-app:
	cd github-app && npm run build

# -----------------------------------------------------------------------------
# Test targets
# -----------------------------------------------------------------------------

test: test-cli test-app

test-cli:
	cd goreview && go test -v ./...

test-app:
	cd github-app && npm test

# -----------------------------------------------------------------------------
# Lint targets
# -----------------------------------------------------------------------------

lint: lint-cli lint-app

lint-cli:
	cd goreview && golangci-lint run

lint-app:
	cd github-app && npm run lint

# -----------------------------------------------------------------------------
# Docker targets
# -----------------------------------------------------------------------------

docker-build:
	docker compose build

docker-dev:
	./scripts/docker-dev.sh

docker-stop:
	docker compose down

docker-logs:
	docker compose logs -f

docker-shell-app:
	docker compose exec github-app sh

docker-shell-ollama:
	docker compose exec ollama bash

# -----------------------------------------------------------------------------
# Clean targets
# -----------------------------------------------------------------------------

clean:
	rm -rf bin/
	rm -rf goreview/bin/
	rm -rf github-app/dist/
	rm -rf github-app/node_modules/

# -----------------------------------------------------------------------------
# Help
# -----------------------------------------------------------------------------

help:
	@echo "Available targets:"
	@echo ""
	@echo "  build         - Build all components"
	@echo "  test          - Run all tests"
	@echo "  lint          - Run linters"
	@echo "  clean         - Clean build artifacts"
	@echo ""
	@echo "  docker-build  - Build Docker images"
	@echo "  docker-dev    - Start development environment"
	@echo "  docker-stop   - Stop Docker services"
	@echo "  docker-logs   - View Docker logs"
	@echo ""
```

---

## Commit 17.5: Agregar configuracion de produccion

**Mensaje de commit:**
```
build(docker): add production compose config

- Production-ready configuration
- Resource limits
- Logging configuration
- Restart policies
```

### `docker-compose.prod.yml`

```yaml
# =============================================================================
# AI Toolkit - Production Docker Compose
# Use with: docker compose -f docker-compose.yml -f docker-compose.prod.yml up
# =============================================================================

version: '3.8'

services:
  github-app:
    build:
      context: ./github-app
      target: runtime  # Use production runtime stage
    environment:
      - NODE_ENV=production
      - LOG_LEVEL=info
    volumes: []  # No source mounts in production
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 1G
        reservations:
          cpus: '0.5'
          memory: 256M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
        window: 120s
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"

  ollama:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 16G
        reservations:
          cpus: '2'
          memory: 8G
    logging:
      driver: "json-file"
      options:
        max-size: "50m"
        max-file: "5"

# Remove ngrok from production
  ngrok:
    deploy:
      replicas: 0
```

---

## Resumen de la Iteracion 17

### Commits:
1. `build(docker): add goreview dockerfile`
2. `build(docker): add github-app dockerfile`
3. `build(docker): add docker-compose for development`
4. `build(docker): add docker helper scripts`
5. `build(docker): add production compose config`

### Archivos:
```
/
├── docker-compose.yml
├── docker-compose.prod.yml
├── Makefile
├── scripts/
│   ├── docker-dev.sh
│   └── docker-pull-models.sh
├── goreview/
│   ├── Dockerfile
│   └── .dockerignore
└── github-app/
    ├── Dockerfile
    └── .dockerignore
```

---

## Siguiente Iteracion

Continua con: **[18-CI-CD.md](18-CI-CD.md)**
