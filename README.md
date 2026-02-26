# Job Aggregation Scraper (SQLite Edition)

A serverless "Ship the Database" architecture. This scraper fetches job postings from `recruitment.nic.in`, stores them in a highly optimized SQLite database (`jobs.db`), and serves them via GitHub Pages.

## Architecture
- **Backend**: Go + `chromedp` (headless Chrome/Chromium) + `goquery` for parsing.
- **Proxy**: Indian proxy rotation with round-robin/random strategies.
- **Anti-Bot**: User-Agent rotation, human delay simulation, retry with exponential backoff.
- **Database**: `modernc.org/sqlite` (Pure Go, zero-CGO).
- **Logging**: Structured JSON logging via `slog` (Go 1.21+ stdlib).
- **Automation**: GitHub Actions on a self-hosted runner (Oracle Cloud Mumbai) + proxy-based fallback.
- **Client**: Flutter app downloading the raw `jobs.db` file.

## Project Structure
```
cmd/scraper/        Main application entry point
pkg/scraper/        Scraping logic (chromedp + goquery + retry + output)
  ├── scraper.go    Core scraper with proxy/retry integration
  ├── client.go     Anti-bot browser client (UA rotation, chromedp options)
  ├── retry.go      Exponential backoff retry logic
  ├── parser.go     HTML parser (goquery, sanitization, ID generation)
  └── output.go     Data output (JSON/CSV append with timestamps)
pkg/proxy/          Proxy rotation (round-robin, random)
pkg/logger/         Structured logging (slog, JSON/text handler)
pkg/db/             SQLite database management
pkg/models/         Protobuf-generated data models
mobile/             Flutter mobile application
.github/workflows/  Automation (scraper, VPS scraper, APK release)
docs/               Guides and debug research
```

## Setup & Running

### Prerequisites
- Go 1.23+
- Chrome or Chromium installed (auto-detected, or set `CHROME_PATH`)

### Run Scraper Locally
```bash
# Development mode (human-readable logs, no proxy)
ENV=development go run cmd/scraper/main.go

# With proxies
PROXY_URLS="http://user:pass@proxy1:8080,http://user:pass@proxy2:8080" \
PROXY_STRATEGY=round-robin \
ENV=development \
go run cmd/scraper/main.go
```

This generates: `jobs.db`, `metadata.json`, `data/jobs.json`, and `output/data.json`.

### Environment Variables

| Variable         | Default        | Description                                                                 |
|-----------------|----------------|-----------------------------------------------------------------------------|
| `CHROME_PATH`    | Auto-detected  | Path to Chrome/Chromium binary. Set to `/usr/bin/chromium-browser` on ARM.  |
| `PROXY_URLS`     | *(empty)*      | Comma-separated proxy URLs: `http://user:pass@host:port`                   |
| `PROXY_STRATEGY` | `round-robin`  | Proxy selection: `round-robin` or `random`                                 |
| `ENV`            | `development`  | `production` = JSON logs, `development` = human-readable logs              |

### Setting Up GitHub Secrets
1. Go to your repo → **Settings** → **Secrets and variables** → **Actions**
2. Add the following secrets:
   - `PROXY_URLS` — comma-separated proxy URLs (e.g., `http://user:pass@proxy1:8080`)
3. The `scraper.yml` workflow automatically injects these as environment variables.

### Running Tests
```bash
# Run all unit tests
go test ./...

# Run with verbose output
go test ./... -v

# Run with coverage report
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out

# Run integration tests only
go test ./... -tags=integration -v

# Run a specific package
go test ./scraper/... -v

# Run with race detector
go test -race ./...
```

## Mobile App (Flutter)
Located in the `mobile/` directory.

```bash
cd mobile
flutter pub get
flutter run
```

*Note: The app fetches the database from the `main` branch of this repository. Ensure `jobs.db` and `metadata.json` are committed.*

## Mobile Architecture
- **State Management**: Riverpod (`jobsProvider`, `searchQueryProvider`).
- **Database**: `sqflite` with `DatabaseManager` for hot-swapping.
- **UI**: Material 3 with Search and Pull-to-Refresh.

## Deployment

### Recommended: Oracle Cloud Free VPS (Mumbai)
1. Set up an **Oracle Cloud Always Free** ARM VM in Mumbai ($0/mo forever).
2. Install Go + Chromium + GitHub Actions self-hosted runner.
3. See `docs/guides/self_hosted_runner_setup.md` for full instructions.
4. The workflow `scraper-vps.yml` runs every 6 hours automatically.

### Backup: GitHub-hosted runner + Proxies
The `scraper.yml` workflow uses proxy rotation on GitHub's cloud runners. Schedule runs every 6 hours and requires `PROXY_URLS` secret for geo-restricted access.

## Dependencies
- [chromedp](https://github.com/chromedp/chromedp) — Headless Chrome/Chromium automation
- [goquery](https://github.com/PuerkitoBio/goquery) — HTML parsing
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — Pure Go SQLite
- [protobuf](https://google.golang.org/protobuf) — Data model serialization
- [slog](https://pkg.go.dev/log/slog) — Structured logging (Go stdlib)

## License
MIT
