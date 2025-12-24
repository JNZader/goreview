package providers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JNZader/goreview/goreview/internal/config"
)

// GroqProvider implements Provider using Groq API.
// Groq uses OpenAI-compatible API format.
type GroqProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
	config  *config.ProviderConfig
}

// NewGroqProvider creates a new Groq provider.
func NewGroqProvider(cfg *config.Config) (*GroqProvider, error) {
	if cfg.Provider.APIKey == "" {
		return nil, fmt.Errorf("Groq API key required (get free at console.groq.com)")
	}

	baseURL := cfg.Provider.BaseURL
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}

	model := cfg.Provider.Model
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}

	return &GroqProvider{
		apiKey:  cfg.Provider.APIKey,
		baseURL: baseURL,
		model:   model,
		config:  &cfg.Provider,
		client:  &http.Client{Timeout: cfg.Provider.Timeout},
	}, nil
}

func (p *GroqProvider) Name() string { return "groq" }

func (p *GroqProvider) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	if empty, err := ValidateReviewInput(req); err != nil {
		return nil, err
	} else if empty {
		return &ReviewResponse{}, nil
	}

	start := time.Now()
	groqReq := BuildChatRequest(p.model, ReviewSystemPrompt, buildReviewPrompt(req), p.config.Temperature, p.config.MaxTokens, true)

	var result ChatCompletionResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+ChatCompletionsPath, groqReq, p.apiKey, &result); err != nil {
		return nil, fmt.Errorf("groq request failed: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("groq error: %s", result.Error.Message)
	}

	return ParseReviewContent(result.GetContent(), result.Usage.TotalTokens, time.Since(start).Milliseconds()), nil
}

func (p *GroqProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	groqReq := map[string]interface{}{
		"model":    p.model,
		"messages": []map[string]string{{"role": "user", "content": fmt.Sprintf(CommitMessagePrompt, diff)}},
	}

	var result ChatCompletionResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+ChatCompletionsPath, groqReq, p.apiKey, &result); err != nil {
		return "", err
	}

	if content := result.GetContent(); content != "" {
		return content, nil
	}
	return "", fmt.Errorf("no response from Groq")
}

func (p *GroqProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	groqReq := map[string]interface{}{
		"model":    p.model,
		"messages": []map[string]string{{"role": "user", "content": fmt.Sprintf(DocumentationPrompt, docContext, diff)}},
	}

	var result ChatCompletionResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+ChatCompletionsPath, groqReq, p.apiKey, &result); err != nil {
		return "", err
	}

	if content := result.GetContent(); content != "" {
		return content, nil
	}
	return "", fmt.Errorf("no response from Groq")
}

func (p *GroqProvider) HealthCheck(ctx context.Context) error {
	return DoHealthCheck(ctx, p.client, p.baseURL+"/models", p.apiKey, "groq")
}

func (p *GroqProvider) Close() error { return nil }
