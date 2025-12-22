# ===========================================
# GoReview Makefile
# ===========================================

# Variables
BINARY_NAME=goreview
BUILD_DIR=build
CMD_DIR=cmd/goreview
MAIN_FILE=$(CMD_DIR)/main.go

# Version info (se sobreescribe en CI)
VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go flags
GOFLAGS=-trimpath
LDFLAGS=-s -w \
	-X 'github.com/JNZader/goreview/goreview/cmd/goreview/commands.Version=$(VERSION)' \
	-X 'github.com/JNZader/goreview/goreview/cmd/goreview/commands.Commit=$(COMMIT)' \
	-X 'github.com/JNZader/goreview/goreview/cmd/goreview/commands.BuildDate=$(BUILD_DATE)'

# Default target
.DEFAULT_GOAL := help

# ===========================================
# Build Targets
# ===========================================

.PHONY: build
build: ## Build the binary
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)
	GOOS=darwin GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	GOOS=windows GOARCH=amd64 go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "All binaries built in $(BUILD_DIR)/"

.PHONY: install
install: build ## Install to GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# ===========================================
# Test Targets
# ===========================================

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

.PHONY: test-short
test-short: ## Run tests (short mode)
	go test -v -short ./...

.PHONY: coverage
coverage: test ## Generate coverage report
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: benchmark
benchmark: ## Run benchmarks
	go test -bench=. -benchmem ./...

# ===========================================
# Lint Targets
# ===========================================

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: lint-fix
lint-fix: ## Run linter with auto-fix
	golangci-lint run --fix --timeout=5m

.PHONY: fmt
fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

.PHONY: vet
vet: ## Run go vet
	go vet ./...

# ===========================================
# Development Targets
# ===========================================

.PHONY: run
run: build ## Build and run
	./$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

.PHONY: dev
dev: ## Run with hot reload (requires air)
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Install with:"; \
		echo "  go install github.com/cosmtrek/air@latest"; \
		echo "Running without hot reload..."; \
		$(MAKE) run; \
	fi

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

.PHONY: deps-update
deps-update: ## Update dependencies
	go get -u ./...
	go mod tidy

.PHONY: generate
generate: ## Run go generate
	go generate ./...

# ===========================================
# Clean Targets
# ===========================================

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache -testcache

.PHONY: clean-all
clean-all: clean ## Clean everything including deps
	go clean -modcache

# ===========================================
# Docker Targets
# ===========================================

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(BINARY_NAME):$(VERSION) -f Dockerfile.production .

.PHONY: docker-run
docker-run: ## Run in Docker
	docker run --rm -it $(BINARY_NAME):$(VERSION)

# ===========================================
# Release Targets
# ===========================================

.PHONY: release-dry
release-dry: ## Dry run of release
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
	else \
		echo "goreleaser not installed"; \
	fi

# ===========================================
# Help
# ===========================================

.PHONY: help
help: ## Show this help
	@echo "GoReview Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
