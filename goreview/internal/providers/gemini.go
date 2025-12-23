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
	start := time.Now()

	geminiReq := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]string{
					{"text": buildReviewPrompt(req)},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":      p.config.Temperature,
			"maxOutputTokens":  p.config.MaxTokens,
			"responseMimeType": "application/json",
		},
	}

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			TotalTokenCount int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
		Error *struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("gemini error %d: %s", result.Error.Code, result.Error.Message)
	}

	var reviewResp ReviewResponse
	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		text := result.Candidates[0].Content.Parts[0].Text
		if err := json.Unmarshal([]byte(text), &reviewResp); err != nil {
			reviewResp = ReviewResponse{Summary: text}
		}
	}
	reviewResp.TokensUsed = result.UsageMetadata.TotalTokenCount
	reviewResp.ProcessingTime = time.Since(start).Milliseconds()

	return &reviewResp, nil
}

func (p *GeminiProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	prompt := fmt.Sprintf(`Generate a conventional commit message for this diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, chore

Diff:
%s

Return ONLY the commit message, nothing else.`, diff)

	geminiReq := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("no response from Gemini")
}

func (p *GeminiProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	prompt := fmt.Sprintf(`Generate documentation for these changes.
Context: %s
Changes:
%s

Format as Markdown.`, docContext, diff)

	geminiReq := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}

	body, err := json.Marshal(geminiReq)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("no response from Gemini")
}

func (p *GeminiProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/models/%s?key=%s", p.baseURL, p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gemini health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (p *GeminiProvider) Close() error { return nil }
