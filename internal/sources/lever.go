package sources

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// LeverFetcher fetches postings from the Lever Postings API:
// https://api.lever.co/v0/postings/{board}
type LeverFetcher struct{}

type leverPosting struct {
	ID               string `json:"id"`
	Text             string `json:"text"` // job title
	HostedURL        string `json:"hostedUrl"`
	CreatedAt        int64  `json:"createdAt"` // ms since epoch
	DescriptionPlain string `json:"descriptionPlain"`
}

func (f LeverFetcher) Fetch(client *http.Client, board string) ([]RawJob, error) {
	url := fmt.Sprintf("https://api.lever.co/v0/postings/%s?mode=json", board)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("lever fetch %s: %w", board, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lever fetch %s: unexpected status %d", board, resp.StatusCode)
	}

	var postings []leverPosting
	if err := json.NewDecoder(resp.Body).Decode(&postings); err != nil {
		return nil, fmt.Errorf("lever decode %s: %w", board, err)
	}

	jobs := make([]RawJob, 0, len(postings))
	for _, p := range postings {
		var postedAt string
		if p.CreatedAt > 0 {
			postedAt = time.UnixMilli(p.CreatedAt).UTC().Format(time.RFC3339)
		}
		jobs = append(jobs, RawJob{
			ATS:         "lever",
			Company:     board,
			ExternalID:  p.ID,
			Title:       p.Text,
			URL:         p.HostedURL,
			PostedAt:    postedAt,
			Description: p.DescriptionPlain,
		})
	}
	return jobs, nil
}
