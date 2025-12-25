package knowledge

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
)

// Fetcher handles fetching documents from knowledge sources.
type Fetcher struct {
	config   Config
	client   *http.Client
	cacheDir string
}

// NewFetcher creates a new knowledge fetcher.
func NewFetcher(cfg Config) (*Fetcher, error) {
	cacheDir := cfg.CacheDir
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		cacheDir = filepath.Join(home, ".goreview", "knowledge-cache")
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

// FetchContext retrieves knowledge context from configured sources.
func (f *Fetcher) FetchContext(ctx context.Context, query string) (*Context, error) {
	if !f.config.Enabled {
		return &Context{}, nil
	}

	var allDocs []Document

	for _, source := range f.config.Sources {
		if !source.Enabled {
			continue
		}

		docs, err := f.fetchFromSource(ctx, source, query)
		if err != nil {
			// Log warning but continue with other sources
			fmt.Fprintf(os.Stderr, "Warning: failed to fetch from %s: %v\n", source.Name, err)
			continue
		}

		allDocs = append(allDocs, docs...)
	}

	// Limit total documents
	maxDocs := f.config.MaxDocs
	if maxDocs <= 0 {
		maxDocs = 10
	}
	if len(allDocs) > maxDocs {
		allDocs = allDocs[:maxDocs]
	}

	return &Context{
		Documents:   allDocs,
		TotalTokens: estimateTokens(allDocs),
	}, nil
}

// fetchFromSource fetches documents from a single source.
func (f *Fetcher) fetchFromSource(ctx context.Context, source Source, query string) ([]Document, error) {
	switch source.Type {
	case SourceTypeNotion:
		return f.fetchFromNotion(ctx, source, query)
	case SourceTypeConfluence:
		return f.fetchFromConfluence(ctx, source, query)
	case SourceTypeObsidian:
		return f.fetchFromObsidian(source, query)
	case SourceTypeLocal:
		return f.fetchFromLocal(source, query)
	case SourceTypeGitHub:
		return f.fetchFromGitHub(ctx, source, query)
	default:
		return nil, fmt.Errorf("unknown source type: %s", source.Type)
	}
}

// fetchFromNotion fetches documents from Notion.
func (f *Fetcher) fetchFromNotion(ctx context.Context, source Source, query string) ([]Document, error) {
	if source.NotionToken == "" {
		return nil, fmt.Errorf("notion token required")
	}

	// Notion API endpoint
	var url string
	if source.NotionDatabaseID != "" {
		url = fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", source.NotionDatabaseID)
	} else if source.NotionPageID != "" {
		url = fmt.Sprintf("https://api.notion.com/v1/pages/%s", source.NotionPageID)
	} else {
		return nil, fmt.Errorf("notion database_id or page_id required")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+source.NotionToken)
	req.Header.Set("Notion-Version", "2022-06-28")
	req.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("notion API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseNotionResponse(body, source)
}

// fetchFromConfluence fetches documents from Confluence.
func (f *Fetcher) fetchFromConfluence(ctx context.Context, source Source, query string) ([]Document, error) {
	if source.ConfluenceURL == "" || source.ConfluenceToken == "" {
		return nil, fmt.Errorf("confluence URL and token required")
	}

	// Confluence REST API search
	searchURL := fmt.Sprintf("%s/rest/api/content/search?cql=space=%s+AND+text~%q",
		strings.TrimSuffix(source.ConfluenceURL, "/"),
		source.ConfluenceSpace,
		query,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(source.ConfluenceUser, source.ConfluenceToken)
	req.Header.Set("Accept", "application/json")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("confluence API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseConfluenceResponse(body, source)
}

// fetchFromObsidian fetches documents from an Obsidian vault.
func (f *Fetcher) fetchFromObsidian(source Source, query string) ([]Document, error) {
	if source.ObsidianVaultPath == "" {
		return nil, fmt.Errorf("obsidian vault path required")
	}

	var docs []Document
	queryLower := strings.ToLower(query)

	err := filepath.Walk(source.ObsidianVaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			return nil
		}

		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".markdown" {
			return nil
		}

		content, err := os.ReadFile(path) //nolint:gosec // Path from config
		if err != nil {
			return nil
		}

		contentStr := string(content)

		// Check if content matches query or tags
		if !strings.Contains(strings.ToLower(contentStr), queryLower) {
			// Check tags
			tags := extractObsidianTags(contentStr)
			hasTag := false
			for _, tag := range tags {
				for _, filterTag := range source.ObsidianTags {
					if strings.Contains(tag, filterTag) {
						hasTag = true
						break
					}
				}
			}
			if !hasTag && len(source.ObsidianTags) > 0 {
				return nil
			}
		}

		relPath, _ := filepath.Rel(source.ObsidianVaultPath, path)
		docs = append(docs, Document{
			ID:        hashString(path),
			Title:     strings.TrimSuffix(filepath.Base(path), ext),
			Content:   contentStr,
			Source:    SourceTypeObsidian,
			Tags:      extractObsidianTags(contentStr),
			FetchedAt: time.Now(),
			Metadata: map[string]string{
				"path": relPath,
			},
		})

		return nil
	})

	return docs, err
}

// fetchFromLocal fetches documents from local files.
func (f *Fetcher) fetchFromLocal(source Source, query string) ([]Document, error) {
	if source.LocalPath == "" {
		return nil, fmt.Errorf("local path required")
	}

	pattern := source.LocalPattern
	if pattern == "" {
		pattern = "**/*.md"
	}

	var docs []Document
	queryLower := strings.ToLower(query)

	err := filepath.Walk(source.LocalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		// Check pattern match
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if !matched && !strings.HasSuffix(pattern, "*") {
			return nil
		}

		content, err := os.ReadFile(path) //nolint:gosec // Path from config
		if err != nil {
			return nil
		}

		contentStr := string(content)
		if query != "" && !strings.Contains(strings.ToLower(contentStr), queryLower) {
			return nil
		}

		relPath, _ := filepath.Rel(source.LocalPath, path)
		ext := filepath.Ext(path)

		docs = append(docs, Document{
			ID:        hashString(path),
			Title:     strings.TrimSuffix(filepath.Base(path), ext),
			Content:   contentStr,
			Source:    SourceTypeLocal,
			FetchedAt: time.Now(),
			Metadata: map[string]string{
				"path": relPath,
			},
		})

		return nil
	})

	return docs, err
}

// fetchFromGitHub fetches documents from GitHub wiki or docs.
func (f *Fetcher) fetchFromGitHub(ctx context.Context, source Source, query string) ([]Document, error) {
	if source.GitHubOwner == "" || source.GitHubRepo == "" {
		return nil, fmt.Errorf("github owner and repo required")
	}

	// GitHub API to get repository contents
	path := source.GitHubPath
	if path == "" {
		path = "docs"
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s",
		source.GitHubOwner, source.GitHubRepo, path)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "GoReview/1.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API error: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseGitHubContents(ctx, f.client, body, source, query)
}

// Search searches across all configured knowledge sources.
func (f *Fetcher) Search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
	docs, err := f.FetchContext(ctx, query.Text)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	queryLower := strings.ToLower(query.Text)

	for _, doc := range docs.Documents {
		// Filter by source type
		if query.SourceType != "" && doc.Source != query.SourceType {
			continue
		}

		// Filter by tags
		if len(query.Tags) > 0 {
			hasTag := false
			for _, queryTag := range query.Tags {
				for _, docTag := range doc.Tags {
					if strings.Contains(strings.ToLower(docTag), strings.ToLower(queryTag)) {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Calculate score
		score := calculateRelevanceScore(doc, queryLower)

		results = append(results, SearchResult{
			Document: doc,
			Score:    score,
			Snippet:  extractSnippet(doc.Content, queryLower, 200),
		})
	}

	// Sort by score
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Apply limit
	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results, nil
}

// Helper functions

func parseNotionResponse(body []byte, source Source) ([]Document, error) {
	var result struct {
		Results []struct {
			ID         string `json:"id"`
			Properties struct {
				Title struct {
					Title []struct {
						PlainText string `json:"plain_text"`
					} `json:"title"`
				} `json:"title"`
			} `json:"properties"`
			URL string `json:"url"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var docs []Document
	for _, r := range result.Results {
		title := ""
		if len(r.Properties.Title.Title) > 0 {
			title = r.Properties.Title.Title[0].PlainText
		}

		docs = append(docs, Document{
			ID:        r.ID,
			Title:     title,
			URL:       r.URL,
			Source:    SourceTypeNotion,
			FetchedAt: time.Now(),
		})
	}

	return docs, nil
}

func parseConfluenceResponse(body []byte, source Source) ([]Document, error) {
	var result struct {
		Results []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Links struct {
				WebUI string `json:"webui"`
			} `json:"_links"`
			Body struct {
				Storage struct {
					Value string `json:"value"`
				} `json:"storage"`
			} `json:"body"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	var docs []Document
	for _, r := range result.Results {
		docs = append(docs, Document{
			ID:        r.ID,
			Title:     r.Title,
			Content:   stripHTML(r.Body.Storage.Value),
			URL:       source.ConfluenceURL + r.Links.WebUI,
			Source:    SourceTypeConfluence,
			FetchedAt: time.Now(),
		})
	}

	return docs, nil
}

func parseGitHubContents(ctx context.Context, client *http.Client, body []byte, source Source, query string) ([]Document, error) {
	var contents []struct {
		Name        string `json:"name"`
		Path        string `json:"path"`
		Type        string `json:"type"`
		DownloadURL string `json:"download_url"`
		HTMLURL     string `json:"html_url"`
	}

	if err := json.Unmarshal(body, &contents); err != nil {
		return nil, err
	}

	var docs []Document
	queryLower := strings.ToLower(query)

	for _, c := range contents {
		if c.Type != "file" {
			continue
		}

		ext := filepath.Ext(c.Name)
		if ext != ".md" && ext != ".rst" && ext != ".txt" {
			continue
		}

		// Fetch file content
		req, err := http.NewRequestWithContext(ctx, "GET", c.DownloadURL, nil)
		if err != nil {
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			continue
		}

		content, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		contentStr := string(content)
		if query != "" && !strings.Contains(strings.ToLower(contentStr), queryLower) {
			continue
		}

		docs = append(docs, Document{
			ID:        hashString(c.Path),
			Title:     strings.TrimSuffix(c.Name, ext),
			Content:   contentStr,
			URL:       c.HTMLURL,
			Source:    SourceTypeGitHub,
			FetchedAt: time.Now(),
			Metadata: map[string]string{
				"path": c.Path,
			},
		})
	}

	return docs, nil
}

func extractObsidianTags(content string) []string {
	var tags []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// Look for tags in format #tag
		words := strings.Fields(line)
		for _, word := range words {
			if strings.HasPrefix(word, "#") && len(word) > 1 {
				tag := strings.TrimPrefix(word, "#")
				tag = strings.Trim(tag, ".,!?;:")
				if tag != "" {
					tags = append(tags, tag)
				}
			}
		}
	}

	return tags
}

func stripHTML(html string) string {
	// Simple HTML stripping
	result := strings.ReplaceAll(html, "<br>", "\n")
	result = strings.ReplaceAll(result, "<br/>", "\n")
	result = strings.ReplaceAll(result, "</p>", "\n")
	result = strings.ReplaceAll(result, "</div>", "\n")

	// Remove remaining tags
	var sb strings.Builder
	inTag := false
	for _, r := range result {
		if r == '<' {
			inTag = true
		} else if r == '>' {
			inTag = false
		} else if !inTag {
			sb.WriteRune(r)
		}
	}

	return strings.TrimSpace(sb.String())
}

func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:8])
}

func estimateTokens(docs []Document) int {
	total := 0
	for _, doc := range docs {
		// Rough estimate: 4 chars per token
		total += len(doc.Content) / 4
	}
	return total
}

func calculateRelevanceScore(doc Document, query string) float64 {
	score := 0.0

	// Title match
	if strings.Contains(strings.ToLower(doc.Title), query) {
		score += 0.5
	}

	// Content match frequency
	contentLower := strings.ToLower(doc.Content)
	count := strings.Count(contentLower, query)
	if count > 0 {
		score += 0.3 + float64(count)*0.02
	}

	// Freshness bonus
	age := time.Since(doc.FetchedAt)
	if age < 24*time.Hour {
		score += 0.2
	} else if age < 7*24*time.Hour {
		score += 0.1
	}

	return score
}

func extractSnippet(content, query string, maxLen int) string {
	contentLower := strings.ToLower(content)
	idx := strings.Index(contentLower, query)

	if idx == -1 {
		if len(content) > maxLen {
			return content[:maxLen] + "..."
		}
		return content
	}

	start := idx - maxLen/2
	if start < 0 {
		start = 0
	}

	end := start + maxLen
	if end > len(content) {
		end = len(content)
	}

	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}

	return snippet
}
