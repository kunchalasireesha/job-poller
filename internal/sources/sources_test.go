package sources

import (
	"net/http"
	"testing"
	"time"
)

func defaultTestClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

func TestStripHTML(t *testing.T) {
	in := "&lt;h2&gt;Who we are&lt;/h2&gt;\n&lt;p&gt;We build things &amp; ship them.&lt;/p&gt;"
	want := "Who we are We build things & ship them."
	got := stripHTML(in)
	if got != want {
		t.Fatalf("stripHTML mismatch:\n got:  %q\n want: %q", got, want)
	}
}

// TestFetchers_LiveAPI hits the real public Greenhouse and Lever APIs. It is
// skipped unless -short is not set and network access is available, since
// it depends on external services being up.
func TestFetchers_LiveAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live API test in -short mode")
	}

	t.Run("greenhouse", func(t *testing.T) {
		jobs, err := GreenhouseFetcher{}.Fetch(defaultTestClient(), "stripe")
		if err != nil {
			t.Fatalf("Fetch: %v", err)
		}
		if len(jobs) == 0 {
			t.Fatalf("expected at least one job from greenhouse/stripe")
		}
		j := jobs[0]
		if j.ATS != "greenhouse" || j.ExternalID == "" || j.Title == "" || j.URL == "" {
			t.Fatalf("unexpected job shape: %+v", j)
		}
	})

	t.Run("lever", func(t *testing.T) {
		jobs, err := LeverFetcher{}.Fetch(defaultTestClient(), "leverdemo")
		if err != nil {
			t.Fatalf("Fetch: %v", err)
		}
		if len(jobs) == 0 {
			t.Fatalf("expected at least one job from lever/leverdemo")
		}
		j := jobs[0]
		if j.ATS != "lever" || j.ExternalID == "" || j.Title == "" || j.URL == "" {
			t.Fatalf("unexpected job shape: %+v", j)
		}
	})
}
