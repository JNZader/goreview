package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Common error format strings (SonarQube S1192)
const (
	ErrMarshalRequest = "marshal request: %w"
	ErrCreateRequest  = "create request: %w"
	ErrDecodeResponse = "decode response: %w"
)

// Common HTTP constants
const (
	ContentTypeJSON     = "application/json"
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
	ChatCompletionsPath = "/chat/completions"
	APIGeneratePath     = "/api/generate"
	AuthBearerPrefix    = "Bearer "
	GeminiGenerateURL   = "%s/models/%s:generateContent?key=%s"
)

// Shared prompt templates (SonarQube duplications)
const (
	CommitMessagePrompt = `Generate a conventional commit message for this diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, chore

Diff:
%s

Return ONLY the commit message.`

	DocumentationPrompt = `Generate documentation for these changes.
Context: %s
Changes:
%s

Format as Markdown.`
)

// ChatCompletionResponse represents common response structure for OpenAI-compatible APIs
type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"message"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// GetContent returns the content from the first choice or empty string
func (r *ChatCompletionResponse) GetContent() string {
	if len(r.Choices) > 0 {
		return r.Choices[0].Message.Content
	}
	return ""
}

// DoJSONPost performs a JSON POST request and decodes the response
func DoJSONPost(ctx context.Context, client *http.Client, url string, reqBody interface{}, apiKey string, result interface{}) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf(ErrMarshalRequest, err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf(ErrCreateRequest, err)
	}
	httpReq.Header.Set(HeaderContentType, ContentTypeJSON)
	if apiKey != "" {
		httpReq.Header.Set("Authorization", AuthBearerPrefix+apiKey)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf(ErrDecodeResponse, err)
	}
	return nil
}

// OllamaResponse represents Ollama API response structure
type OllamaResponse struct {
	Response string `json:"response"`
}

// GeminiResponse represents Gemini API response structure
type GeminiResponse struct {
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

// GetText returns the text from the first candidate or empty string
func (r *GeminiResponse) GetText() string {
	if len(r.Candidates) > 0 && len(r.Candidates[0].Content.Parts) > 0 {
		return r.Candidates[0].Content.Parts[0].Text
	}
	return ""
}

// ReviewSystemPrompt is the standard system prompt for code review
const ReviewSystemPrompt = "You are an expert code reviewer. Return valid JSON only."

// ValidateReviewInput validates the review request and returns true if empty (should return empty response)
func ValidateReviewInput(req *ReviewRequest) (empty bool, err error) {
	if err := ValidateReviewRequest(req); err != nil {
		return false, fmt.Errorf("invalid request: %w", err)
	}
	return len(req.Diff) == 0, nil
}

// ParseReviewContent parses JSON content into ReviewResponse with fallback to summary
func ParseReviewContent(content string, tokensUsed int, processingTime int64) *ReviewResponse {
	var reviewResp ReviewResponse
	if err := json.Unmarshal([]byte(content), &reviewResp); err != nil {
		reviewResp = ReviewResponse{Summary: content}
	}
	reviewResp.TokensUsed = tokensUsed
	reviewResp.ProcessingTime = processingTime
	return &reviewResp
}

// BuildChatRequest builds a standard chat completion request for OpenAI-compatible APIs
func BuildChatRequest(model string, systemPrompt string, userContent string, temp float64, maxTokens int, jsonMode bool) map[string]interface{} {
	req := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userContent},
		},
		"temperature": temp,
		"max_tokens":  maxTokens,
	}
	if jsonMode {
		req["response_format"] = map[string]string{"type": "json_object"}
	}
	return req
}

// BuildGeminiRequest builds a Gemini API request
func BuildGeminiRequest(text string, temp float64, maxTokens int, jsonMode bool) map[string]interface{} {
	req := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": text}}},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     temp,
			"maxOutputTokens": maxTokens,
		},
	}
	if jsonMode {
		req["generationConfig"].(map[string]interface{})["responseMimeType"] = "application/json"
	}
	return req
}

// BuildOllamaRequest builds an Ollama API request
func BuildOllamaRequest(model string, prompt string, temp float64, maxTokens int, jsonMode bool) map[string]interface{} {
	req := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": temp,
			"num_predict": maxTokens,
		},
	}
	if jsonMode {
		req["format"] = "json"
	}
	return req
}

// DoHealthCheck performs a health check GET request
func DoHealthCheck(ctx context.Context, client *http.Client, url string, apiKey string, providerName string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf(ErrCreateRequest, err)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", AuthBearerPrefix+apiKey)
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s health check failed: %d", providerName, resp.StatusCode)
	}
	return nil
}
