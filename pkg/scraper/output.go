package scraper

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/entreya/job-aggregation/pkg/models"
)

// OutputConfig defines where and how scraped data is persisted.
type OutputConfig struct {
	Dir    string // Output directory (e.g., "output")
	Format string // "json" or "csv"
}

// DefaultOutputConfig returns sensible defaults.
func DefaultOutputConfig() OutputConfig {
	return OutputConfig{
		Dir:    "output",
		Format: "json",
	}
}

// OutputRecord wraps a JobPosting with a scrape timestamp for output.
type OutputRecord struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Department string `json:"department"`
	Location   string `json:"location"`
	URL        string `json:"url"`
	Date       string `json:"date"`
	ScrapedAt  string `json:"scraped_at"`
}

// AppendResults writes job postings to the output file WITHOUT overwriting
// existing data. Each record includes a `scraped_at` UTC timestamp.
//
// On failure, logs the error and returns it (caller decides whether to crash).
func AppendResults(jobs []*models.JobPosting, cfg OutputConfig, logger *slog.Logger) error {
	if len(jobs) == 0 {
		logger.Info("no jobs to output — skipping write")
		return nil
	}

	// Ensure output directory exists
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		logger.Error("failed to create output directory",
			slog.String("dir", cfg.Dir),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("create output dir %s: %w", cfg.Dir, err)
	}

	scrapedAt := time.Now().UTC().Format(time.RFC3339)

	// Convert to output records
	records := make([]OutputRecord, 0, len(jobs))
	for _, j := range jobs {
		records = append(records, OutputRecord{
			ID:         j.GetId(),
			Title:      j.GetTitle(),
			Department: j.GetDepartment(),
			Location:   j.GetLocation(),
			URL:        j.GetUrl(),
			Date:       j.GetDate(),
			ScrapedAt:  scrapedAt,
		})
	}

	switch cfg.Format {
	case "csv":
		return appendCSV(records, cfg.Dir, logger)
	default:
		return appendJSON(records, cfg.Dir, logger)
	}
}

// appendJSON reads existing JSON records, appends new ones, and writes back.
func appendJSON(records []OutputRecord, dir string, logger *slog.Logger) error {
	filePath := filepath.Join(dir, "data.json")

	// Read existing records if file exists
	existing := make([]OutputRecord, 0)
	if data, err := os.ReadFile(filePath); err == nil && len(data) > 0 {
		if jsonErr := json.Unmarshal(data, &existing); jsonErr != nil {
			logger.Warn("existing JSON is malformed — starting fresh",
				slog.String("file", filePath),
				slog.String("error", jsonErr.Error()),
			)
		}
	}

	// Append new records
	combined := append(existing, records...)

	// Write back
	data, err := json.MarshalIndent(combined, "", "  ")
	if err != nil {
		logger.Error("failed to marshal JSON",
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("marshal JSON: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		logger.Error("failed to write JSON output",
			slog.String("file", filePath),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("write JSON %s: %w", filePath, err)
	}

	logger.Info("JSON output written",
		slog.String("file", filePath),
		slog.Int("new_records", len(records)),
		slog.Int("total_records", len(combined)),
	)
	return nil
}

// appendCSV opens the CSV file in append mode, writing headers only if the file is new.
func appendCSV(records []OutputRecord, dir string, logger *slog.Logger) error {
	filePath := filepath.Join(dir, "data.csv")

	// Check if file exists to decide whether to write headers
	isNew := false
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		isNew = true
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("failed to open CSV for append",
			slog.String("file", filePath),
			slog.String("error", err.Error()),
		)
		return fmt.Errorf("open CSV %s: %w", filePath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header row only for new files
	if isNew {
		header := []string{"id", "title", "department", "location", "url", "date", "scraped_at"}
		if err := writer.Write(header); err != nil {
			return fmt.Errorf("write CSV header: %w", err)
		}
	}

	for _, r := range records {
		row := []string{
			r.ID,
			r.Title,
			r.Department,
			r.Location,
			r.URL,
			r.Date,
			r.ScrapedAt,
		}
		if err := writer.Write(row); err != nil {
			logger.Error("failed to write CSV row",
				slog.String("id", r.ID),
				slog.String("error", err.Error()),
			)
			// Continue writing remaining rows
			continue
		}
	}

	logger.Info("CSV output written",
		slog.String("file", filePath),
		slog.Int("records_appended", len(records)),
	)
	return nil
}
