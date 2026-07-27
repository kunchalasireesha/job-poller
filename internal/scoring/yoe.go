package scoring

import (
	"regexp"
	"strconv"
)

// yoePattern pairs a regexp with the index of the submatch group holding the
// minimum required years of experience. Patterns are tried in order and the
// first match wins.
type yoePattern struct {
	re    *regexp.Regexp
	group int
}

var yoePatterns = []yoePattern{
	// "3-5 years", "3 - 5 years of experience" -> take the lower bound.
	{regexp.MustCompile(`(?i)(\d+)\s*-\s*\d+\+?\s*years?`), 1},
	// "5+ years"
	{regexp.MustCompile(`(?i)(\d+)\+\s*years?`), 1},
	// "minimum 2 years", "at least 2 years", "min. 2 years"
	{regexp.MustCompile(`(?i)(?:minimum|at least|min\.?)\s*(?:of\s*)?(\d+)\+?\s*years?`), 1},
	// plain fallback: "2 years of experience", "2 years experience"
	{regexp.MustCompile(`(?i)(\d+)\+?\s*years?\s*(?:of\s*)?experience`), 1},
}

// ExtractYOE scans text for a years-of-experience requirement and returns the
// minimum required years, or nil if none is found.
func ExtractYOE(text string) *int {
	for _, p := range yoePatterns {
		m := p.re.FindStringSubmatch(text)
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[p.group])
		if err != nil {
			continue
		}
		return &n
	}
	return nil
}
