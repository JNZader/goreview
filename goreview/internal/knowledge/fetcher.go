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
	"sort"
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
func (f *Fetcher) fetchFromNotion(ctx context.Context, source Source, _ string) ([]Document, error) {
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
		if err != nil || info.IsDir() {
			return nil
		}

		doc := processObsidianFile(path, source, queryLower)
		if doc != nil {
			docs = append(docs, *doc)
		}
		return nil
	})

	return docs, err
}

func processObsidianFile(path string, source Source, queryLower string) *Document {
	ext := filepath.Ext(path)
	if !isMarkdownFile(ext) {
		return nil
	}

	content, err := os.ReadFile(path) //nolint:gosec // Path from config
	if err != nil {
		return nil
	}

	contentStr := string(content)
	if !matchesObsidianQuery(contentStr, queryLower, source.ObsidianTags) {
		return nil
	}

	relPath, _ := filepath.Rel(source.ObsidianVaultPath, path)
	return &Document{
		ID:        hashString(path),
		Title:     strings.TrimSuffix(filepath.Base(path), ext),
		Content:   contentStr,
		Source:    SourceTypeObsidian,
		Tags:      extractObsidianTags(contentStr),
		FetchedAt: time.Now(),
		Metadata: map[string]string{
			"path": relPath,
		},
	}
}

func isMarkdownFile(ext string) bool {
	return ext == ".md" || ext == ".markdown"
}

func matchesObsidianQuery(content, queryLower string, filterTags []string) bool {
	if strings.Contains(strings.ToLower(content), queryLower) {
		return true
	}
	if len(filterTags) == 0 {
		return true
	}
	return hasObsidianTagMatch(content, filterTags)
}

func hasObsidianTagMatch(content string, filterTags []string) bool {
	tags := extractObsidianTags(content)
	for _, tag := range tags {
		for _, filterTag := range filterTags {
			if strings.Contains(tag, filterTag) {
				return true
			}
		}
	}
	return false
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

	results := filterAndScoreDocuments(docs.Documents, query)
	sortResultsByScore(results)
	return applyResultLimit(results, query.Limit), nil
}

func filterAndScoreDocuments(docs []Document, query SearchQuery) []SearchResult {
	results := make([]SearchResult, 0, len(docs))
	queryLower := strings.ToLower(query.Text)

	for _, doc := range docs {
		if !matchesFilters(doc, query) {
			continue
		}
		results = append(results, SearchResult{
			Document: doc,
			Score:    calculateRelevanceScore(doc, queryLower),
			Snippet:  extractSnippet(doc.Content, queryLower, 200),
		})
	}
	return results
}

func matchesFilters(doc Document, query SearchQuery) bool {
	if query.SourceType != "" && doc.Source != query.SourceType {
		return false
	}
	if len(query.Tags) > 0 && !hasMatchingTag(doc.Tags, query.Tags) {
		return false
	}
	return true
}

func hasMatchingTag(docTags, queryTags []string) bool {
	for _, queryTag := range queryTags {
		queryTagLower := strings.ToLower(queryTag)
		for _, docTag := range docTags {
			if strings.Contains(strings.ToLower(docTag), queryTagLower) {
				return true
			}
		}
	}
	return false
}

func sortResultsByScore(results []SearchResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
}

func applyResultLimit(results []SearchResult, limit int) []SearchResult {
	if limit > 0 && len(results) > limit {
		return results[:limit]
	}
	return results
}

// Helper functions

func parseNotionResponse(body []byte, _ Source) ([]Document, error) {
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

	docs := make([]Document, 0, len(result.Results))
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

	docs := make([]Document, 0, len(result.Results))
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

func parseGitHubContents(ctx context.Context, client *http.Client, body []byte, _ Source, query string) ([]Document, error) {
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

	docs := make([]Document, 0, len(contents))
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
