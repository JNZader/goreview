package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JNZader/goreview/goreview/internal/config"
)

// OllamaProvider implements Provider using Ollama.
type OllamaProvider struct {
	baseURL     string
	model       string
	client      *http.Client
	config      *config.ProviderConfig
	rateLimiter *RateLimiter
}

// NewOllamaProvider creates a new Ollama provider.
func NewOllamaProvider(cfg *config.Config) (*OllamaProvider, error) {
	var limiter *RateLimiter
	if cfg.Provider.RateLimitRPS > 0 {
		limiter = NewRateLimiter(cfg.Provider.RateLimitRPS)
	}

	return &OllamaProvider{
		baseURL: cfg.Provider.BaseURL,
		model:   cfg.Provider.Model,
		config:  &cfg.Provider,
		client: &http.Client{
			Timeout: cfg.Provider.Timeout,
		},
		rateLimiter: limiter,
	}, nil
}

func (p *OllamaProvider) Name() string { return "ollama" }

func (p *OllamaProvider) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	// Validate input
	if err := ValidateReviewRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Empty diff returns empty response
	if len(req.Diff) == 0 {
		return &ReviewResponse{}, nil
	}

	if p.rateLimiter != nil {
		if err := p.rateLimiter.Wait(ctx); err != nil {
			return nil, err
		}
	}

	start := time.Now()
	prompt := buildReviewPrompt(req)

	ollamaReq := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
		"format": "json",
		"options": map[string]interface{}{
			"temperature": p.config.Temperature,
			"num_predict": p.config.MaxTokens,
		},
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf(ErrMarshalRequest, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+APIGeneratePath, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf(ErrCreateRequest, err)
	}
	httpReq.Header.Set("Content-Type", ContentTypeJSON)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body) //nolint:errcheck // best effort for error message
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, bodyBytes)
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	var reviewResp ReviewResponse
	if err := json.Unmarshal([]byte(ollamaResp.Response), &reviewResp); err != nil {
		// Fallback for non-JSON response
		reviewResp = ReviewResponse{Summary: ollamaResp.Response}
	}
	reviewResp.ProcessingTime = time.Since(start).Milliseconds()

	return &reviewResp, nil
}

func (p *OllamaProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	prompt := fmt.Sprintf(`Generate a conventional commit message for this diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, chore

Diff:
%s

Return ONLY the commit message.`, diff)

	ollamaReq := map[string]interface{}{
		"model": p.model, "prompt": prompt, "stream": false,
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return "", fmt.Errorf(ErrMarshalRequest, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+APIGeneratePath, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf(ErrCreateRequest, err)
	}
	httpReq.Header.Set("Content-Type", ContentTypeJSON)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf(ErrDecodeResponse, err)
	}
	return result.Response, nil
}

func (p *OllamaProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	prompt := fmt.Sprintf(`Generate documentation for these changes.
Context: %s
Changes:
%s

Format as Markdown.`, docContext, diff)

	ollamaReq := map[string]interface{}{
		"model": p.model, "prompt": prompt, "stream": false,
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return "", fmt.Errorf(ErrMarshalRequest, err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+APIGeneratePath, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf(ErrCreateRequest, err)
	}
	httpReq.Header.Set("Content-Type", ContentTypeJSON)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf(ErrDecodeResponse, err)
	}
	return result.Response, nil
}

func (p *OllamaProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf(ErrCreateRequest, err)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (p *OllamaProvider) Close() error { return nil }

func buildReviewPrompt(req *ReviewRequest) string {
	return fmt.Sprintf(`You are an expert code reviewer. Analyze this code and identify issues.

File: %s
Language: %s

Code:
%s

Return a JSON object:
{
  "issues": [{"id": "1", "type": "bug|security|performance|style", "severity": "info|warning|error|critical", "message": "description", "suggestion": "how to fix"}],
  "summary": "brief summary",
  "score": 85
}

Only report real issues, not nitpicks.`, req.FilePath, req.Language, req.Diff)
}
