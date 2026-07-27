package scoring

import (
	"fmt"
	"testing"
)

func TestExtractYOE(t *testing.T) {
	cases := []struct {
		name string
		text string
		want *int
	}{
		{"plus form", "We're looking for someone with 5+ years of experience.", intPtr(5)},
		{"range form", "3-5 years of experience with backend systems.", intPtr(3)},
		{"range with spaces", "3 - 5 years experience required.", intPtr(3)},
		{"minimum keyword", "minimum 2 years of relevant experience", intPtr(2)},
		{"at least keyword", "at least 4 years building distributed systems", intPtr(4)},
		{"plain experience", "2 years of experience with Go", intPtr(2)},
		{"no match", "We're looking for a passionate engineer.", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractYOE(tc.text)
			if (got == nil) != (tc.want == nil) {
				t.Fatalf("ExtractYOE(%q) = %s, want %s", tc.text, fmtIntPtr(got), fmtIntPtr(tc.want))
			}
			if got != nil && *got != *tc.want {
				t.Fatalf("ExtractYOE(%q) = %d, want %d", tc.text, *got, *tc.want)
			}
		})
	}
}

func intPtr(n int) *int { return &n }

func fmtIntPtr(p *int) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%d", *p)
}
