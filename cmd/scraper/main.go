package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/entreya/job-aggregation/pkg/db"
	"github.com/entreya/job-aggregation/pkg/logger"
	"github.com/entreya/job-aggregation/pkg/models"
	"github.com/entreya/job-aggregation/pkg/proxy"
	"github.com/entreya/job-aggregation/pkg/scraper"
)

const (
	jobsJSONPath = "data/jobs.json"
)

// Metadata represents the sync metadata for client-side update checks.
type Metadata struct {
	LastUpdated int64  `json:"last_updated"`
	Checksum    string `json:"checksum"`
	JobCount    int    `json:"job_count"`
}

func main() {
	// ─── 1. Initialize structured logger ───────────────────────────────
	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}
	log := logger.Init(env)
	log.Info("starting job scraper",
		slog.String("env", env),
	)

	// ─── 2. Initialize proxy rotator ───────────────────────────────────
	proxyURLs := os.Getenv("PROXY_URLS")
	proxyStrategy := os.Getenv("PROXY_STRATEGY")
	if proxyStrategy == "" {
		proxyStrategy = "round-robin"
	}

	rotator, err := proxy.NewRotator(proxyURLs, proxyStrategy, log)
	if err != nil {
		log.Error("failed to initialize proxy rotator",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	// ─── 3. Initialize database ────────────────────────────────────────
	dbPath := "jobs.db"
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Error("failed to initialize database",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	// ─── 4. Configure and run scraper ──────────────────────────────────
	chromePath := os.Getenv("CHROME_PATH")
	s := scraper.NewScraper(scraper.Config{
		TargetURL:  "https://recruitment.nic.in/index_new.php",
		Rotator:    rotator,
		RetryCfg:   scraper.DefaultRetryConfig(),
		Logger:     log,
		ChromePath: chromePath,
		Timeout:    90 * time.Second,
	})

	jobsList, err := s.Scrape()
	if err != nil {
		log.Error("scrape failed",
			slog.String("error", err.Error()),
		)
		// Close DB before exiting
		if closeErr := database.OptimizeAndClose(); closeErr != nil {
			log.Error("failed to close DB", slog.String("error", closeErr.Error()))
		}
		os.Exit(1)
	}

	log.Info("scrape successful",
		slog.Int("jobs_count", len(jobsList.Jobs)),
	)

	// ─── 5. Insert jobs into SQLite (upsert) ───────────────────────────
	for _, j := range jobsList.Jobs {
		job := db.Job{
			ID:         j.Id,
			Title:      j.Title,
			Department: j.Department,
			Location:   j.Location,
			PostedDate: time.Now().Unix(),
			URL:        j.Url,
		}
		if err := database.UpsertJob(job); err != nil {
			log.Warn("failed to upsert job",
				slog.String("job_id", j.Id),
				slog.String("error", err.Error()),
			)
		}
	}

	// ─── 6. Optimize and close DB ──────────────────────────────────────
	if err := database.OptimizeAndClose(); err != nil {
		log.Error("failed to optimize and close DB",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	// ─── 7. Generate metadata ──────────────────────────────────────────
	if err := generateMetadata(dbPath, len(jobsList.Jobs)); err != nil {
		log.Error("failed to generate metadata",
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	// ─── 8. Export to JSON (legacy format for Flutter client) ───────────
	if err := exportToJSON(jobsList); err != nil {
		log.Warn("error exporting to JSON (non-fatal)",
			slog.String("error", err.Error()),
		)
	}

	// ─── 9. Append to output file (new format with timestamps) ─────────
	outputCfg := scraper.DefaultOutputConfig()
	if err := scraper.AppendResults(jobsList.Jobs, outputCfg, log); err != nil {
		log.Warn("error appending output (non-fatal)",
			slog.String("error", err.Error()),
		)
	}

	log.Info("pipeline complete",
		slog.String("db", dbPath),
		slog.String("metadata", "metadata.json"),
		slog.String("json_export", jobsJSONPath),
	)
}

func generateMetadata(dbPath string, count int) error {
	file, err := os.Open(dbPath)
	if err != nil {
		return fmt.Errorf("failed to open db for hashing: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate hash: %w", err)
	}
	checksum := hex.EncodeToString(hash.Sum(nil))

	meta := Metadata{
		LastUpdated: time.Now().Unix(),
		Checksum:    checksum,
		JobCount:    count,
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile("metadata.json", data, 0644)
}

func exportToJSON(jobList *models.JobList) error {
	dir := filepath.Dir(jobsJSONPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(jobsJSONPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(jobList)
}
