package history

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Store provides SQLite-based review history storage.
type Store struct {
	db *sql.DB
}

// StoreConfig configures the history store.
type StoreConfig struct {
	// Path is the SQLite database file path
	Path string
}

// NewStore creates a new history store.
func NewStore(cfg StoreConfig) (*Store, error) {
	db, err := sql.Open("sqlite", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting WAL mode: %w", err)
	}

	store := &Store{db: db}

	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	return store, nil
}

// migrate runs database migrations.
func (s *Store) migrate() error {
	migrations := []string{
		// Main reviews table
		`CREATE TABLE IF NOT EXISTS reviews (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			commit_hash TEXT NOT NULL,
			file_path TEXT NOT NULL,
			issue_type TEXT NOT NULL,
			severity TEXT NOT NULL,
			message TEXT NOT NULL,
			suggestion TEXT,
			line INTEGER,
			author TEXT,
			branch TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved BOOLEAN DEFAULT FALSE,
			resolved_at DATETIME,
			review_round INTEGER DEFAULT 1
		)`,

		// Full-text search virtual table
		`CREATE VIRTUAL TABLE IF NOT EXISTS reviews_fts USING fts5(
			message,
			suggestion,
			content='reviews',
			content_rowid='id'
		)`,

		// Triggers to keep FTS in sync
		`CREATE TRIGGER IF NOT EXISTS reviews_ai AFTER INSERT ON reviews BEGIN
			INSERT INTO reviews_fts(rowid, message, suggestion)
			VALUES (new.id, new.message, new.suggestion);
		END`,

		`CREATE TRIGGER IF NOT EXISTS reviews_ad AFTER DELETE ON reviews BEGIN
			INSERT INTO reviews_fts(reviews_fts, rowid, message, suggestion)
			VALUES ('delete', old.id, old.message, old.suggestion);
		END`,

		`CREATE TRIGGER IF NOT EXISTS reviews_au AFTER UPDATE ON reviews BEGIN
			INSERT INTO reviews_fts(reviews_fts, rowid, message, suggestion)
			VALUES ('delete', old.id, old.message, old.suggestion);
			INSERT INTO reviews_fts(rowid, message, suggestion)
			VALUES (new.id, new.message, new.message);
		END`,

		// Indexes for common queries
		`CREATE INDEX IF NOT EXISTS idx_reviews_file ON reviews(file_path)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_commit ON reviews(commit_hash)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_author ON reviews(author)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_severity ON reviews(severity)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_created ON reviews(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_reviews_resolved ON reviews(resolved)`,
	}

	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// Store saves a review record.
func (s *Store) Store(ctx context.Context, record *ReviewRecord) error {
	query := `INSERT INTO reviews (
		commit_hash, file_path, issue_type, severity, message, suggestion,
		line, author, branch, created_at, resolved, review_round
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := s.db.ExecContext(ctx, query,
		record.CommitHash, record.FilePath, record.IssueType, record.Severity,
		record.Message, record.Suggestion, record.Line, record.Author,
		record.Branch, record.CreatedAt, record.Resolved, record.ReviewRound,
	)
	if err != nil {
		return fmt.Errorf("inserting record: %w", err)
	}

	id, _ := result.LastInsertId()
	record.ID = id

	return nil
}

// StoreBatch saves multiple review records in a transaction.
func (s *Store) StoreBatch(ctx context.Context, records []*ReviewRecord) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO reviews (
		commit_hash, file_path, issue_type, severity, message, suggestion,
		line, author, branch, created_at, resolved, review_round
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("preparing statement: %w", err)
	}
	defer stmt.Close()

	for _, record := range records {
		result, err := stmt.ExecContext(ctx,
			record.CommitHash, record.FilePath, record.IssueType, record.Severity,
			record.Message, record.Suggestion, record.Line, record.Author,
			record.Branch, record.CreatedAt, record.Resolved, record.ReviewRound,
		)
		if err != nil {
			return fmt.Errorf("inserting record: %w", err)
		}
		id, _ := result.LastInsertId()
		record.ID = id
	}

	return tx.Commit()
}

// Search performs full-text search on review history.
//
//nolint:gocyclo,funlen // Complex query builder with multiple filter conditions
func (s *Store) Search(ctx context.Context, q SearchQuery) (*SearchResult, error) {
	var args []interface{}
	var conditions []string

	// Full-text search
	if q.Text != "" {
		conditions = append(conditions, "r.id IN (SELECT rowid FROM reviews_fts WHERE reviews_fts MATCH ?)")
		args = append(args, q.Text)
	}

	// File filter (supports LIKE patterns)
	if q.File != "" {
		pattern := strings.ReplaceAll(q.File, "*", "%")
		conditions = append(conditions, "r.file_path LIKE ?")
		args = append(args, pattern)
	}

	// Author filter
	if q.Author != "" {
		conditions = append(conditions, "r.author = ?")
		args = append(args, q.Author)
	}

	// Severity filter
	if q.Severity != "" {
		conditions = append(conditions, "r.severity = ?")
		args = append(args, q.Severity)
	}

	// Type filter
	if q.Type != "" {
		conditions = append(conditions, "r.issue_type = ?")
		args = append(args, q.Type)
	}

	// Branch filter
	if q.Branch != "" {
		conditions = append(conditions, "r.branch = ?")
		args = append(args, q.Branch)
	}

	// Date filters
	if !q.Since.IsZero() {
		conditions = append(conditions, "r.created_at >= ?")
		args = append(args, q.Since)
	}
	if !q.Until.IsZero() {
		conditions = append(conditions, "r.created_at <= ?")
		args = append(args, q.Until)
	}

	// Resolved filter
	if q.Resolved != nil {
		conditions = append(conditions, "r.resolved = ?")
		args = append(args, *q.Resolved)
	}

	// Build WHERE clause
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := "SELECT COUNT(*) FROM reviews r " + whereClause //nolint:gosec // Query built with parameterized args
	var totalCount int64
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("counting results: %w", err)
	}

	// Fetch results
	limit := q.Limit
	if limit <= 0 {
		limit = 100
	}
	offset := q.Offset

	//nolint:gosec // Query built with parameterized args, whereClause uses placeholders
	selectQuery := `
		SELECT id, commit_hash, file_path, issue_type, severity, message, suggestion,
		       line, author, branch, created_at, resolved, resolved_at, review_round
		FROM reviews r
		` + whereClause + `
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("querying records: %w", err)
	}
	defer rows.Close()

	records := make([]ReviewRecord, 0)
	for rows.Next() {
		var r ReviewRecord
		var resolvedAt sql.NullTime
		var suggestion, author, branch sql.NullString
		var line sql.NullInt64

		if err := rows.Scan(
			&r.ID, &r.CommitHash, &r.FilePath, &r.IssueType, &r.Severity,
			&r.Message, &suggestion, &line, &author, &branch,
			&r.CreatedAt, &r.Resolved, &resolvedAt, &r.ReviewRound,
		); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		if suggestion.Valid {
			r.Suggestion = suggestion.String
		}
		if line.Valid {
			r.Line = int(line.Int64)
		}
		if author.Valid {
			r.Author = author.String
		}
		if branch.Valid {
			r.Branch = branch.String
		}
		if resolvedAt.Valid {
			r.ResolvedAt = resolvedAt.Time
		}

		records = append(records, r)
	}

	return &SearchResult{
		Records:    records,
		TotalCount: totalCount,
		Query:      q,
	}, nil
}

// GetFileHistory returns the review history for a file or directory.
//
//nolint:gocyclo,funlen // Aggregation query with multiple statistics calculations
func (s *Store) GetFileHistory(ctx context.Context, path string) (*FileHistory, error) {
	pattern := path
	if strings.HasSuffix(path, "/") || !strings.Contains(filepath.Base(path), ".") {
		pattern = path + "%"
	}

	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN resolved THEN 1 ELSE 0 END) as resolved_count,
			MIN(created_at) as first_review,
			MAX(created_at) as last_review,
			MAX(review_round) as max_round
		FROM reviews
		WHERE file_path LIKE ?
	`

	var total int64
	var resolved sql.NullInt64
	var firstReviewStr, lastReviewStr sql.NullString
	var maxRound sql.NullInt64

	if err := s.db.QueryRowContext(ctx, query, pattern).Scan(
		&total, &resolved, &firstReviewStr, &lastReviewStr, &maxRound,
	); err != nil {
		return nil, fmt.Errorf("querying file history: %w", err)
	}

	var firstReview, lastReview time.Time
	if firstReviewStr.Valid {
		firstReview, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", firstReviewStr.String)
		if firstReview.IsZero() {
			firstReview, _ = time.Parse("2006-01-02T15:04:05Z", firstReviewStr.String)
		}
		if firstReview.IsZero() {
			firstReview, _ = time.Parse(time.RFC3339, firstReviewStr.String)
		}
	}
	if lastReviewStr.Valid {
		lastReview, _ = time.Parse("2006-01-02 15:04:05.999999999-07:00", lastReviewStr.String)
		if lastReview.IsZero() {
			lastReview, _ = time.Parse("2006-01-02T15:04:05Z", lastReviewStr.String)
		}
		if lastReview.IsZero() {
			lastReview, _ = time.Parse(time.RFC3339, lastReviewStr.String)
		}
	}

	// Get severity breakdown
	severityQuery := `
		SELECT severity, COUNT(*)
		FROM reviews
		WHERE file_path LIKE ?
		GROUP BY severity
	`
	bySeverity := make(map[string]int)
	rows, err := s.db.QueryContext(ctx, severityQuery, pattern)
	if err != nil {
		return nil, fmt.Errorf("querying severity breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sev string
		var count int
		if scanErr := rows.Scan(&sev, &count); scanErr != nil {
			continue
		}
		bySeverity[sev] = count
	}

	// Get type breakdown
	typeQuery := `
		SELECT issue_type, COUNT(*)
		FROM reviews
		WHERE file_path LIKE ?
		GROUP BY issue_type
	`
	byType := make(map[string]int)
	rows, err = s.db.QueryContext(ctx, typeQuery, pattern)
	if err != nil {
		return nil, fmt.Errorf("querying type breakdown: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var typ string
		var count int
		if scanErr := rows.Scan(&typ, &count); scanErr != nil {
			continue
		}
		byType[typ] = count
	}

	resolvedCount := resolved.Int64
	maxRoundVal := int(maxRound.Int64)

	return &FileHistory{
		Path:         path,
		TotalIssues:  total,
		Resolved:     resolvedCount,
		Pending:      total - resolvedCount,
		BySeverity:   bySeverity,
		ByType:       byType,
		FirstReview:  firstReview,
		LastReview:   lastReview,
		ReviewRounds: maxRoundVal,
	}, nil
}

// GetStats returns aggregate statistics.
func (s *Store) GetStats(ctx context.Context) (*Stats, error) {
	var total, resolved int64

	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*), SUM(CASE WHEN resolved THEN 1 ELSE 0 END)
		FROM reviews
	`).Scan(&total, &resolved); err != nil {
		return nil, fmt.Errorf("querying stats: %w", err)
	}

	// By severity
	bySeverity := make(map[string]int64)
	rows, _ := s.db.QueryContext(ctx, `SELECT severity, COUNT(*) FROM reviews GROUP BY severity`)
	for rows.Next() {
		var sev string
		var count int64
		if err := rows.Scan(&sev, &count); err == nil {
			bySeverity[sev] = count
		}
	}
	rows.Close()

	// By type
	byType := make(map[string]int64)
	rows, _ = s.db.QueryContext(ctx, `SELECT issue_type, COUNT(*) FROM reviews GROUP BY issue_type`)
	for rows.Next() {
		var typ string
		var count int64
		if err := rows.Scan(&typ, &count); err == nil {
			byType[typ] = count
		}
	}
	rows.Close()

	// Top files
	byFile := make(map[string]int64)
	rows, _ = s.db.QueryContext(ctx, `
		SELECT file_path, COUNT(*) as cnt
		FROM reviews
		GROUP BY file_path
		ORDER BY cnt DESC
		LIMIT 10
	`)
	for rows.Next() {
		var file string
		var count int64
		if err := rows.Scan(&file, &count); err == nil {
			byFile[file] = count
		}
	}
	rows.Close()

	return &Stats{
		TotalReviews:   0, // Would need separate tracking
		TotalIssues:    total,
		ResolvedIssues: resolved,
		BySeverity:     bySeverity,
		ByType:         byType,
		ByFile:         byFile,
	}, nil
}

// MarkResolved marks an issue as resolved.
func (s *Store) MarkResolved(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE reviews SET resolved = TRUE, resolved_at = ? WHERE id = ?
	`, time.Now(), id)
	return err
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
