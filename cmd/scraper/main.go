package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/entreya/job-aggregation/pkg/db"
	"github.com/entreya/job-aggregation/pkg/models"
	"github.com/entreya/job-aggregation/pkg/scraper"
)

const (
	jobsJSONPath = "data/jobs.json"
)

type Metadata struct {
	LastUpdated int64  `json:"last_updated"`
	Checksum    string `json:"checksum"`
	JobCount    int    `json:"job_count"`
}

func main() {
	log.Println("Starting job scraper (Chromedp Mode)...")

	// Initialize DB
	dbPath := "jobs.db"
	database, err := db.InitDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Scrape
	s := scraper.NewScraper("https://recruitment.nic.in/index_new.php")
	jobsList, err := s.Scrape()
	if err != nil {
		log.Fatalf("Failed to scrape jobs: %v", err)
	}

	log.Printf("Scraped %d jobs. Inserting into DB...", len(jobsList.Jobs))

	// Insert into SQLite and upsert
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
			log.Printf("Failed to upsert job %s: %v", j.Id, err)
		}
	}

	// Optimize and Close
	if err := database.OptimizeAndClose(); err != nil {
		log.Fatalf("Failed to optimize and close DB: %v", err)
	}

	// Generate Metadata
	if err := generateMetadata(dbPath, len(jobsList.Jobs)); err != nil {
		log.Fatalf("Failed to generate metadata: %v", err)
	}

	// Export to JSON
	if err := exportToJSON(jobsList); err != nil {
		log.Printf("Error exporting to JSON: %v", err)
	}

	log.Println("Successfully updated jobs.db, metadata.json, and data/jobs.json")
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
	// Ensure directory exists
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
