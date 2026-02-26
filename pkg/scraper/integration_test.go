//go:build integration

package scraper

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func testIntegrationLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

// TestIntegration_FullPipeline tests the complete scrape pipeline:
// fetch HTML from a local test server → parse → output → verify file.
//
// This test does NOT use chromedp — it tests the parse + output pipeline
// using raw HTML served by httptest. For full chromedp integration,
// a real browser is needed (CI only).
func TestIntegration_FullPipeline(t *testing.T) {
	// Serve the test fixture HTML
	fixture, err := os.ReadFile("testdata/sample.html")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(fixture)
	}))
	defer server.Close()

	logger := testIntegrationLogger()

	// Step 1: Fetch HTML from test server
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("failed to fetch from test server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// Read body
	body := make([]byte, 0, 4096)
	buf := make([]byte, 1024)
	for {
		n, readErr := resp.Body.Read(buf)
		body = append(body, buf[:n]...)
		if readErr != nil {
			break
		}
	}

	htmlContent := string(body)
	if len(htmlContent) < 50 {
		t.Fatalf("response too short: %d bytes", len(htmlContent))
	}

	// Step 2: Parse HTML
	jobs, err := ParseJobs(htmlContent, logger)
	if err != nil {
		t.Fatalf("ParseJobs failed: %v", err)
	}

	if len(jobs) == 0 {
		t.Fatal("expected at least 1 job, got 0")
	}

	t.Logf("parsed %d jobs from test server", len(jobs))

	// Verify job structure
	for i, j := range jobs {
		if j.Id == "" {
			t.Errorf("job %d: ID is empty", i)
		}
		if j.Title == "" {
			t.Errorf("job %d: Title is empty", i)
		}
		if j.Url == "" {
			t.Errorf("job %d: URL is empty", i)
		}
	}

	// Step 3: Write to JSON output
	outputDir := filepath.Join(t.TempDir(), "output")
	cfg := OutputConfig{Dir: outputDir, Format: "json"}

	err = AppendResults(jobs, cfg, logger)
	if err != nil {
		t.Fatalf("AppendResults failed: %v", err)
	}

	// Step 4: Verify output file
	outputPath := filepath.Join(outputDir, "data.json")
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	var records []OutputRecord
	if err := json.Unmarshal(data, &records); err != nil {
		t.Fatalf("failed to parse output JSON: %v", err)
	}

	if len(records) != len(jobs) {
		t.Errorf("expected %d records in output, got %d", len(jobs), len(records))
	}

	// Verify scraped_at is present on all records
	for i, r := range records {
		if r.ScrapedAt == "" {
			t.Errorf("record %d: scraped_at is empty", i)
		}
		if r.ID == "" {
			t.Errorf("record %d: ID is empty", i)
		}
		if r.Title == "" {
			t.Errorf("record %d: Title is empty", i)
		}
	}

	// Step 5: Verify CSV output too
	csvCfg := OutputConfig{Dir: outputDir, Format: "csv"}
	err = AppendResults(jobs, csvCfg, logger)
	if err != nil {
		t.Fatalf("CSV AppendResults failed: %v", err)
	}

	csvPath := filepath.Join(outputDir, "data.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		t.Error("expected CSV output file to exist")
	}

	t.Log("integration test passed: fetch → parse → JSON output → CSV output")
}
