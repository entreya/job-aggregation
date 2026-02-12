package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

// Job represents the job structure for the database.
type Job struct {
	ID         string
	Title      string
	Department string
	Location   string
	PostedDate int64 // Unix timestamp
	URL        string
}

// DB wraps the sql.DB connection.
type DB struct {
	conn *sql.DB
}

// InitDB initializes the SQLite database and creates the jobs table if it doesn't exist.
func InitDB(filepath string) (*DB, error) {
	db, err := sql.Open("sqlite", filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	// Optimize for single-writer
	// Use WAL for better concurrency if needed, but here simple is fine.
	// However, user requested portability, so let's stick to standard journal for now unless performance dictates otherwise.
	// Actually, user requested "PRAGMA journal_mode = DELETE" at the end. We'll set that in Optimize().

	createTableSQL := `
	CREATE TABLE IF NOT EXISTS jobs (
		id TEXT PRIMARY KEY,
		title TEXT,
		department TEXT,
		location TEXT,
		posted_date INTEGER,
		url TEXT
	);
	`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &DB{conn: db}, nil
}

// UpsertJob insterts a new job or updates an existing one on conflict.
// We use INSERT OR REPLACE which is standard for SQLite upserts on PK.
func (d *DB) UpsertJob(job Job) error {
	query := `
	INSERT OR REPLACE INTO jobs (id, title, department, location, posted_date, url)
	VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := d.conn.Exec(query, job.ID, job.Title, job.Department, job.Location, job.PostedDate, job.URL)
	if err != nil {
		return fmt.Errorf("failed to upsert job %s: %w", job.ID, err)
	}
	return nil
}

// Optimize runs VACUUM and sets journal_mode to DELETE to ensure the file is as small and portable as possible.
// It also closes the connection.
func (d *DB) OptimizeAndClose() error {
	log.Println("Optimizing database...")

	// VACUUM to reclaim space
	_, err := d.conn.Exec("VACUUM;")
	if err != nil {
		return fmt.Errorf("failed to VACUUM: %w", err)
	}

	// Set journal_mode to DELETE to remove -wal and -shm files if any, ensuring a single file.
	_, err = d.conn.Exec("PRAGMA journal_mode = DELETE;")
	if err != nil {
		return fmt.Errorf("failed to set journal_mode: %w", err)
	}

	return d.conn.Close()
}
