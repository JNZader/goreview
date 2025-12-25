package providers

import (
	"context"
	"fmt"
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
	if empty, err := ValidateReviewInput(req); err != nil {
		return nil, err
	} else if empty {
		return &ReviewResponse{}, nil
	}

	if p.rateLimiter != nil {
		if err := p.rateLimiter.Wait(ctx); err != nil {
			return nil, err
		}
	}

	start := time.Now()
	ollamaReq := BuildOllamaRequest(p.model, buildReviewPrompt(req), p.config.Temperature, p.config.MaxTokens, true)

	var result OllamaResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+APIGeneratePath, ollamaReq, "", &result); err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}

	return ParseReviewContent(result.Response, 0, time.Since(start).Milliseconds()), nil
}

func (p *OllamaProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	ollamaReq := map[string]interface{}{
		"model": p.model, "prompt": fmt.Sprintf(CommitMessagePrompt, diff), "stream": false,
	}

	var result OllamaResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+APIGeneratePath, ollamaReq, "", &result); err != nil {
		return "", err
	}
	return result.Response, nil
}

func (p *OllamaProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	ollamaReq := map[string]interface{}{
		"model": p.model, "prompt": fmt.Sprintf(DocumentationPrompt, docContext, diff), "stream": false,
	}

	var result OllamaResponse
	if err := DoJSONPost(ctx, p.client, p.baseURL+APIGeneratePath, ollamaReq, "", &result); err != nil {
		return "", err
	}
	return result.Response, nil
}

func (p *OllamaProvider) HealthCheck(ctx context.Context) error {
	return DoHealthCheck(ctx, p.client, p.baseURL+"/api/tags", "", "ollama")
}

func (p *OllamaProvider) Close() error { return nil }

func buildReviewPrompt(req *ReviewRequest) string {
	personalityPrompt := GetPersonalityPrompt(req.Personality)
	modePrompt := CombineModePrompts(req.Modes)

	return fmt.Sprintf(`%s

%s

File: %s
Language: %s

Code:
%s

Return a JSON object:
{
  "issues": [{"id": "1", "type": "bug|security|performance|style", "severity": "info|warning|error|critical", "message": "description", "suggestion": "how to fix"}],
  "summary": "brief summary",
  "score": 85
}`, personalityPrompt, modePrompt, req.FilePath, req.Language, req.Diff)
}
