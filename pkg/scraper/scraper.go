package scraper

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/chromedp"
	"github.com/entreya/job-aggregation/pkg/models"
)

// Scraper handles the job scraping logic.
type Scraper struct {
	TargetURL string
}

func NewScraper(targetURL string) *Scraper {
	return &Scraper{
		TargetURL: targetURL,
	}
}

// Scrape fetches job postings from the recruitment site using Chromedp.
func (s *Scraper) Scrape() (*models.JobList, error) {
	// Create context
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Set a timeout for the entire operation
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	log.Printf("Visiting %s using Chromedp...", s.TargetURL)

	var htmlContent string
	err := chromedp.Run(ctx,
		chromedp.Navigate(s.TargetURL),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		chromedp.OuterHTML("html", &htmlContent),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %v", s.TargetURL, err)
	}

	log.Println("Page loaded. Parsing HTML...")

	// Parse HTML with GoQuery to extract jobs
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	jobList := &models.JobList{
		LastUpdated: time.Now().Unix(),
		Jobs:        make([]*models.JobPosting, 0),
	}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}
		text := strings.TrimSpace(s.Text())

		if link == "" || text == "" {
			return
		}

		// Handle relative URLs
		if !strings.HasPrefix(link, "http") {
			// Basic relative URL handling (assuming base is recruitment.nic.in/index_new.php or similar)
			// Ideally we resolve against base URL properly, but for this specific site:
			// If it starts with /, it's root relative. If not, it's relative to current path.
			// The original colly code used e.Request.AbsoluteURL(link).
			// We can reconstruct it manually or use a helper.
			// Given extraction is from "recruitment.nic.in", we can prefix.
			baseURL := "https://recruitment.nic.in/"
			if strings.HasPrefix(link, "/") {
				link = strings.TrimSuffix(baseURL, "/") + link
			} else {
				// Simple join if not absolute
				link = baseURL + link
			}
		}

		job := &models.JobPosting{
			Id:         generateID(link),
			Title:      text,
			Department: "NIC",
			Location:   "All India",
			Url:        link,
			Date:       time.Now().Format("2006-01-02"),
		}

		jobList.Jobs = append(jobList.Jobs, job)
	})

	log.Printf("Found %d potential job links.", len(jobList.Jobs))
	return jobList, nil
}

func generateID(url string) string {
	// Simple ID generation
	return url
}
