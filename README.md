# Job Aggregation Scraper (SQLite Edition)

A serverless "Ship the Database" architecture. This scraper fetches job postings from `recruitment.nic.in`, stores them in a highly optimized SQLite database (`jobs.db`), and serves them via GitHub Pages.

## Architecture
- **Backend**: Go + `modernc.org/sqlite` (Pure Go).
- **Automation**: GitHub Actions (Runs every 6 hours).
- **Client**: Flutter app downloading the raw `jobs.db` file.

## Project Structure
- `cmd/scraper`: Main application entry point.
- `pkg/scraper`: Scraping logic (`colly`).
- `pkg/db`: SQLite database management.
- `mobile`: Flutter mobile application.
- `.github/workflows`: Automation configuration.

## Getting Started

### Backend (Go)
1.  **Run Scraper**:
    ```bash
    go run cmd/scraper/main.go
    ```
    This generates `jobs.db` and `metadata.json`.

### Mobile App (Flutter)
The mobile app is located in the `mobile/` directory.

1.  **Dependencies**:
    ```bash
    cd mobile
    flutter pub add flutter_riverpod dio sqflite path_provider path crypto url_launcher google_fonts
    ```

2.  **Run App**:
    ```bash
    flutter run
    ```
    
    *Note: The app is configured to fetch the database from the `main` branch of this repository. Ensure you have pushed the generated `jobs.db` and `metadata.json` to GitHub.*

## Mobile Architecture
- **State Management**: Riverpod (`jobsProvider`, `searchQueryProvider`).
- **Database**: `sqflite` with `DatabaseManager` for hot-swapping.
- **UI**: Material 3 with Search and Pull-to-Refresh.

## Deployment
1.  Push this repository to GitHub.
2.  Enable GitHub Actions.
3.  The workflow `Job Scraper` will run automatically every 6 hours and commit the updated `jobs.db`.

## License
MIT
