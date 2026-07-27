package scoring

import (
	"os"
	"testing"
)

func TestKeywordScorer_Score(t *testing.T) {
	s := KeywordScorer{Keywords: []string{"Go", "React", "Kubernetes"}}

	t.Run("all keywords in title score highest", func(t *testing.T) {
		got := s.Score("Senior Go React Kubernetes Engineer", "")
		if got != 1.0 {
			t.Fatalf("got %v, want 1.0", got)
		}
	})

	t.Run("description-only matches score lower than title matches", func(t *testing.T) {
		titleScore := s.Score("Go Engineer", "")
		descScore := s.Score("Software Engineer", "We use Go extensively.")
		if descScore >= titleScore {
			t.Fatalf("expected description match (%v) to score lower than title match (%v)", descScore, titleScore)
		}
	})

	t.Run("no matches scores zero", func(t *testing.T) {
		got := s.Score("Sales Manager", "Grow our enterprise pipeline.")
		if got != 0 {
			t.Fatalf("got %v, want 0", got)
		}
	})

	t.Run("word boundary avoids substring false positives", func(t *testing.T) {
		// "Go" should not match "Google" or "Golang" as a substring.
		got := s.Score("Golang Engineer at Google", "")
		if got != 0 {
			t.Fatalf("got %v, want 0 (word-boundary match should not fire on substrings)", got)
		}
	})

	t.Run("empty keyword list scores zero", func(t *testing.T) {
		empty := KeywordScorer{}
		if got := empty.Score("anything", "anything"); got != 0 {
			t.Fatalf("got %v, want 0", got)
		}
	})
}

func TestLoadKeywords(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/keywords.yaml"
	content := "keywords:\n  - Java\n  - Spring Boot\n  - \" \"\n  - Go\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := LoadKeywords(path)
	if err != nil {
		t.Fatalf("LoadKeywords: %v", err)
	}
	want := []string{"Java", "Spring Boot", "Go"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}
