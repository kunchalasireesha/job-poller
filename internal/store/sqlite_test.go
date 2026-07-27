package store

import (
	"path/filepath"
	"testing"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "jobs.db")
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestJobIDDeterministic(t *testing.T) {
	id1 := JobID("greenhouse", "example-co", "12345")
	id2 := JobID("greenhouse", "example-co", "12345")
	if id1 != id2 {
		t.Fatalf("expected same id for same inputs, got %q and %q", id1, id2)
	}

	id3 := JobID("greenhouse", "example-co", "67890")
	if id1 == id3 {
		t.Fatalf("expected different ids for different external ids")
	}

	id4 := JobID("lever", "example-co", "12345")
	if id1 == id4 {
		t.Fatalf("expected different ids across different ATS sources")
	}
}

func TestInsertAndExists(t *testing.T) {
	s := newTestStore(t)

	id := JobID("greenhouse", "example-co", "12345")
	exists, err := s.Exists(id)
	if err != nil {
		t.Fatalf("Exists: %v", err)
	}
	if exists {
		t.Fatalf("expected job to not exist before insert")
	}

	yoe := 5
	err = s.Insert(Job{
		ID:             id,
		Company:        "example-co",
		Title:          "Senior Software Engineer",
		URL:            "https://example.com/jobs/12345",
		PostedAt:       "2026-07-01T00:00:00Z",
		RelevanceScore: 0.8,
		YOERequired:    &yoe,
	})
	if err != nil {
		t.Fatalf("Insert: %v", err)
	}

	exists, err = s.Exists(id)
	if err != nil {
		t.Fatalf("Exists after insert: %v", err)
	}
	if !exists {
		t.Fatalf("expected job to exist after insert")
	}
}

func TestInsertDedupesOnConflict(t *testing.T) {
	s := newTestStore(t)
	id := JobID("lever", "another-co", "abc")

	job := Job{ID: id, Company: "another-co", Title: "Backend Engineer", RelevanceScore: 0.5}
	if err := s.Insert(job); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	// Second insert with same id should not error and should not duplicate.
	job.Title = "Different Title"
	if err := s.Insert(job); err != nil {
		t.Fatalf("second insert: %v", err)
	}

	var count int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM jobs WHERE id = ?`, id).Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly 1 row for id, got %d", count)
	}
}

func TestCountUnviewedRelevant(t *testing.T) {
	s := newTestStore(t)

	jobs := []Job{
		{ID: JobID("greenhouse", "co", "1"), RelevanceScore: 0.9, Viewed: false},
		{ID: JobID("greenhouse", "co", "2"), RelevanceScore: 0.9, Viewed: true},
		{ID: JobID("greenhouse", "co", "3"), RelevanceScore: 0.1, Viewed: false},
		{ID: JobID("greenhouse", "co", "4"), RelevanceScore: 0.6, Viewed: false},
	}
	for _, j := range jobs {
		if err := s.Insert(j); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	count, err := s.CountUnviewedRelevant(0.5)
	if err != nil {
		t.Fatalf("CountUnviewedRelevant: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 unviewed relevant jobs, got %d", count)
	}
}

func TestMarkViewed(t *testing.T) {
	s := newTestStore(t)
	id := JobID("greenhouse", "co", "1")

	if err := s.Insert(Job{ID: id, RelevanceScore: 0.9, Viewed: false}); err != nil {
		t.Fatalf("insert: %v", err)
	}

	count, err := s.CountUnviewedRelevant(0.5)
	if err != nil {
		t.Fatalf("CountUnviewedRelevant: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 unviewed relevant job before marking viewed, got %d", count)
	}

	if err := s.MarkViewed(id); err != nil {
		t.Fatalf("MarkViewed: %v", err)
	}

	count, err = s.CountUnviewedRelevant(0.5)
	if err != nil {
		t.Fatalf("CountUnviewedRelevant after mark: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 unviewed relevant jobs after marking viewed, got %d", count)
	}
}
