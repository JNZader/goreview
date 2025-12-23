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
LABEL org.opencontainers.image.source="https://github.com/JNZader/goreview"
