// Package rag provides Retrieval-Augmented Generation support
// for enriching code reviews with external documentation.
package rag

import "time"

// SourceType categorizes the type of documentation source.
type SourceType string

const (
	SourceTypeStyleGuide   SourceType = "style_guide"
	SourceTypeSecurity     SourceType = "security"
	SourceTypeBestPractice SourceType = "best_practice"
	SourceTypeAPI          SourceType = "api"
	SourceTypeFramework    SourceType = "framework"
)

// Source represents an external documentation source.
type Source struct {
	URL      string     `yaml:"url" json:"url"`
	Type     SourceType `yaml:"type" json:"type"`
	Name     string     `yaml:"name" json:"name"`
	Language string     `yaml:"language,omitempty" json:"language,omitempty"`
	CacheTTL string     `yaml:"cache_ttl,omitempty" json:"cache_ttl,omitempty"`
	Enabled  bool       `yaml:"enabled" json:"enabled"`
}

// CachedDocument represents a cached documentation entry.
type CachedDocument struct {
	URL         string    `json:"url"`
	Content     string    `json:"content"`
	Summary     string    `json:"summary"`
	FetchedAt   time.Time `json:"fetched_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	ContentHash string    `json:"content_hash"`
}

// RAGConfig configures the RAG system.
//
//nolint:revive // RAGConfig name is intentional for clarity
type RAGConfig struct {
	Enabled         bool     `yaml:"enabled" mapstructure:"enabled"`
	CacheDir        string   `yaml:"cache_dir" mapstructure:"cache_dir"`
	DefaultCacheTTL string   `yaml:"default_cache_ttl" mapstructure:"default_cache_ttl"`
	MaxCacheSize    int64    `yaml:"max_cache_size" mapstructure:"max_cache_size"` // bytes
	Sources         []Source `yaml:"sources" mapstructure:"sources"`
	AutoDetect      bool     `yaml:"auto_detect" mapstructure:"auto_detect"` // auto-detect frameworks
}

// Context represents RAG context to be injected into prompts.
type Context struct {
	Sources     []SourceContext `json:"sources"`
	TotalTokens int             `json:"total_tokens"`
}

// SourceContext represents context from a single source.
type SourceContext struct {
	Name     string     `json:"name"`
	Type     SourceType `json:"type"`
	Content  string     `json:"content"`
	URL      string     `json:"url"`
	Relevant bool       `json:"relevant"`
}

// DetectedFramework represents an auto-detected framework/library.
type DetectedFramework struct {
	Name       string  `json:"name"`
	Version    string  `json:"version,omitempty"`
	Language   string  `json:"language"`
	DocsURL    string  `json:"docs_url,omitempty"`
	Confidence float64 `json:"confidence"`
}
