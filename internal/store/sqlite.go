// Package store handles persistence of seen job postings in SQLite.
package store

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS jobs (
	id TEXT PRIMARY KEY,
	company TEXT,
	title TEXT,
	url TEXT,
	posted_at TEXT,
	first_seen_at TEXT,
	relevance_score REAL,
	yoe_required INTEGER,
	viewed INTEGER DEFAULT 0
);
`

// Job represents a single job posting row.
type Job struct {
	ID              string
	Company         string
	Title           string
	URL             string
	PostedAt        string
	FirstSeenAt     string
	RelevanceScore  float64
	YOERequired     *int
	Viewed          bool
}

// Store wraps a SQLite database of seen jobs.
type Store struct {
	db *sql.DB
}

// Open opens (creating if needed) the SQLite database at path and ensures the schema exists.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close closes the underlying database handle.
func (s *Store) Close() error {
	return s.db.Close()
}

// JobID computes the deterministic id for a job from its ATS name, company, and
// external id assigned by that ATS.
func JobID(ats, company, externalID string) string {
	sum := sha256.Sum256([]byte(ats + "|" + company + "|" + externalID))
	return hex.EncodeToString(sum[:])
}

// Exists reports whether a job with the given id is already stored.
func (s *Store) Exists(id string) (bool, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(1) FROM jobs WHERE id = ?`, id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check existence: %w", err)
	}
	return count > 0, nil
}

// Insert adds a new job row. FirstSeenAt is set to now if empty.
func (s *Store) Insert(j Job) error {
	if j.FirstSeenAt == "" {
		j.FirstSeenAt = time.Now().UTC().Format(time.RFC3339)
	}
	_, err := s.db.Exec(`
		INSERT INTO jobs (id, company, title, url, posted_at, first_seen_at, relevance_score, yoe_required, viewed)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO NOTHING
	`, j.ID, j.Company, j.Title, j.URL, j.PostedAt, j.FirstSeenAt, j.RelevanceScore, j.YOERequired, boolToInt(j.Viewed))
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	return nil
}

// CountUnviewedRelevant returns the number of unviewed jobs with relevance_score
// strictly greater than threshold.
func (s *Store) CountUnviewedRelevant(threshold float64) (int, error) {
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(1) FROM jobs WHERE viewed = 0 AND relevance_score > ?
	`, threshold).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unviewed relevant: %w", err)
	}
	return count, nil
}

// MarkViewed sets viewed = 1 for the given job id.
func (s *Store) MarkViewed(id string) error {
	_, err := s.db.Exec(`UPDATE jobs SET viewed = 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("mark viewed: %w", err)
	}
	return nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
