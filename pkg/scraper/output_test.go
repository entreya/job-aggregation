package scraper

import (
	"encoding/csv"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/entreya/job-aggregation/pkg/models"
)

func testOutputLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

func testJobs() []*models.JobPosting {
	return []*models.JobPosting{
		{
			Id:         "test-id-1",
			Title:      "Assistant Director",
			Department: "NIC",
			Location:   "All India",
			Url:        "https://recruitment.nic.in/vacancy.php?id=1",
			Date:       "2026-02-26",
		},
		{
			Id:         "test-id-2",
			Title:      "Junior Engineer",
			Department: "NIC",
			Location:   "Delhi",
			Url:        "https://recruitment.nic.in/vacancy.php?id=2",
			Date:       "2026-02-26",
		},
	}
}

func TestAppendResults_JSON_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	cfg := OutputConfig{Dir: dir, Format: "json"}

	err := AppendResults(testJobs(), cfg, testOutputLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(filepath.Join(dir, "data.json"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	var records []OutputRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 records, got %d", len(records))
	}
}

func TestAppendResults_JSON_AppendsNotOverwrites(t *testing.T) {
	dir := t.TempDir()
	cfg := OutputConfig{Dir: dir, Format: "json"}

	// First write
	err := AppendResults(testJobs()[:1], cfg, testOutputLogger())
	if err != nil {
		t.Fatalf("first write error: %v", err)
	}

	// Second write
	err = AppendResults(testJobs()[1:], cfg, testOutputLogger())
	if err != nil {
		t.Fatalf("second write error: %v", err)
	}

	// Read back
	data, err := os.ReadFile(filepath.Join(dir, "data.json"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	var records []OutputRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}

	if len(records) != 2 {
		t.Errorf("expected 2 records after append, got %d", len(records))
	}
}

func TestAppendResults_JSON_HasScrapedAt(t *testing.T) {
	dir := t.TempDir()
	cfg := OutputConfig{Dir: dir, Format: "json"}

	err := AppendResults(testJobs()[:1], cfg, testOutputLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "data.json"))
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	var records []OutputRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if len(records) == 0 {
		t.Fatal("expected at least 1 record")
	}

	if records[0].ScrapedAt == "" {
		t.Error("scraped_at field is empty")
	}
}

func TestAppendResults_CSV_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	cfg := OutputConfig{Dir: dir, Format: "csv"}

	err := AppendResults(testJobs(), cfg, testOutputLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read and verify
	file, err := os.Open(filepath.Join(dir, "data.csv"))
	if err != nil {
		t.Fatalf("failed to open CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// 1 header + 2 data rows = 3 total
	if len(rows) != 3 {
		t.Errorf("expected 3 rows (header + 2 data), got %d", len(rows))
	}

	// Check header
	expectedHeader := []string{"id", "title", "department", "location", "url", "date", "scraped_at"}
	for i, h := range expectedHeader {
		if rows[0][i] != h {
			t.Errorf("header column %d: expected %q, got %q", i, h, rows[0][i])
		}
	}
}

func TestAppendResults_CSV_AppendsWithoutDuplicateHeader(t *testing.T) {
	dir := t.TempDir()
	cfg := OutputConfig{Dir: dir, Format: "csv"}

	// First write (creates header)
	err := AppendResults(testJobs()[:1], cfg, testOutputLogger())
	if err != nil {
		t.Fatalf("first write error: %v", err)
	}

	// Second write (should NOT add another header)
	err = AppendResults(testJobs()[1:], cfg, testOutputLogger())
	if err != nil {
		t.Fatalf("second write error: %v", err)
	}

	file, err := os.Open(filepath.Join(dir, "data.csv"))
	if err != nil {
		t.Fatalf("failed to open CSV: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to parse CSV: %v", err)
	}

	// 1 header + 1 row + 1 row = 3 total
	if len(rows) != 3 {
		t.Errorf("expected 3 rows (1 header + 2 data), got %d", len(rows))
	}
}

func TestAppendResults_AutoCreatesDirectory(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "output")
	cfg := OutputConfig{Dir: dir, Format: "json"}

	err := AppendResults(testJobs(), cfg, testOutputLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "data.json")); os.IsNotExist(err) {
		t.Error("expected output file to be created in nested directory")
	}
}

func TestAppendResults_EmptyJobs_NoFile(t *testing.T) {
	dir := t.TempDir()
	cfg := OutputConfig{Dir: dir, Format: "json"}

	err := AppendResults(nil, cfg, testOutputLogger())
	if err != nil {
		t.Fatalf("unexpected error for empty jobs: %v", err)
	}

	// Should not create a file
	path := filepath.Join(dir, "data.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("expected no file to be created for empty jobs")
	}
}
