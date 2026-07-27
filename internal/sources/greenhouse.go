package sources

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
)

// GreenhouseFetcher fetches postings from the Greenhouse Job Board API:
// https://boards-api.greenhouse.io/v1/boards/{board}/jobs
type GreenhouseFetcher struct{}

type greenhouseResponse struct {
	Jobs []greenhouseJob `json:"jobs"`
}

type greenhouseJob struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	AbsoluteURL string `json:"absolute_url"`
	UpdatedAt   string `json:"updated_at"`
	Content     string `json:"content"`
}

func (f GreenhouseFetcher) Fetch(client *http.Client, board string) ([]RawJob, error) {
	url := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true", board)
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("greenhouse fetch %s: %w", board, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("greenhouse fetch %s: unexpected status %d", board, resp.StatusCode)
	}

	var parsed greenhouseResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("greenhouse decode %s: %w", board, err)
	}

	jobs := make([]RawJob, 0, len(parsed.Jobs))
	for _, j := range parsed.Jobs {
		jobs = append(jobs, RawJob{
			ATS:         "greenhouse",
			Company:     board,
			ExternalID:  fmt.Sprintf("%d", j.ID),
			Title:       j.Title,
			URL:         j.AbsoluteURL,
			PostedAt:    j.UpdatedAt,
			Description: stripHTML(j.Content),
		})
	}
	return jobs, nil
}

var htmlTagRE = regexp.MustCompile(`<[^>]*>`)

// stripHTML decodes HTML entities and removes tags, giving a plain-text
// approximation of an HTML job description suitable for keyword/regex scoring.
// Greenhouse's "content" field is entity-encoded (e.g. "&lt;h2&gt;"), so
// entities must be decoded before tags become visible to strip.
func stripHTML(s string) string {
	s = html.UnescapeString(s)
	s = htmlTagRE.ReplaceAllString(s, " ")
	return strings.Join(strings.Fields(s), " ")
}
