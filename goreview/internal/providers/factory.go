package providers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/JNZader/goreview/goreview/internal/config"
)

// NewProvider creates a new Provider based on configuration.
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.Provider.Name {
	case "ollama":
		return NewOllamaProvider(cfg)
	case "openai":
		return NewOpenAIProvider(cfg)
	case "gemini":
		return NewGeminiProvider(cfg)
	case "groq":
		return NewGroqProvider(cfg)
	case "mistral":
		return NewMistralProvider(cfg)
	case "fallback":
		return NewFallbackFromEnv()
	case "auto", "":
		return NewAutoProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider.Name)
	}
}

// NewAutoProvider automatically selects the best provider based on environment.
// - If Ollama is running locally -> use Ollama (fast, free, offline)
// - If cloud API keys exist -> use Fallback chain
// - Otherwise -> error
func NewAutoProvider(cfg *config.Config) (Provider, error) {
	// Check if we're in CI/GitHub Actions environment
	if os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("CI") == "true" {
		log.Println("[auto] CI environment detected, using cloud providers")
		return NewFallbackFromEnv()
	}

	// Check if Ollama is running locally
	if isOllamaRunning() {
		log.Println("[auto] Ollama detected locally, using Ollama")
		if cfg.Provider.BaseURL == "" {
			cfg.Provider.BaseURL = "http://localhost:11434"
		}
		if cfg.Provider.Model == "" {
			cfg.Provider.Model = "qwen2.5-coder:14b"
		}
		return NewOllamaProvider(cfg)
	}

	// Check if any cloud API keys are set
	if hasCloudAPIKeys() {
		log.Println("[auto] Cloud API keys found, using fallback chain")
		return NewFallbackFromEnv()
	}

	return nil, fmt.Errorf("no provider available. Either:\n" +
		"  1. Start Ollama locally: ollama serve\n" +
		"  2. Set API keys: GEMINI_API_KEY, GROQ_API_KEY, or MISTRAL_API_KEY")
}

// isOllamaRunning checks if Ollama is running on localhost.
func isOllamaRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}

// hasCloudAPIKeys checks if any cloud provider API keys are set.
func hasCloudAPIKeys() bool {
	keys := []string{"GEMINI_API_KEY", "GROQ_API_KEY", "MISTRAL_API_KEY", "OPENAI_API_KEY"}
	for _, key := range keys {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}

// NewFallbackFromEnv creates a fallback provider chain from environment variables.
// Priority: Gemini (quality) -> Groq (speed) -> Mistral (code) -> OpenAI (paid)
func NewFallbackFromEnv() (Provider, error) {
	var providers []Provider

	// Try Gemini first (best quality, free)
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		cfg := &config.Config{
			Provider: config.ProviderConfig{
				APIKey:      key,
				Temperature: 0.1,
				MaxTokens:   4096,
				Timeout:     60 * time.Second,
			},
		}
		if p, err := NewGeminiProvider(cfg); err == nil {
			providers = append(providers, p)
			log.Println("[fallback] Added Gemini provider")
		}
	}

	// Try Groq second (fastest, free)
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		cfg := &config.Config{
			Provider: config.ProviderConfig{
				APIKey:      key,
				Temperature: 0.1,
				MaxTokens:   4096,
				Timeout:     30 * time.Second,
			},
		}
		if p, err := NewGroqProvider(cfg); err == nil {
			providers = append(providers, p)
			log.Println("[fallback] Added Groq provider")
		}
	}

	// Try Mistral third (code-specialized, free)
	if key := os.Getenv("MISTRAL_API_KEY"); key != "" {
		cfg := &config.Config{
			Provider: config.ProviderConfig{
				APIKey:      key,
				Temperature: 0.1,
				MaxTokens:   4096,
				Timeout:     60 * time.Second,
			},
		}
		if p, err := NewMistralProvider(cfg); err == nil {
			providers = append(providers, p)
			log.Println("[fallback] Added Mistral provider")
		}
	}

	// Try OpenAI last (paid, but reliable)
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		cfg := &config.Config{
			Provider: config.ProviderConfig{
				APIKey:      key,
				Model:       "gpt-4",
				Temperature: 0.1,
				MaxTokens:   4096,
				Timeout:     60 * time.Second,
			},
		}
		if p, err := NewOpenAIProvider(cfg); err == nil {
			providers = append(providers, p)
			log.Println("[fallback] Added OpenAI provider")
		}
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no API keys found. Set one of: GEMINI_API_KEY, GROQ_API_KEY, MISTRAL_API_KEY, OPENAI_API_KEY")
	}

	return NewFallbackProvider(providers...)
}

// AvailableProviders returns a list of available provider names.
func AvailableProviders() []string {
	return []string{"ollama", "openai", "gemini", "groq", "mistral", "fallback", "auto"}
}
