package providers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JNZader/goreview/goreview/internal/config"
)

// OpenAIProvider implements Provider using OpenAI API.
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
	config  *config.ProviderConfig
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider(cfg *config.Config) (*OpenAIProvider, error) {
	if cfg.Provider.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key required")
	}

	baseURL := cfg.Provider.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIProvider{
		apiKey:  cfg.Provider.APIKey,
		baseURL: baseURL,
		model:   cfg.Provider.Model,
		config:  &cfg.Provider,
		client:  &http.Client{Timeout: cfg.Provider.Timeout},
	}, nil
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	if empty, err := ValidateReviewInput(req); err != nil {
		return nil, err
	} else if empty {
		return &ReviewResponse{}, nil
	}

	start := time.Now()
	openaiReq := BuildChatRequest(p.model, ReviewSystemPrompt, buildReviewPrompt(req), p.config.Temperature, p.config.MaxTokens, false)

	var result ChatCompletionResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+ChatCompletionsPath, openaiReq, p.apiKey, &result); err != nil {
		return nil, err
	}

	return ParseReviewContent(result.GetContent(), result.Usage.TotalTokens, time.Since(start).Milliseconds()), nil
}

func (p *OpenAIProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (p *OpenAIProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	return DoHealthCheck(ctx, p.client, p.baseURL+"/models", p.apiKey, "openai")
}

func (p *OpenAIProvider) Close() error { return nil }
