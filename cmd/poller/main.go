// Command poller runs one pass of the job-poller pipeline: fetch postings
// from configured ATS boards, dedupe against previously seen jobs, score
// relevance, persist new jobs, and notify via ntfy if enough unviewed
// relevant jobs have accumulated.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"job-poller/internal/config"
	"job-poller/internal/notify"
	"job-poller/internal/scoring"
	"job-poller/internal/sources"
	"job-poller/internal/store"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "mark-viewed" {
		markViewedCmd := flag.NewFlagSet("mark-viewed", flag.ExitOnError)
		dbPath := markViewedCmd.String("db", "jobs.db", "path to the SQLite jobs database")
		markViewedCmd.Parse(os.Args[2:])

		if markViewedCmd.NArg() != 1 {
			log.Fatalf("usage: poller mark-viewed [-db path] <job-id>")
		}
		if err := markViewed(*dbPath, markViewedCmd.Arg(0)); err != nil {
			log.Fatalf("poller: %v", err)
		}
		return
	}

	companiesPath := flag.String("companies", "config/companies.yaml", "path to companies.yaml")
	keywordsPath := flag.String("keywords", "config/keywords.yaml", "path to keywords.yaml")
	dbPath := flag.String("db", "jobs.db", "path to the SQLite jobs database")
	relevanceThreshold := flag.Float64("relevance-threshold", 0.3, "minimum relevance score (0-1) to count a job as relevant")
	notifyThreshold := flag.Int("notify-threshold", 2, "notify when unviewed relevant job count exceeds this")
	ntfyTopic := flag.String("ntfy-topic", os.Getenv("NTFY_TOPIC"), "ntfy topic to publish to (defaults to $NTFY_TOPIC)")
	ntfyBaseURL := flag.String("ntfy-base-url", "https://ntfy.sh", "ntfy server base URL")
	dryRun := flag.Bool("dry-run", false, "fetch/score/persist as normal but only log the notification instead of sending it")
	flag.Parse()

	if err := run(*companiesPath, *keywordsPath, *dbPath, *relevanceThreshold, *notifyThreshold, *ntfyTopic, *ntfyBaseURL, *dryRun); err != nil {
		log.Fatalf("poller: %v", err)
	}
}

func markViewed(dbPath, id string) error {
	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	if err := db.MarkViewed(id); err != nil {
		return fmt.Errorf("mark viewed: %w", err)
	}
	log.Printf("marked %s as viewed", id)
	return nil
}

func run(companiesPath, keywordsPath, dbPath string, relevanceThreshold float64, notifyThreshold int, ntfyTopic, ntfyBaseURL string, dryRun bool) error {
	companies, err := config.LoadCompanies(companiesPath)
	if err != nil {
		return fmt.Errorf("load companies: %w", err)
	}
	keywords, err := scoring.LoadKeywords(keywordsPath)
	if err != nil {
		return fmt.Errorf("load keywords: %w", err)
	}
	scorer := scoring.KeywordScorer{Keywords: keywords}

	db, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}
	defer db.Close()

	httpClient := &http.Client{Timeout: 15 * time.Second}
	fetchers := map[string]sources.Fetcher{
		"greenhouse": sources.GreenhouseFetcher{},
		"lever":      sources.LeverFetcher{},
	}

	var newJobs, skipped int
	for _, c := range companies {
		fetcher, ok := fetchers[c.ATS]
		if !ok {
			log.Printf("skipping %s: unknown ats %q", c.Name, c.ATS)
			continue
		}

		raw, err := fetcher.Fetch(httpClient, c.Board)
		if err != nil {
			log.Printf("fetch %s (%s): %v", c.Name, c.ATS, err)
			continue
		}

		for _, j := range raw {
			id := store.JobID(j.ATS, j.Company, j.ExternalID)
			exists, err := db.Exists(id)
			if err != nil {
				log.Printf("check exists %s: %v", id, err)
				continue
			}
			if exists {
				skipped++
				continue
			}

			score := scorer.Score(j.Title, j.Description)
			yoe := scoring.ExtractYOE(j.Description)

			if err := db.Insert(store.Job{
				ID:             id,
				Company:        c.Name,
				Title:          j.Title,
				URL:            j.URL,
				PostedAt:       j.PostedAt,
				RelevanceScore: score,
				YOERequired:    yoe,
			}); err != nil {
				log.Printf("insert %s: %v", id, err)
				continue
			}
			newJobs++
		}
	}
	log.Printf("processed %d companies: %d new jobs, %d already seen", len(companies), newJobs, skipped)

	count, err := db.CountUnviewedRelevant(relevanceThreshold)
	if err != nil {
		return fmt.Errorf("count unviewed relevant: %w", err)
	}
	log.Printf("unviewed relevant jobs: %d (threshold: %d)", count, notifyThreshold)

	if count <= notifyThreshold {
		return nil
	}

	message := fmt.Sprintf("You have %d unviewed relevant postings", count)
	if dryRun {
		log.Printf("[dry-run] would notify: %s", message)
		return nil
	}

	notifier := notify.Notifier{BaseURL: ntfyBaseURL, Topic: ntfyTopic}
	if err := notifier.Notify("New relevant jobs", message); err != nil {
		return fmt.Errorf("notify: %w", err)
	}
	return nil
}
