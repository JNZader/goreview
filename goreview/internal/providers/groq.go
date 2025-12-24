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
	// Validate input
	if err := ValidateReviewRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Empty diff returns empty response
	if len(req.Diff) == 0 {
		return &ReviewResponse{}, nil
	}

	start := time.Now()

	messages := []map[string]string{
		{"role": "system", "content": "You are an expert code reviewer. Return valid JSON only."},
		{"role": "user", "content": buildReviewPrompt(req)},
	}

	groqReq := map[string]interface{}{
		"model":           p.model,
		"messages":        messages,
		"temperature":     p.config.Temperature,
		"max_tokens":      p.config.MaxTokens,
		"response_format": map[string]string{"type": "json_object"},
	}

	body, err := json.Marshal(groqReq)
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
		return nil, fmt.Errorf("groq request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("groq error: %s", result.Error.Message)
	}

	var reviewResp ReviewResponse
	if len(result.Choices) > 0 {
		if err := json.Unmarshal([]byte(result.Choices[0].Message.Content), &reviewResp); err != nil {
			reviewResp = ReviewResponse{Summary: result.Choices[0].Message.Content}
		}
	}
	reviewResp.TokensUsed = result.Usage.TotalTokens
	reviewResp.ProcessingTime = time.Since(start).Milliseconds()

	return &reviewResp, nil
}

func (p *GroqProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	prompt := fmt.Sprintf(`Generate a conventional commit message for this diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, chore

Diff:
%s

Return ONLY the commit message.`, diff)

	messages := []map[string]string{
		{"role": "user", "content": prompt},
	}

	groqReq := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
	}

	body, err := json.Marshal(groqReq)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from Groq")
}

func (p *GroqProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	prompt := fmt.Sprintf(`Generate documentation for these changes.
Context: %s
Changes:
%s

Format as Markdown.`, docContext, diff)

	messages := []map[string]string{
		{"role": "user", "content": prompt},
	}

	groqReq := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
	}

	body, err := json.Marshal(groqReq)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from Groq")
}

func (p *GroqProvider) HealthCheck(ctx context.Context) error {
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
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("groq health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (p *GroqProvider) Close() error { return nil }
