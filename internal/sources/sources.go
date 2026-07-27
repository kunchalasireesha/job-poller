// Package sources fetches current job postings from supported ATS platforms.
package sources

import "net/http"

// RawJob is a job posting as fetched from an ATS, before dedup/scoring.
type RawJob struct {
	ATS         string // "greenhouse" or "lever"
	Company     string // board/company slug from companies.yaml
	ExternalID  string // ATS-assigned id, unique within (ATS, Company)
	Title       string
	URL         string
	PostedAt    string // RFC3339 if known, else empty
	Description string // plain text, used for scoring/YOE extraction
}

// Fetcher pulls the current list of job postings for one company board.
type Fetcher interface {
	Fetch(client *http.Client, board string) ([]RawJob, error)
}
