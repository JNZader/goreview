package providers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/JNZader/goreview/goreview/internal/config"
)

// GeminiProvider implements Provider using Google Gemini API.
type GeminiProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
	config  *config.ProviderConfig
}

// NewGeminiProvider creates a new Gemini provider.
func NewGeminiProvider(cfg *config.Config) (*GeminiProvider, error) {
	if cfg.Provider.APIKey == "" {
		return nil, fmt.Errorf("Gemini API key required (get free at aistudio.google.com)")
	}

	baseURL := cfg.Provider.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	model := cfg.Provider.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	return &GeminiProvider{
		apiKey:  cfg.Provider.APIKey,
		baseURL: baseURL,
		model:   model,
		config:  &cfg.Provider,
		client:  &http.Client{Timeout: cfg.Provider.Timeout},
	}, nil
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	if empty, err := ValidateReviewInput(req); err != nil {
		return nil, err
	} else if empty {
		return &ReviewResponse{}, nil
	}

	start := time.Now()
	geminiReq := BuildGeminiRequest(buildReviewPrompt(req), p.config.Temperature, p.config.MaxTokens, true)

	url := fmt.Sprintf(GeminiGenerateURL, p.baseURL, p.model, p.apiKey)
	var result GeminiResponse
	if err := DoJSONPost(ctx, p.client, url, geminiReq, "", &result); err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("gemini error %d: %s", result.Error.Code, result.Error.Message)
	}

	return ParseReviewContent(result.GetText(), result.UsageMetadata.TotalTokenCount, time.Since(start).Milliseconds()), nil
}

func (p *GeminiProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	geminiReq := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": fmt.Sprintf(CommitMessagePrompt, diff)}}},
		},
	}

	url := fmt.Sprintf(GeminiGenerateURL, p.baseURL, p.model, p.apiKey)
	var result GeminiResponse
	if err := DoJSONPost(ctx, p.client, url, geminiReq, "", &result); err != nil {
		return "", err
	}

	if text := result.GetText(); text != "" {
		return text, nil
	}
	return "", fmt.Errorf("no response from Gemini")
}

func (p *GeminiProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	geminiReq := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": fmt.Sprintf(DocumentationPrompt, docContext, diff)}}},
		},
	}

	url := fmt.Sprintf(GeminiGenerateURL, p.baseURL, p.model, p.apiKey)
	var result GeminiResponse
	if err := DoJSONPost(ctx, p.client, url, geminiReq, "", &result); err != nil {
		return "", err
	}

	if text := result.GetText(); text != "" {
		return text, nil
	}
	return "", fmt.Errorf("no response from Gemini")
}

func (p *GeminiProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/models/%s?key=%s", p.baseURL, p.model, p.apiKey)
	return DoHealthCheck(ctx, p.client, url, "", "gemini")
}

func (p *GeminiProvider) Close() error { return nil }
