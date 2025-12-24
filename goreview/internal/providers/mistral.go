package providers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JNZader/goreview/goreview/internal/config"
)

// MistralProvider implements Provider using Mistral API.
type MistralProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
	config  *config.ProviderConfig
}

// NewMistralProvider creates a new Mistral provider.
func NewMistralProvider(cfg *config.Config) (*MistralProvider, error) {
	if cfg.Provider.APIKey == "" {
		return nil, fmt.Errorf("Mistral API key required (get free at console.mistral.ai)")
	}

	baseURL := cfg.Provider.BaseURL
	if baseURL == "" {
		baseURL = "https://api.mistral.ai/v1"
	}

	model := cfg.Provider.Model
	if model == "" {
		model = "codestral-latest"
	}

	return &MistralProvider{
		apiKey:  cfg.Provider.APIKey,
		baseURL: baseURL,
		model:   model,
		config:  &cfg.Provider,
		client:  &http.Client{Timeout: cfg.Provider.Timeout},
	}, nil
}

func (p *MistralProvider) Name() string { return "mistral" }

func (p *MistralProvider) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	if empty, err := ValidateReviewInput(req); err != nil {
		return nil, err
	} else if empty {
		return &ReviewResponse{}, nil
	}

	start := time.Now()
	mistralReq := BuildChatRequest(p.model, ReviewSystemPrompt, buildReviewPrompt(req), p.config.Temperature, p.config.MaxTokens, true)

	var result ChatCompletionResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+ChatCompletionsPath, mistralReq, p.apiKey, &result); err != nil {
		return nil, fmt.Errorf("mistral request failed: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("mistral error: %s", result.Error.Message)
	}

	return ParseReviewContent(result.GetContent(), result.Usage.TotalTokens, time.Since(start).Milliseconds()), nil
}

func (p *MistralProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	mistralReq := map[string]interface{}{
		"model":    p.model,
		"messages": []map[string]string{{"role": "user", "content": fmt.Sprintf(CommitMessagePrompt, diff)}},
	}

	var result ChatCompletionResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+ChatCompletionsPath, mistralReq, p.apiKey, &result); err != nil {
		return "", err
	}

	if content := result.GetContent(); content != "" {
		return content, nil
	}
	return "", fmt.Errorf("no response from Mistral")
}

func (p *MistralProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	mistralReq := map[string]interface{}{
		"model":    p.model,
		"messages": []map[string]string{{"role": "user", "content": fmt.Sprintf(DocumentationPrompt, docContext, diff)}},
	}

	var result ChatCompletionResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+ChatCompletionsPath, mistralReq, p.apiKey, &result); err != nil {
		return "", err
	}

	if content := result.GetContent(); content != "" {
		return content, nil
	}
	return "", fmt.Errorf("no response from Mistral")
}

func (p *MistralProvider) HealthCheck(ctx context.Context) error {
	return DoHealthCheck(ctx, p.client, p.baseURL+"/models", p.apiKey, "mistral")
}

func (p *MistralProvider) Close() error { return nil }
