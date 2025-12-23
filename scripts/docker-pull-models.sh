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
