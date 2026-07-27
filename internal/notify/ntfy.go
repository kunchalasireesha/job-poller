// Package notify sends push notifications via ntfy.sh (https://ntfy.sh).
package notify

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPClient is the subset of *http.Client used by Notifier, allowing tests
// to inject a fake implementation.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Notifier posts messages to an ntfy topic.
type Notifier struct {
	// BaseURL is the ntfy server base URL, e.g. "https://ntfy.sh". Defaults
	// to https://ntfy.sh if empty.
	BaseURL string
	// Topic is the (ideally hard-to-guess) ntfy topic name to publish to.
	Topic  string
	Client HTTPClient
}

const defaultBaseURL = "https://ntfy.sh"

// Notify posts a message with the given title to the configured ntfy topic.
func (n Notifier) Notify(title, message string) error {
	if n.Topic == "" {
		return fmt.Errorf("notify: topic is required")
	}

	base := n.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	url := strings.TrimRight(base, "/") + "/" + n.Topic

	client := n.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(message))
	if err != nil {
		return fmt.Errorf("notify: build request: %w", err)
	}
	req.Header.Set("Title", title)
	req.Header.Set("Priority", "default")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("notify: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notify: unexpected status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
