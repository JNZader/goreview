package providers

import (
	"fmt"

	"github.com/JNZader/goreview/goreview/internal/config"
)

// NewProvider creates a new Provider based on configuration.
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.Provider.Name {
	case "ollama":
		return NewOllamaProvider(cfg)
	case "openai":
		return NewOpenAIProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider.Name)
	}
}

// AvailableProviders returns a list of available provider names.
func AvailableProviders() []string {
	return []string{"ollama", "openai"}
}

// NewOpenAIProvider creates a new OpenAI provider.
// Placeholder - will be implemented in commit 4.5.
func NewOpenAIProvider(cfg *config.Config) (Provider, error) {
	return nil, fmt.Errorf("openai provider not implemented yet")
}
