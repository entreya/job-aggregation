package scraper

import (
	"fmt"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/entreya/job-aggregation/pkg/models"
)

// Scraper handles the job scraping logic.
type Scraper struct {
	collector *colly.Collector
}

// NewScraper creates a new Scraper instance.
func NewScraper() *Scraper {
	c := colly.NewCollector(
		colly.AllowedDomains("recruitment.nic.in"),
		colly.UserAgent("Mozilla/5.0 (compatible; JobScraper/1.0; +https://github.com/entreya/job-aggregation)"),
	)

	return &Scraper{
		collector: c,
	}
}

// Scrape fetches job postings from the recruitment site.
func (s *Scraper) Scrape() (*models.JobList, error) {
	jobList := &models.JobList{
		LastUpdated: time.Now().Unix(),
		Jobs:        make([]*models.JobPosting, 0),
	}

	s.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		text := strings.TrimSpace(e.Text)

		// Basic validation to ensure it's a relevant link
		if link == "" || text == "" {
			return
		}
		
		// Resolve relative URLs
		absoluteURL := e.Request.AbsoluteURL(link)

        // Filter out irrelevant links if possible (e.g. only pdfs or specific keywords)
        // For now, we take all links that look like job postings or notifications
        // We can refine this logic based on actual needs.
        
		job := &models.JobPosting{
			Id:         fmt.Sprintf("%s-%d", "nic", time.Now().UnixNano()), // Generate a temp ID, ideally should be based on URL hash
			Title:      text,
			Department: "NIC", // Default
			Location:   "All India", // Default
			Url:        absoluteURL,
			Date:       time.Now().Format("2006-01-02"), // Current date as proxy
		}
        
        // Refine ID generation to be deterministic based on URL
        job.Id = generateID(absoluteURL)

		jobList.Jobs = append(jobList.Jobs, job)
	})

	s.collector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	err := s.collector.Visit("https://recruitment.nic.in/index_new.php")
	if err != nil {
		return nil, err
	}
    
    // Wait for scraping to finish
	s.collector.Wait()

	return jobList, nil
}

func generateID(url string) string {
    // Simple hash of the URL to ensure uniqueness and stability
    // For a real system we might use md5 or sha256
    // avoiding external deps for simplicity if possible, but fine to just use string for now
    return url 
}
