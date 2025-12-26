package rag

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// Fetcher handles fetching and caching external documentation.
type Fetcher struct {
	config   RAGConfig
	client   *http.Client
	cacheDir string
}

// NewFetcher creates a new documentation fetcher.
func NewFetcher(cfg RAGConfig) (*Fetcher, error) {
	cacheDir := cfg.CacheDir
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".goreview", "rag-cache")
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	return &Fetcher{
		config:   cfg,
		cacheDir: cacheDir,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// FetchContext retrieves RAG context for a review based on detected frameworks.
func (f *Fetcher) FetchContext(ctx context.Context, language string, frameworks []DetectedFramework) (*Context, error) {
	if !f.config.Enabled {
		return &Context{}, nil
	}

	ragCtx := &Context{
		Sources: make([]SourceContext, 0),
	}

	configuredSources := f.fetchConfiguredSources(ctx, language)
	ragCtx.Sources = append(ragCtx.Sources, configuredSources...)

	if f.config.AutoDetect {
		frameworkSources := f.fetchFrameworkSources(ctx, frameworks)
		ragCtx.Sources = append(ragCtx.Sources, frameworkSources...)
	}

	return ragCtx, nil
}

func (f *Fetcher) fetchConfiguredSources(ctx context.Context, language string) []SourceContext {
	sources := make([]SourceContext, 0, len(f.config.Sources))

	for _, source := range f.config.Sources {
		if !isSourceRelevant(source, language) {
			continue
		}

		doc, err := f.fetchWithCache(ctx, source)
		if err != nil {
			continue
		}

		sources = append(sources, SourceContext{
			Name:     source.Name,
			Type:     source.Type,
			Content:  doc.Summary,
			URL:      source.URL,
			Relevant: true,
		})
	}

	return sources
}

func isSourceRelevant(source Source, language string) bool {
	if !source.Enabled {
		return false
	}
	if source.Language != "" && source.Language != language {
		return false
	}
	return true
}

func (f *Fetcher) fetchFrameworkSources(ctx context.Context, frameworks []DetectedFramework) []SourceContext {
	sources := make([]SourceContext, 0, len(frameworks))

	for _, fw := range frameworks {
		if fw.DocsURL == "" {
			continue
		}

		source := f.createFrameworkSource(fw)
		doc, err := f.fetchWithCache(ctx, source)
		if err != nil {
			continue
		}

		sources = append(sources, SourceContext{
			Name:     fw.Name,
			Type:     SourceTypeFramework,
			Content:  doc.Summary,
			URL:      fw.DocsURL,
			Relevant: true,
		})
	}

	return sources
}

func (f *Fetcher) createFrameworkSource(fw DetectedFramework) Source {
	return Source{
		URL:      fw.DocsURL,
		Type:     SourceTypeFramework,
		Name:     fw.Name,
		Language: fw.Language,
		Enabled:  true,
	}
}

// fetchWithCache fetches a document with caching support.
func (f *Fetcher) fetchWithCache(ctx context.Context, source Source) (*CachedDocument, error) {
	cacheKey := hashURL(source.URL)
	cachePath := filepath.Join(f.cacheDir, cacheKey+".json")

	// Check cache
	if cached, err := f.loadFromCache(cachePath); err == nil {
		if time.Now().Before(cached.ExpiresAt) {
			return cached, nil
		}
	}

	// Fetch fresh content
	content, err := f.fetchURL(ctx, source.URL)
	if err != nil {
		return nil, err
	}

	// Parse TTL
	ttl := parseTTL(source.CacheTTL, f.config.DefaultCacheTTL)

	doc := &CachedDocument{
		URL:         source.URL,
		Content:     content,
		Summary:     extractSummary(content, 2000),
		FetchedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
		ContentHash: hashContent(content),
	}

	// Save to cache
	_ = f.saveToCache(cachePath, doc)

	return doc, nil
}

// fetchURL fetches content from a URL.
func (f *Fetcher) fetchURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "GoReview/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return "", err
	}

	// If HTML, extract text content
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/html") {
		return extractTextFromHTML(string(body)), nil
	}

	return string(body), nil
}

// loadFromCache loads a cached document.
func (f *Fetcher) loadFromCache(path string) (*CachedDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var doc CachedDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}

// saveToCache saves a document to cache.
func (f *Fetcher) saveToCache(path string, doc *CachedDocument) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// ClearCache removes all cached documents.
func (f *Fetcher) ClearCache() error {
	return os.RemoveAll(f.cacheDir)
}

// GetCacheStats returns cache statistics.
func (f *Fetcher) GetCacheStats() (int, int64, error) {
	var count int
	var size int64

	err := filepath.Walk(f.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			count++
			size += info.Size()
		}
		return nil
	})

	return count, size, err
}

// Helper functions

func hashURL(url string) string {
	h := sha256.Sum256([]byte(url))
	return hex.EncodeToString(h[:16])
}

func hashContent(content string) string {
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func parseTTL(ttl, defaultTTL string) time.Duration {
	if ttl == "" {
		ttl = defaultTTL
	}
	if ttl == "" {
		ttl = "24h"
	}

	d, err := time.ParseDuration(ttl)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func extractSummary(content string, maxLen int) string {
	// Clean and truncate content
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "\n\n\n", "\n\n")

	if len(content) > maxLen {
		// Find a good break point
		if idx := strings.LastIndex(content[:maxLen], ". "); idx > maxLen/2 {
			return content[:idx+1]
		}
		return content[:maxLen] + "..."
	}
	return content
}

func extractTextFromHTML(htmlContent string) string {
	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return htmlContent
	}

	var sb strings.Builder
	var extractText func(*html.Node)

	extractText = func(n *html.Node) {
		if n.Type == html.TextNode {
			text := strings.TrimSpace(n.Data)
			if text != "" {
				sb.WriteString(text)
				sb.WriteString(" ")
			}
		}

		// Skip script and style elements
		if n.Type == html.ElementNode {
			switch n.Data {
			case "script", "style", "nav", "footer", "header":
				return
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}

		// Add newlines after block elements
		if n.Type == html.ElementNode {
			switch n.Data {
			case "p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li", "br":
				sb.WriteString("\n")
			}
		}
	}

	extractText(doc)
	return strings.TrimSpace(sb.String())
}
