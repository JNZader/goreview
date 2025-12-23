# Iteracion 04: Providers de IA

## Objetivos

- Interface Provider para abstraer LLMs
- Factory Pattern para crear providers
- Implementacion Ollama (local)
- Implementacion OpenAI
- **Implementacion Gemini** (primario para GitHub Bot - GRATIS)
- **Implementacion Groq** (fallback - GRATIS)
- **Implementacion Mistral** (fallback - GRATIS)
- **Sistema de Fallback** con múltiples providers
- Rate limiting integrado
- Tests completos

## Providers Disponibles

```
┌─────────────────────────────────────────────────────────────────┐
│                    ESTRATEGIA DE PROVIDERS                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  LOCAL (desarrollo):                                            │
│  └─ Ollama (Qwen 2.5 14B) → Gratis, rápido, sin internet       │
│                                                                 │
│  GITHUB BOT (producción):                                       │
│  ├─ 1. Gemini 2.5 Pro    → Primario (gratis, mejor calidad)    │
│  ├─ 2. Groq Llama 3.3    → Fallback (gratis, más rápido)       │
│  └─ 3. Mistral Codestral → Fallback (gratis, código)           │
│                                                                 │
│  PAGO (opcional):                                               │
│  └─ OpenAI GPT-4         → Si necesitas máxima calidad         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisitos

- Iteración 03 completada
- API Keys (todas gratis, sin tarjeta):
  - Gemini: [aistudio.google.com](https://aistudio.google.com)
  - Groq: [console.groq.com](https://console.groq.com)
  - Mistral: [console.mistral.ai](https://console.mistral.ai)

---

## Workflow GitFlow

Esta iteración tiene **9 commits atómicos**.

### Ramas de Esta Iteración

| Commit | Rama | Tipo |
|--------|------|------|
| 4.1 | `feature/04-01-provider-interface` | feat |
| 4.2 | `feature/04-02-provider-factory` | feat |
| 4.3 | `feature/04-03-ollama-provider` | feat |
| 4.4 | `feature/04-04-rate-limiter` | feat |
| 4.5 | `feature/04-05-openai-provider` | feat |
| 4.6 | `feature/04-06-gemini-provider` | feat |
| 4.7 | `feature/04-07-groq-provider` | feat |
| 4.8 | `feature/04-08-mistral-provider` | feat |
| 4.9 | `feature/04-09-fallback-provider` | feat |

### Flujo para Cada Commit

```bash
git checkout develop && git pull origin develop
git checkout -b feature/04-XX-slug
# ... implementar ...
git add . && git commit -m "mensaje"
git push -u origin feature/04-XX-slug
gh pr create --base develop --fill
gh pr merge --squash --delete-branch
git checkout develop && git pull
```

> **Guía completa:** [GITFLOW-SOLO-DEV.md](GITFLOW-SOLO-DEV.md)

---

## Commit 4.1: Crear interface Provider

**Mensaje de commit:**
```
feat(providers): add provider interface

- Define Provider interface for LLM operations
- Add ReviewRequest and ReviewResponse types
- Add Issue and Severity types
- Define Location for code references
```

### `goreview/internal/providers/types.go`

```go
package providers

import "context"

// Provider defines the interface for AI/LLM providers.
type Provider interface {
	// Name returns the provider name.
	Name() string

	// Review analyzes code and returns issues.
	Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error)

	// GenerateCommitMessage generates a commit message from diff.
	GenerateCommitMessage(ctx context.Context, diff string) (string, error)

	// GenerateDocumentation generates documentation from diff.
	GenerateDocumentation(ctx context.Context, diff, context string) (string, error)

	// HealthCheck verifies the provider is available.
	HealthCheck(ctx context.Context) error

	// Close releases any resources.
	Close() error
}

// ReviewRequest contains the input for a code review.
type ReviewRequest struct {
	Diff        string   `json:"diff"`
	Language    string   `json:"language"`
	FilePath    string   `json:"file_path"`
	FileContent string   `json:"file_content,omitempty"`
	Context     string   `json:"context,omitempty"`
	Rules       []string `json:"rules,omitempty"`
}

// ReviewResponse contains the review results.
type ReviewResponse struct {
	Issues         []Issue `json:"issues"`
	Summary        string  `json:"summary"`
	Score          int     `json:"score"` // 0-100
	TokensUsed     int     `json:"tokens_used"`
	ProcessingTime int64   `json:"processing_time_ms"`
}

// Issue represents a code review issue.
type Issue struct {
	ID         string    `json:"id"`
	Type       IssueType `json:"type"`
	Severity   Severity  `json:"severity"`
	Message    string    `json:"message"`
	Suggestion string    `json:"suggestion,omitempty"`
	Location   *Location `json:"location,omitempty"`
	RuleID     string    `json:"rule_id,omitempty"`
	FixedCode  string    `json:"fixed_code,omitempty"`
}

// Location represents a position in code.
type Location struct {
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	StartCol  int    `json:"start_col,omitempty"`
	EndCol    int    `json:"end_col,omitempty"`
}

// IssueType categorizes the type of issue.
type IssueType string

const (
	IssueTypeBug           IssueType = "bug"
	IssueTypeSecurity      IssueType = "security"
	IssueTypePerformance   IssueType = "performance"
	IssueTypeStyle         IssueType = "style"
	IssueTypeMaintenance   IssueType = "maintenance"
	IssueTypeBestPractice  IssueType = "best_practice"
)

// Severity indicates the importance of an issue.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)
```

---

## Commit 4.2: Crear Factory de providers

**Mensaje de commit:**
```
feat(providers): add provider factory

- Create factory function for providers
- Support ollama and openai providers
- Return error for unknown providers
```

### `goreview/internal/providers/factory.go`

```go
package providers

import (
	"fmt"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
)

// NewProvider creates a new Provider based on configuration.
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.Provider.Name {
	case "ollama":
		return NewOllamaProvider(cfg)
	case "openai":
		return NewOpenAIProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider.Name)
	}
}

// AvailableProviders returns a list of available provider names.
func AvailableProviders() []string {
	return []string{"ollama", "openai"}
}
```

---

## Commit 4.3: Implementar Ollama Provider

**Mensaje de commit:**
```
feat(providers): add ollama provider

- Implement Ollama API client
- Build review prompts dynamically
- Parse JSON responses from LLM
- Add retry logic for transient failures
- Implement rate limiting
```

### `goreview/internal/providers/ollama.go`

```go
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
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

	body, _ := json.Marshal(ollamaReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
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

	body, _ := json.Marshal(ollamaReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct{ Response string `json:"response"` }
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Response, nil
}

func (p *OllamaProvider) GenerateDocumentation(ctx context.Context, diff, context string) (string, error) {
	prompt := fmt.Sprintf(`Generate documentation for these changes.
Context: %s
Changes:
%s

Format as Markdown.`, context, diff)

	ollamaReq := map[string]interface{}{
		"model": p.model, "prompt": prompt, "stream": false,
	}

	body, _ := json.Marshal(ollamaReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/generate", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct{ Response string `json:"response"` }
	json.NewDecoder(resp.Body).Decode(&result)
	return result.Response, nil
}

func (p *OllamaProvider) HealthCheck(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/api/tags", nil)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
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
```

---

## Commit 4.4: Agregar Rate Limiter

**Mensaje de commit:**
```
feat(providers): add rate limiter

- Implement token bucket algorithm
- Support configurable RPS limit
- Handle context cancellation
```

### `goreview/internal/providers/ratelimit.go`

```go
package providers

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements a simple token bucket rate limiter.
type RateLimiter struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter with the given RPS.
func NewRateLimiter(rps int) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(rps),
		maxTokens:  float64(rps),
		refillRate: float64(rps),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available or context is cancelled.
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		r.refill()

		if r.tokens >= 1 {
			r.tokens--
			r.mu.Unlock()
			return nil
		}

		waitTime := time.Duration(float64(time.Second) / r.refillRate)
		r.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue loop
		}
	}
}

func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.tokens += elapsed * r.refillRate
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}
	r.lastRefill = now
}
```

---

## Commit 4.5: Implementar OpenAI Provider

**Mensaje de commit:**
```
feat(providers): add openai provider

- Implement OpenAI Chat Completions API
- Use gpt-4 by default
- Handle API authentication
```

### `goreview/internal/providers/openai.go`

```go
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
)

type OpenAIProvider struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
	config  *config.ProviderConfig
}

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

	body, _ := json.Marshal(openaiReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
		Usage struct{ TotalTokens int `json:"total_tokens"` } `json:"usage"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	var reviewResp ReviewResponse
	if len(result.Choices) > 0 {
		json.Unmarshal([]byte(result.Choices[0].Message.Content), &reviewResp)
	}
	reviewResp.TokensUsed = result.Usage.TotalTokens
	reviewResp.ProcessingTime = time.Since(start).Milliseconds()

	return &reviewResp, nil
}

func (p *OpenAIProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (p *OpenAIProvider) GenerateDocumentation(ctx context.Context, diff, context string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (p *OpenAIProvider) Close() error { return nil }
```

---

## Commit 4.6: Implementar Gemini Provider (PRIMARIO - GRATIS)

**Rama:** `feature/04-06-gemini-provider`

**Mensaje de commit:**
```
feat(providers): add gemini provider

- Implement Google Gemini API client
- Use gemini-2.5-pro model (best free tier)
- Handle Google AI Studio authentication
- Add response parsing for JSON output
```

> **Obtener API Key:** [aistudio.google.com](https://aistudio.google.com) → Get API Key → Create

### `goreview/internal/providers/gemini.go`

```go
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
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
		model = "gemini-2.5-pro-preview-06-05"
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
			"temperature":     p.config.Temperature,
			"maxOutputTokens": p.config.MaxTokens,
			"responseMimeType": "application/json",
		},
	}

	body, _ := json.Marshal(geminiReq)
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini request failed: %w", err)
	}
	defer resp.Body.Close()

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

	body, _ := json.Marshal(geminiReq)
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct{ Text string `json:"text"` } `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

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

	body, _ := json.Marshal(geminiReq)
	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct{ Text string `json:"text"` } `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Candidates) > 0 && len(result.Candidates[0].Content.Parts) > 0 {
		return result.Candidates[0].Content.Parts[0].Text, nil
	}
	return "", fmt.Errorf("no response from Gemini")
}

func (p *GeminiProvider) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/models/%s?key=%s", p.baseURL, p.model, p.apiKey)
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gemini health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (p *GeminiProvider) Close() error { return nil }
```

**Verificación:**
```bash
go build ./...
```

---

## Commit 4.7: Implementar Groq Provider (FALLBACK - GRATIS)

**Rama:** `feature/04-07-groq-provider`

**Mensaje de commit:**
```
feat(providers): add groq provider

- Implement Groq API client (OpenAI-compatible)
- Use llama-3.3-70b-versatile model
- Ultra-fast inference (~300 tokens/sec)
- 14,400 requests/day free tier
```

> **Obtener API Key:** [console.groq.com](https://console.groq.com) → API Keys → Create

### `goreview/internal/providers/groq.go`

```go
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
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
	start := time.Now()

	messages := []map[string]string{
		{"role": "system", "content": "You are an expert code reviewer. Return valid JSON only."},
		{"role": "user", "content": buildReviewPrompt(req)},
	}

	groqReq := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"temperature": p.config.Temperature,
		"max_tokens":  p.config.MaxTokens,
		"response_format": map[string]string{"type": "json_object"},
	}

	body, _ := json.Marshal(groqReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("groq request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
		Usage struct{ TotalTokens int `json:"total_tokens"` } `json:"usage"`
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

	body, _ := json.Marshal(groqReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

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

	body, _ := json.Marshal(groqReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from Groq")
}

func (p *GroqProvider) HealthCheck(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("groq health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (p *GroqProvider) Close() error { return nil }
```

**Verificación:**
```bash
go build ./...
```

---

## Commit 4.8: Implementar Mistral Provider (FALLBACK - GRATIS)

**Rama:** `feature/04-08-mistral-provider`

**Mensaje de commit:**
```
feat(providers): add mistral provider

- Implement Mistral API client
- Use codestral-latest model (code-specialized)
- 1 billion tokens/month free tier
- Optimized for code analysis
```

> **Obtener API Key:** [console.mistral.ai](https://console.mistral.ai) → API Keys → Create

### `goreview/internal/providers/mistral.go`

```go
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
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
	start := time.Now()

	messages := []map[string]string{
		{"role": "system", "content": "You are an expert code reviewer specialized in finding bugs and security issues. Return valid JSON only."},
		{"role": "user", "content": buildReviewPrompt(req)},
	}

	mistralReq := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"temperature": p.config.Temperature,
		"max_tokens":  p.config.MaxTokens,
		"response_format": map[string]string{"type": "json_object"},
	}

	body, _ := json.Marshal(mistralReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("mistral request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
		Usage struct{ TotalTokens int `json:"total_tokens"` } `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("mistral error: %s", result.Error.Message)
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

func (p *MistralProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	prompt := fmt.Sprintf(`Generate a conventional commit message for this diff.
Format: <type>(<scope>): <description>
Types: feat, fix, docs, style, refactor, perf, test, chore

Diff:
%s

Return ONLY the commit message.`, diff)

	messages := []map[string]string{
		{"role": "user", "content": prompt},
	}

	mistralReq := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
	}

	body, _ := json.Marshal(mistralReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from Mistral")
}

func (p *MistralProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	prompt := fmt.Sprintf(`Generate documentation for these changes.
Context: %s
Changes:
%s

Format as Markdown.`, docContext, diff)

	messages := []map[string]string{
		{"role": "user", "content": prompt},
	}

	mistralReq := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
	}

	body, _ := json.Marshal(mistralReq)
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct{ Content string } `json:"message"`
		} `json:"choices"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from Mistral")
}

func (p *MistralProvider) HealthCheck(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	resp, err := p.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mistral health check failed: %d", resp.StatusCode)
	}
	return nil
}

func (p *MistralProvider) Close() error { return nil }
```

**Verificación:**
```bash
go build ./...
```

---

## Commit 4.9: Implementar Fallback Provider

**Rama:** `feature/04-09-fallback-provider`

**Mensaje de commit:**
```
feat(providers): add fallback provider chain

- Implement FallbackProvider with ordered provider list
- Automatic failover on errors
- Health check for all providers
- Configurable retry behavior
```

### `goreview/internal/providers/fallback.go`

```go
package providers

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// FallbackProvider wraps multiple providers with automatic failover.
type FallbackProvider struct {
	providers []Provider
	primary   int // Index of current primary provider
	mu        sync.RWMutex
}

// NewFallbackProvider creates a provider chain with automatic failover.
// Providers are tried in order: first one that works becomes primary.
func NewFallbackProvider(providers ...Provider) (*FallbackProvider, error) {
	if len(providers) == 0 {
		return nil, fmt.Errorf("at least one provider required")
	}

	return &FallbackProvider{
		providers: providers,
		primary:   0,
	}, nil
}

func (f *FallbackProvider) Name() string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return fmt.Sprintf("fallback(%s)", f.providers[f.primary].Name())
}

func (f *FallbackProvider) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	f.mu.RLock()
	startIdx := f.primary
	f.mu.RUnlock()

	var lastErr error
	for i := 0; i < len(f.providers); i++ {
		idx := (startIdx + i) % len(f.providers)
		provider := f.providers[idx]

		resp, err := provider.Review(ctx, req)
		if err == nil {
			// Update primary if we fell back to a different provider
			if idx != startIdx {
				f.mu.Lock()
				f.primary = idx
				f.mu.Unlock()
				log.Printf("[fallback] Switched to provider: %s", provider.Name())
			}
			return resp, nil
		}

		lastErr = err
		log.Printf("[fallback] Provider %s failed: %v, trying next...", provider.Name(), err)
	}

	return nil, fmt.Errorf("all providers failed, last error: %w", lastErr)
}

func (f *FallbackProvider) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	f.mu.RLock()
	startIdx := f.primary
	f.mu.RUnlock()

	var lastErr error
	for i := 0; i < len(f.providers); i++ {
		idx := (startIdx + i) % len(f.providers)
		provider := f.providers[idx]

		msg, err := provider.GenerateCommitMessage(ctx, diff)
		if err == nil {
			return msg, nil
		}
		lastErr = err
		log.Printf("[fallback] Provider %s failed for commit msg: %v", provider.Name(), err)
	}

	return "", fmt.Errorf("all providers failed: %w", lastErr)
}

func (f *FallbackProvider) GenerateDocumentation(ctx context.Context, diff, docContext string) (string, error) {
	f.mu.RLock()
	startIdx := f.primary
	f.mu.RUnlock()

	var lastErr error
	for i := 0; i < len(f.providers); i++ {
		idx := (startIdx + i) % len(f.providers)
		provider := f.providers[idx]

		doc, err := provider.GenerateDocumentation(ctx, diff, docContext)
		if err == nil {
			return doc, nil
		}
		lastErr = err
	}

	return "", fmt.Errorf("all providers failed: %w", lastErr)
}

func (f *FallbackProvider) HealthCheck(ctx context.Context) error {
	// Check all providers and report which ones are healthy
	var healthy []string
	var unhealthy []string

	for _, p := range f.providers {
		if err := p.HealthCheck(ctx); err != nil {
			unhealthy = append(unhealthy, p.Name())
		} else {
			healthy = append(healthy, p.Name())
		}
	}

	if len(healthy) == 0 {
		return fmt.Errorf("no healthy providers available, unhealthy: %v", unhealthy)
	}

	log.Printf("[fallback] Healthy providers: %v, Unhealthy: %v", healthy, unhealthy)
	return nil
}

func (f *FallbackProvider) Close() error {
	var errs []error
	for _, p := range f.providers {
		if err := p.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing providers: %v", errs)
	}
	return nil
}

// GetProviderStatus returns the status of all providers in the chain.
func (f *FallbackProvider) GetProviderStatus(ctx context.Context) map[string]bool {
	status := make(map[string]bool)
	for _, p := range f.providers {
		status[p.Name()] = p.HealthCheck(ctx) == nil
	}
	return status
}
```

### Actualizar Factory (`goreview/internal/providers/factory.go`)

```go
package providers

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/TU-USUARIO/ai-toolkit/goreview/internal/config"
)

// NewProvider creates a new Provider based on configuration.
func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.Provider.Name {
	case "ollama":
		return NewOllamaProvider(cfg)
	case "openai":
		return NewOpenAIProvider(cfg)
	case "gemini":
		return NewGeminiProvider(cfg)
	case "groq":
		return NewGroqProvider(cfg)
	case "mistral":
		return NewMistralProvider(cfg)
	case "fallback":
		return NewFallbackFromEnv()
	case "auto", "":
		return NewAutoProvider(cfg)
	default:
		return nil, fmt.Errorf("unknown provider: %s", cfg.Provider.Name)
	}
}

// NewAutoProvider automatically selects the best provider based on environment.
// - If Ollama is running locally → use Ollama (fast, free, offline)
// - If cloud API keys exist → use Fallback chain
// - Otherwise → error
func NewAutoProvider(cfg *config.Config) (Provider, error) {
	// Check if we're in CI/GitHub Actions environment
	if os.Getenv("GITHUB_ACTIONS") == "true" || os.Getenv("CI") == "true" {
		log.Println("[auto] CI environment detected, using cloud providers")
		return NewFallbackFromEnv()
	}

	// Check if Ollama is running locally
	if isOllamaRunning() {
		log.Println("[auto] Ollama detected locally, using Ollama")
		if cfg.Provider.BaseURL == "" {
			cfg.Provider.BaseURL = "http://localhost:11434"
		}
		if cfg.Provider.Model == "" {
			cfg.Provider.Model = "qwen2.5-coder:14b"
		}
		return NewOllamaProvider(cfg)
	}

	// Check if any cloud API keys are set
	if hasCloudAPIKeys() {
		log.Println("[auto] Cloud API keys found, using fallback chain")
		return NewFallbackFromEnv()
	}

	return nil, fmt.Errorf("no provider available. Either:\n" +
		"  1. Start Ollama locally: ollama serve\n" +
		"  2. Set API keys: GEMINI_API_KEY, GROQ_API_KEY, or MISTRAL_API_KEY")
}

// isOllamaRunning checks if Ollama is running on localhost.
func isOllamaRunning() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// hasCloudAPIKeys checks if any cloud provider API keys are set.
func hasCloudAPIKeys() bool {
	keys := []string{"GEMINI_API_KEY", "GROQ_API_KEY", "MISTRAL_API_KEY", "OPENAI_API_KEY"}
	for _, key := range keys {
		if os.Getenv(key) != "" {
			return true
		}
	}
	return false
}

// NewFallbackFromEnv creates a fallback provider chain from environment variables.
// Priority: Gemini (quality) -> Groq (speed) -> Mistral (code) -> OpenAI (paid)
func NewFallbackFromEnv() (Provider, error) {
	var providers []Provider

	// Try Gemini first (best quality, free)
	if key := os.Getenv("GEMINI_API_KEY"); key != "" {
		cfg := &config.Config{
			Provider: config.ProviderConfig{
				APIKey:      key,
				Temperature: 0.1,
				MaxTokens:   4096,
				Timeout:     60 * time.Second,
			},
		}
		if p, err := NewGeminiProvider(cfg); err == nil {
			providers = append(providers, p)
			log.Println("[fallback] Added Gemini provider")
		}
	}

	// Try Groq second (fastest, free)
	if key := os.Getenv("GROQ_API_KEY"); key != "" {
		cfg := &config.Config{
			Provider: config.ProviderConfig{
				APIKey:      key,
				Temperature: 0.1,
				MaxTokens:   4096,
				Timeout:     30 * time.Second,
			},
		}
		if p, err := NewGroqProvider(cfg); err == nil {
			providers = append(providers, p)
			log.Println("[fallback] Added Groq provider")
		}
	}

	// Try Mistral third (code-specialized, free)
	if key := os.Getenv("MISTRAL_API_KEY"); key != "" {
		cfg := &config.Config{
			Provider: config.ProviderConfig{
				APIKey:      key,
				Temperature: 0.1,
				MaxTokens:   4096,
				Timeout:     60 * time.Second,
			},
		}
		if p, err := NewMistralProvider(cfg); err == nil {
			providers = append(providers, p)
			log.Println("[fallback] Added Mistral provider")
		}
	}

	// Try OpenAI last (paid, but reliable)
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		cfg := &config.Config{
			Provider: config.ProviderConfig{
				APIKey:      key,
				Model:       "gpt-4",
				Temperature: 0.1,
				MaxTokens:   4096,
				Timeout:     60 * time.Second,
			},
		}
		if p, err := NewOpenAIProvider(cfg); err == nil {
			providers = append(providers, p)
			log.Println("[fallback] Added OpenAI provider")
		}
	}

	if len(providers) == 0 {
		return nil, fmt.Errorf("no API keys found. Set one of: GEMINI_API_KEY, GROQ_API_KEY, MISTRAL_API_KEY, OPENAI_API_KEY")
	}

	return NewFallbackProvider(providers...)
}

// AvailableProviders returns a list of available provider names.
func AvailableProviders() []string {
	return []string{"ollama", "openai", "gemini", "groq", "mistral", "fallback", "auto"}
}
```

### Comportamiento Automático

```
┌─────────────────────────────────────────────────────────────────┐
│                     DETECCIÓN AUTOMÁTICA                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  provider: "auto" (o vacío)                                     │
│  │                                                              │
│  ├─► ¿GITHUB_ACTIONS=true?                                      │
│  │     └─► SÍ → Usar Fallback (Gemini→Groq→Mistral)            │
│  │                                                              │
│  ├─► ¿Ollama corriendo en localhost:11434?                      │
│  │     └─► SÍ → Usar Ollama (qwen2.5-coder:14b)                │
│  │                                                              │
│  └─► ¿Hay API keys configuradas?                                │
│        └─► SÍ → Usar Fallback                                   │
│        └─► NO → Error con instrucciones                         │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Configuración por Entorno

**Local (tu máquina):**
```bash
# Solo necesitas Ollama corriendo
ollama serve &
ollama pull qwen2.5-coder:14b

# El provider "auto" detectará Ollama automáticamente
./goreview review --staged
```

**GitHub Bot (Actions/Docker):**
```yaml
# .github/workflows/review.yml o docker-compose.yml
env:
  GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
  GROQ_API_KEY: ${{ secrets.GROQ_API_KEY }}
  MISTRAL_API_KEY: ${{ secrets.MISTRAL_API_KEY }}
```

**Forzar provider específico:**
```bash
# Forzar Ollama aunque haya API keys
GOREVIEW_PROVIDER=ollama ./goreview review

# Forzar Gemini
GOREVIEW_PROVIDER=gemini ./goreview review
```

**Verificación:**
```bash
go build ./...
go test ./internal/providers/...
```

---

## Resumen de la Iteracion 04

### Commits (9 total):
1. `feat(providers): add provider interface`
2. `feat(providers): add provider factory`
3. `feat(providers): add ollama provider`
4. `feat(providers): add rate limiter`
5. `feat(providers): add openai provider`
6. `feat(providers): add gemini provider` (PRIMARIO - GRATIS)
7. `feat(providers): add groq provider` (FALLBACK - GRATIS)
8. `feat(providers): add mistral provider` (FALLBACK - GRATIS)
9. `feat(providers): add fallback provider chain`

### Archivos:
```
goreview/internal/providers/
├── types.go      # Interface y tipos comunes
├── factory.go    # Factory con soporte para todos los providers
├── ratelimit.go  # Rate limiter compartido
├── ollama.go     # Provider local (Qwen 2.5 14B)
├── openai.go     # Provider OpenAI (pago)
├── gemini.go     # Provider Gemini (gratis, primario)
├── groq.go       # Provider Groq (gratis, fallback)
├── mistral.go    # Provider Mistral (gratis, fallback)
└── fallback.go   # Fallback chain con auto-failover
```

### Variables de Entorno:
```bash
# Para usar el sistema de fallback automatico:
export GEMINI_API_KEY="tu-api-key"   # aistudio.google.com
export GROQ_API_KEY="tu-api-key"     # console.groq.com
export MISTRAL_API_KEY="tu-api-key"  # console.mistral.ai

# Opcional (pago):
export OPENAI_API_KEY="tu-api-key"
```

### Uso:
```go
// Usar provider especifico
cfg.Provider.Name = "gemini"
provider, _ := providers.NewProvider(cfg)

// Usar fallback automatico (recomendado para GitHub Bot)
cfg.Provider.Name = "fallback"
provider, _ := providers.NewProvider(cfg)
```

---

## Siguiente Iteracion

Continua con: **[05-MOTOR-REVIEW.md](05-MOTOR-REVIEW.md)**
