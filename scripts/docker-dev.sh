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
