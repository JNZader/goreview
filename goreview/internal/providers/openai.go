package providers

import (
	"bytes"
	"context"
	"encoding/json"
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
	start := time.Now()

	messages := []map[string]string{
		{"role": "system", "content": "You are an expert code reviewer. Return JSON."},
		{"role": "user", "content": buildReviewPrompt(req)},
	}

	openaiReq := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"temperature": p.config.Temperature,
		"max_tokens":  p.config.MaxTokens,
	}

	body, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var reviewResp ReviewResponse
	if len(result.Choices) > 0 {
		// Ignore unmarshal error - fallback to empty response is acceptable
		_ = json.Unmarshal([]byte(result.Choices[0].Message.Content), &reviewResp)
	}
	reviewResp.TokensUsed = result.Usage.TotalTokens
	reviewResp.ProcessingTime = time.Since(start).Milliseconds()

	return &reviewResp, nil
}

func (p *OpenAIProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (p *OpenAIProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

func (p *OpenAIProvider) Close() error { return nil }
