package scraper

import (
	"log/slog"
	"os"
	"testing"
)

func testParserLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

func loadFixture(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", path, err)
	}
	return string(data)
}

func TestParseJobs_ValidHTML(t *testing.T) {
	html := loadFixture(t, "testdata/sample.html")
	jobs, err := ParseJobs(html, testParserLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The fixture has 6 links total:
	// 3 job links with text (rows 1, 2, 3)
	// 1 empty href (row 4) → skipped
	// 1 whitespace title (row 5) → parsed with sanitization
	// 1 empty text (row 6) → skipped
	// 2 navigation links (Contact, About) → parsed
	// Total parsed: 3 + 1 + 2 = 6

	if len(jobs) < 5 {
		t.Errorf("expected at least 5 parsed jobs, got %d", len(jobs))
	}
}

func TestParseJobs_SkipsEmptyRows(t *testing.T) {
	html := `<html><body>
		<a href="">Empty link</a>
		<a href="http://example.com"></a>
		<a href="http://example.com">Valid</a>
	</body></html>`

	jobs, err := ParseJobs(html, testParserLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 1 {
		t.Errorf("expected 1 job (2 should be skipped), got %d", len(jobs))
	}
}

func TestParseJobs_ResolvesRelativeURLs(t *testing.T) {
	html := `<html><body>
		<a href="vacancy.php?id=1">Job A</a>
		<a href="/root/vacancy.php?id=2">Job B</a>
	</body></html>`

	jobs, err := ParseJobs(html, testParserLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(jobs))
	}

	// Relative URL (no leading slash)
	if jobs[0].Url != "https://recruitment.nic.in/vacancy.php?id=1" {
		t.Errorf("expected resolved URL, got %q", jobs[0].Url)
	}

	// Root-relative URL (leading slash)
	if jobs[1].Url != "https://recruitment.nic.in/root/vacancy.php?id=2" {
		t.Errorf("expected root-resolved URL, got %q", jobs[1].Url)
	}
}

func TestParseJobs_PreservesAbsoluteURLs(t *testing.T) {
	html := `<html><body>
		<a href="https://recruitment.nic.in/vacancy.php?id=100">Absolute Job</a>
	</body></html>`

	jobs, err := ParseJobs(html, testParserLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(jobs) != 1 {
		t.Fatalf("expected 1 job, got %d", len(jobs))
	}
	if jobs[0].Url != "https://recruitment.nic.in/vacancy.php?id=100" {
		t.Errorf("absolute URL should be preserved, got %q", jobs[0].Url)
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"trimWhitespace", "  hello world  ", "hello world"},
		{"collapseSpaces", "hello     world", "hello world"},
		{"collapseTabsNewlines", "hello\t\n\r  world", "hello world"},
		{"removeControlChars", "hello\x00\x01\x02world", "helloworld"},
		{"emptyString", "", ""},
		{"onlyWhitespace", "   \t\n  ", ""},
		{"mixedContent", "\n  Assistant   Director \t (IT)  \n", "Assistant Director (IT)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input)
			if got != tt.want {
				t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateID_Deterministic(t *testing.T) {
	url := "https://recruitment.nic.in/vacancy.php?id=42"
	id1 := GenerateID(url)
	id2 := GenerateID(url)

	if id1 != id2 {
		t.Errorf("GenerateID should be deterministic: got %q and %q", id1, id2)
	}

	if len(id1) != 32 {
		t.Errorf("expected 32 hex chars, got %d (%q)", len(id1), id1)
	}
}

func TestGenerateID_UniquenessForDifferentURLs(t *testing.T) {
	id1 := GenerateID("https://example.com/a")
	id2 := GenerateID("https://example.com/b")

	if id1 == id2 {
		t.Errorf("expected unique IDs for different URLs, both got %q", id1)
	}
}

func TestParseJobs_EmptyHTML(t *testing.T) {
	jobs, err := ParseJobs("<html><body></body></html>", testParserLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs for empty HTML, got %d", len(jobs))
	}
}

func TestParseJobs_MalformedHTML(t *testing.T) {
	// goquery is tolerant of malformed HTML
	html := "<a href='test.php'>Unclosed"
	jobs, err := ParseJobs(html, testParserLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("expected 1 job from malformed HTML, got %d", len(jobs))
	}
}
