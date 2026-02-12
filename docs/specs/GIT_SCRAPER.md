# Serverless Git Scraper Specification (SQLite)

## Overview
This document outlines the "Ship the Database" architecture. Instead of an API, the backend generates a static SQLite file which is downloaded by client applications.

## Architecture

### 1. Data Source
- **Target**: `https://recruitment.nic.in/index_new.php`
- **Method**: Content fetching via `colly`.

### 2. Data Model (`jobs.db`)
- **Table**: `jobs`
  - `id` (TEXT PK): Unique identifier.
  - `title` (TEXT): Job title.
  - `department` (TEXT): Department name.
  - `location` (TEXT): Job location.
  - `posted_date` (INTEGER): Unix timestamp.
  - `url` (TEXT): Link to posting.
- **Optimization**: `VACUUM` and `PRAGMA journal_mode = DELETE` are run before distribution to ensure a single, compact file.

### 3. Synchronization (`metadata.json`)
Generated alongside `jobs.db` to allow clients to check for updates without downloading the full DB.
```json
{
  "last_updated": 1700000000,
  "checksum": "sha256-hash-of-jobs.db",
  "job_count": 42
}
```

### 4. Client Implementation
- **Strategy**: "Hot-Swap"
- **Logic**:
  1. Fetch `metadata.json`.
  2. Compare checksum with local state.
  3. If different, download `jobs.db` to a temp path.
  4. Verify SHA256 matches.
  5. Close existing DB connection.
  6. Rename temp file to `jobs.db`.
  7. Re-open DB.

### 5. Automation (GitHub Actions)
- **Schedule**: Every 6 hours (`0 */6 * * *`).
- **Workflow**:
  1. Run Scraper.
  2. Generate `jobs.db` + `metadata.json`.
  3. Commit & Push.
