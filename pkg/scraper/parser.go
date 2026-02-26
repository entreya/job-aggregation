package scraper

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"regexp"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"github.com/entreya/job-aggregation/pkg/models"
)

const (
	// baseURL is the root URL for resolving relative links.
	baseURL = "https://recruitment.nic.in/"
)

// controlCharRegex matches non-printable control characters (excluding \n, \r, \t).
var controlCharRegex = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)

// ParseJobs extracts job postings from raw HTML content.
// It parses all <a href> links within the page, resolves relative URLs,
// sanitizes text, and generates stable IDs from URL hashes.
//
// Rows with empty title or URL are skipped and logged.
func ParseJobs(htmlContent string, logger *slog.Logger) ([]*models.JobPosting, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, err
	}

	// Pre-allocate with a reasonable capacity
	jobs := make([]*models.JobPosting, 0, 64)
	skipped := 0

	doc.Find("a[href]").Each(func(i int, sel *goquery.Selection) {
		link, exists := sel.Attr("href")
		if !exists {
			return
		}

		text := SanitizeString(sel.Text())
		link = strings.TrimSpace(link)

		// Skip entries with empty title or link
		if link == "" || text == "" {
			skipped++
			logger.Debug("skipped row: empty title or URL",
				slog.Int("row_index", i),
				slog.String("raw_link", link),
				slog.String("raw_text", text),
			)
			return
		}

		// Resolve relative URLs
		link = resolveURL(link)

		job := &models.JobPosting{
			Id:         GenerateID(link),
			Title:      text,
			Department: "NIC",
			Location:   "All India",
			Url:        link,
		}

		jobs = append(jobs, job)
	})

	if skipped > 0 {
		logger.Info("rows skipped during parsing",
			slog.Int("skipped_count", skipped),
			slog.Int("parsed_count", len(jobs)),
		)
	}

	return jobs, nil
}

// SanitizeString cleans a scraped string by:
//   - Trimming leading/trailing whitespace
//   - Removing non-printable control characters
//   - Collapsing multiple whitespace into single spaces
func SanitizeString(s string) string {
	// Remove control characters
	s = controlCharRegex.ReplaceAllString(s, "")

	// Collapse whitespace (spaces, tabs, newlines) into single spaces
	fields := strings.FieldsFunc(s, unicode.IsSpace)
	s = strings.Join(fields, " ")

	return strings.TrimSpace(s)
}

// GenerateID creates a stable, unique ID by SHA256-hashing the URL.
// This ensures IDs are safe for use as DB primary keys regardless of
// URL special characters.
func GenerateID(rawURL string) string {
	hash := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(hash[:16]) // First 16 bytes = 32 hex chars (sufficient uniqueness)
}

// resolveURL converts relative URLs to absolute URLs using baseURL.
func resolveURL(link string) string {
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}

	base := strings.TrimSuffix(baseURL, "/")
	if strings.HasPrefix(link, "/") {
		return base + link
	}

	return baseURL + link
}
