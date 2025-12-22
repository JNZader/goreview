# ===========================================
# AI-Toolkit Root Makefile
# ===========================================

.DEFAULT_GOAL := help

# ===========================================
# Setup
# ===========================================

.PHONY: setup
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@echo ""
	@echo "1. Installing Go dependencies..."
	cd goreview && go mod download
	@echo ""
	@echo "2. Installing Node.js dependencies..."
	cd integrations/github-app && npm install
	@echo ""
	@echo "3. Installing development tools..."
	go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
	go install github.com/cosmtrek/air@latest
	@echo ""
	@echo "4. Pulling Ollama model..."
	ollama pull qwen2.5-coder:14b || echo "Ollama not available, skipping..."
	@echo ""
	@echo "Setup complete!"

# ===========================================
# Build
# ===========================================

.PHONY: build
build: ## Build all components
	@echo "Building GoReview CLI..."
	cd goreview && $(MAKE) build
	@echo ""
	@echo "Building GitHub App..."
	cd integrations/github-app && npm run build

.PHONY: build-goreview
build-goreview: ## Build only GoReview CLI
	cd goreview && $(MAKE) build

.PHONY: build-github-app
build-github-app: ## Build only GitHub App
	cd integrations/github-app && npm run build

# ===========================================
# Test
# ===========================================

.PHONY: test
test: ## Run all tests
	@echo "Testing GoReview CLI..."
	cd goreview && $(MAKE) test
	@echo ""
	@echo "Testing GitHub App..."
	cd integrations/github-app && npm test

.PHONY: test-goreview
test-goreview: ## Test only GoReview CLI
	cd goreview && $(MAKE) test

.PHONY: test-github-app
test-github-app: ## Test only GitHub App
	cd integrations/github-app && npm test

# ===========================================
# Lint
# ===========================================

.PHONY: lint
lint: ## Lint all code
	@echo "Linting GoReview CLI..."
	cd goreview && $(MAKE) lint
	@echo ""
	@echo "Linting GitHub App..."
	cd integrations/github-app && npm run lint

# ===========================================
# Docker
# ===========================================

.PHONY: docker-up
docker-up: ## Start all services with Docker Compose
	docker compose up -d

.PHONY: docker-down
docker-down: ## Stop all services
	docker compose down

.PHONY: docker-logs
docker-logs: ## Show logs
	docker compose logs -f

.PHONY: docker-build
docker-build: ## Build Docker images
	docker compose build

# ===========================================
# Development
# ===========================================

.PHONY: dev
dev: ## Start development mode
	@echo "Starting development mode..."
	@echo "GoReview: make -C goreview dev"
	@echo "GitHub App: cd integrations/github-app && npm run dev"

.PHONY: ollama-pull
ollama-pull: ## Pull Ollama model
	ollama pull qwen2.5-coder:14b

.PHONY: ollama-serve
ollama-serve: ## Start Ollama server
	ollama serve

# ===========================================
# Clean
# ===========================================

.PHONY: clean
clean: ## Clean all build artifacts
	cd goreview && $(MAKE) clean
	cd integrations/github-app && rm -rf dist node_modules
	rm -rf .cache logs

# ===========================================
# Help
# ===========================================

.PHONY: help
help: ## Show this help
	@echo "AI-Toolkit Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
