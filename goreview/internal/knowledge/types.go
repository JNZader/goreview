// Package knowledge provides integration with external knowledge systems
// like Notion, Confluence, and Obsidian vaults.
package knowledge

import "time"

// SourceType represents the type of knowledge source.
type SourceType string

const (
	SourceTypeNotion     SourceType = "notion"
	SourceTypeConfluence SourceType = "confluence"
	SourceTypeObsidian   SourceType = "obsidian"
	SourceTypeLocal      SourceType = "local"
	SourceTypeGitHub     SourceType = "github"
)

// Source represents a knowledge source configuration.
type Source struct {
	Type     SourceType `yaml:"type" json:"type"`
	Name     string     `yaml:"name" json:"name"`
	Enabled  bool       `yaml:"enabled" json:"enabled"`

	// Notion-specific
	NotionToken      string `yaml:"notion_token,omitempty" json:"notion_token,omitempty"`
	NotionDatabaseID string `yaml:"notion_database_id,omitempty" json:"notion_database_id,omitempty"`
	NotionPageID     string `yaml:"notion_page_id,omitempty" json:"notion_page_id,omitempty"`

	// Confluence-specific
	ConfluenceURL   string `yaml:"confluence_url,omitempty" json:"confluence_url,omitempty"`
	ConfluenceUser  string `yaml:"confluence_user,omitempty" json:"confluence_user,omitempty"`
	ConfluenceToken string `yaml:"confluence_token,omitempty" json:"confluence_token,omitempty"`
	ConfluenceSpace string `yaml:"confluence_space,omitempty" json:"confluence_space,omitempty"`

	// Obsidian-specific
	ObsidianVaultPath string   `yaml:"obsidian_vault_path,omitempty" json:"obsidian_vault_path,omitempty"`
	ObsidianTags      []string `yaml:"obsidian_tags,omitempty" json:"obsidian_tags,omitempty"`

	// Local docs
	LocalPath    string   `yaml:"local_path,omitempty" json:"local_path,omitempty"`
	LocalPattern string   `yaml:"local_pattern,omitempty" json:"local_pattern,omitempty"` // glob pattern

	// GitHub wiki/docs
	GitHubOwner string `yaml:"github_owner,omitempty" json:"github_owner,omitempty"`
	GitHubRepo  string `yaml:"github_repo,omitempty" json:"github_repo,omitempty"`
	GitHubPath  string `yaml:"github_path,omitempty" json:"github_path,omitempty"` // e.g., "wiki" or "docs"

	// Caching
	CacheTTL string `yaml:"cache_ttl,omitempty" json:"cache_ttl,omitempty"`
}

// Document represents a document from a knowledge source.
type Document struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Content   string            `json:"content"`
	URL       string            `json:"url,omitempty"`
	Source    SourceType        `json:"source"`
	Tags      []string          `json:"tags,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	FetchedAt time.Time         `json:"fetched_at"`
}

// SearchQuery represents a search query for knowledge sources.
type SearchQuery struct {
	Text       string     `json:"text"`
	Tags       []string   `json:"tags,omitempty"`
	SourceType SourceType `json:"source_type,omitempty"` // Empty = all sources
	Limit      int        `json:"limit,omitempty"`
}

// SearchResult represents a search result.
type SearchResult struct {
	Document Document `json:"document"`
	Score    float64  `json:"score"`
	Snippet  string   `json:"snippet"`
}

// Context represents aggregated knowledge context.
type Context struct {
	Documents   []Document `json:"documents"`
	TotalTokens int        `json:"total_tokens"`
}

// Config configures the knowledge integration system.
type Config struct {
	Enabled  bool     `yaml:"enabled" mapstructure:"enabled"`
	Sources  []Source `yaml:"sources" mapstructure:"sources"`
	CacheDir string   `yaml:"cache_dir" mapstructure:"cache_dir"`
	MaxDocs  int      `yaml:"max_docs" mapstructure:"max_docs"` // Max docs to include in context
}
