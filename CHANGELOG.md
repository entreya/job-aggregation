# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- `[FEAT]` Indian proxy rotation (`pkg/proxy/rotator.go`) — random and round-robin strategies, thread-safe via `sync/atomic`.
- `[FEAT]` Anti-bot browser client (`pkg/scraper/client.go`) — 10 real User-Agent strings, human delay simulation, chromedp allocator builder.
- `[FEAT]` Retry with exponential backoff (`pkg/scraper/retry.go`) — retries on network/timeout/chromedp errors, aborts on non-retryable.
- `[FEAT]` HTML parser extracted (`pkg/scraper/parser.go`) — SHA256-based stable IDs, string sanitization, relative URL resolution.
- `[FEAT]` Data output module (`pkg/scraper/output.go`) — JSON/CSV append (never overwrite), `scraped_at` UTC timestamps.
- `[FEAT]` Structured logging (`pkg/logger/logger.go`) — `slog` JSON handler (production) / text handler (development).
- `[FEAT]` `PROXY_URLS`, `PROXY_STRATEGY`, `ENV` environment variables for configuration.
- `[TEST]` 37 unit + integration tests across `pkg/proxy` and `pkg/scraper` packages.
- `[TEST]` HTML test fixture (`testdata/sample.html`) for parser testing.
- `[TEST]` Integration test with `httptest.Server` for full pipeline validation.
- `[FEAT]` Full light/dark mode with 30+ color tokens and theme-aware extension.
- `[FEAT]` 6 detailed sample job listings (SSC, Railway, UPSC, IBPS, Army, NTA).
- `[FEAT]` Staggered slide-in animations and gradient stat cards.
- `[FEAT]` Oracle Cloud Always Free VPS support (ARM64/aarch64, Mumbai).
- `[FEAT]` `CHROME_PATH` environment variable for flexible browser path detection (Chrome vs Chromium).
- `[FEAT]` Auto-detect browser step in `scraper-vps.yml` — works with both Chrome (x64) and Chromium (ARM).
- `[FEAT]` `disable-dev-shm-usage` flag in Chromedp for low-memory VPS stability.
- `[FEAT]` Human-like random delay before HTML extraction to reduce detection risk.
- `[FEAT]` Failure notification (GitHub Issue creation) in VPS workflow.
- `[DOCS]` Comprehensive scraping strategy research (`docs/debug/SCRAPING_STRATEGY.md`).
- `[DOCS]` Created `CHANGELOG.md`.

### Changed
- `[REFACTOR]` Rewrote `self_hosted_runner_setup.md` for Oracle Cloud ARM64 (was DigitalOcean x64).
- `[REFACTOR]` Rewrote `scraper-vps.yml` with browser auto-detection and ARM support.
- `[DOCS]` Updated `README.md` — corrected `colly` → `chromedp`, added env vars, deployment guide.

### Deprecated
- `[DEPRECATE]` `scraper.yml` schedule trigger disabled — replaced by `scraper-vps.yml`.

## [0.1.0] — 2026-02-12

### Added
- Initial Go scraper with `chromedp` + `goquery`.
- SQLite backend via `modernc.org/sqlite`.
- Flutter mobile app with Riverpod state management.
- GitHub Actions workflows for automated scraping and APK releases.
- Protobuf data model definitions.
