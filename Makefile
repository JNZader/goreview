# =============================================================================
# AI Toolkit Makefile
# =============================================================================

.DEFAULT_GOAL := help
.PHONY: all build test lint clean docker-build docker-dev docker-stop help

# Default target
all: build

# -----------------------------------------------------------------------------
# Build targets
# -----------------------------------------------------------------------------

build: build-cli build-app ## Build all components

build-cli: ## Build GoReview CLI
	cd goreview && go build -o ../bin/goreview ./cmd/goreview

build-app: ## Build GitHub App
	cd github-app && npm run build

# -----------------------------------------------------------------------------
# Test targets
# -----------------------------------------------------------------------------

test: test-cli test-app ## Run all tests

test-cli: ## Test GoReview CLI
	cd goreview && go test -v ./...

test-app: ## Test GitHub App
	cd github-app && npm test

# -----------------------------------------------------------------------------
# Lint targets
# -----------------------------------------------------------------------------

lint: lint-cli lint-app ## Lint all code

lint-cli: ## Lint GoReview CLI
	cd goreview && golangci-lint run

lint-app: ## Lint GitHub App
	cd github-app && npm run lint

# -----------------------------------------------------------------------------
# Docker targets
# -----------------------------------------------------------------------------

docker-build: ## Build Docker images
	docker compose build

docker-dev: ## Start development environment
	./scripts/docker-dev.sh

docker-stop: ## Stop Docker services
	docker compose down

docker-logs: ## View Docker logs
	docker compose logs -f

docker-shell-app: ## Shell into GitHub App container
	docker compose exec github-app sh

docker-shell-ollama: ## Shell into Ollama container
	docker compose exec ollama bash

docker-pull-models: ## Pull Ollama models
	./scripts/docker-pull-models.sh

# -----------------------------------------------------------------------------
# Development
# -----------------------------------------------------------------------------

setup: ## Setup development environment
	@echo "Setting up development environment..."
	@echo ""
	@echo "1. Installing Go dependencies..."
	cd goreview && go mod download
	@echo ""
	@echo "2. Installing Node.js dependencies..."
	cd github-app && npm install
	@echo ""
	@echo "3. Installing development tools..."
	go install github.com/golangci-lint/golangci-lint/cmd/golangci-lint@latest
	@echo ""
	@echo "Setup complete!"

ollama-pull: ## Pull Ollama model
	ollama pull qwen2.5-coder:14b

ollama-serve: ## Start Ollama server
	ollama serve

# -----------------------------------------------------------------------------
# Clean targets
# -----------------------------------------------------------------------------

clean: ## Clean build artifacts
	rm -rf bin/
	rm -rf goreview/bin/
	rm -rf github-app/dist/
	rm -rf github-app/node_modules/

# -----------------------------------------------------------------------------
# Help
# -----------------------------------------------------------------------------

help: ## Show this help
	@echo "AI Toolkit Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
