// Package scoring computes a relevance score for job postings and extracts
// structured signals like years-of-experience requirements.
package scoring

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Scorer assigns a relevance score in [0, 1] to a job posting based on its
// title and description. Kept as an interface so a v2 embedding-based
// scorer can be swapped in without touching callers.
type Scorer interface {
	Score(title, description string) float64
}

// KeywordScorer scores relevance as the fraction of configured keywords found
// in the job title/description, weighting title matches higher than
// description-only matches.
type KeywordScorer struct {
	Keywords []string
}

const (
	titleMatchWeight = 2.0
	descMatchWeight  = 1.0
)

// Score implements Scorer.
func (s KeywordScorer) Score(title, description string) float64 {
	if len(s.Keywords) == 0 {
		return 0
	}

	var matched float64
	for _, kw := range s.Keywords {
		re := keywordRegexp(kw)
		switch {
		case re.MatchString(title):
			matched += titleMatchWeight
		case re.MatchString(description):
			matched += descMatchWeight
		}
	}

	maxPossible := float64(len(s.Keywords)) * titleMatchWeight
	return matched / maxPossible
}

// keywordRegexp builds a case-insensitive, word-boundary regexp for a
// keyword phrase, escaping any regexp metacharacters it contains (e.g. "C++").
func keywordRegexp(keyword string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(keyword)
	return regexp.MustCompile(`(?i)\b` + escaped + `\b`)
}

// KeywordsConfig is the shape of the keywords YAML config file.
type KeywordsConfig struct {
	Keywords []string `yaml:"keywords"`
}

// LoadKeywords reads a keyword list from a YAML file of the form:
//
//	keywords:
//	  - Java
//	  - Spring Boot
func LoadKeywords(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read keywords config %s: %w", path, err)
	}

	var cfg KeywordsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse keywords config %s: %w", path, err)
	}

	keywords := make([]string, 0, len(cfg.Keywords))
	for _, kw := range cfg.Keywords {
		kw = strings.TrimSpace(kw)
		if kw != "" {
			keywords = append(keywords, kw)
		}
	}
	return keywords, nil
}
