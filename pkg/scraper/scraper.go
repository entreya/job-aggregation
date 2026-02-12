package scraper

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/entreya/job-aggregation/pkg/models"
	"github.com/gocolly/colly/v2"
)

// Scraper handles the job scraping logic.
type Scraper struct {
	collector *colly.Collector
	TargetURL string
}

func NewScraper(targetURL string) *Scraper {
	c := colly.NewCollector(
		colly.AllowedDomains("recruitment.nic.in"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		colly.AllowURLRevisit(),
	)

	// Only apply custom transport if NOT using a local file
	if !strings.HasPrefix(targetURL, "file://") {
		// Custom Transport to handle slow gov sites and potential SSL issues
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		}
		c.WithTransport(transport)
		c.SetRequestTimeout(120 * time.Second)
	}

	return &Scraper{
		collector: c,
		TargetURL: targetURL,
	}
}

// Scrape fetches job postings from the recruitment site.
func (s *Scraper) Scrape() (*models.JobList, error) {
	jobList := &models.JobList{
		LastUpdated: time.Now().Unix(),
		Jobs:        make([]*models.JobPosting, 0),
	}

	// Register callbacks
	s.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// ... (rest of logic same) ...
		link := e.Attr("href")
		text := strings.TrimSpace(e.Text)

		if link == "" || text == "" {
			return
		}

		absoluteURL := e.Request.AbsoluteURL(link)

		job := &models.JobPosting{
			Id:         generateID(absoluteURL),
			Title:      text,
			Department: "NIC",
			Location:   "All India",
			Url:        absoluteURL,
			Date:       time.Now().Format("2006-01-02"),
		}

		jobList.Jobs = append(jobList.Jobs, job)
	})

	s.collector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	fmt.Println("Visiting:", s.TargetURL)
	err := s.collector.Visit(s.TargetURL)
	if err != nil {
		return nil, err
	}

	s.collector.Wait()

	return jobList, nil
}

func generateID(url string) string {
	// Simple hash of the URL to ensure uniqueness and stability
	// For a real system we might use md5 or sha256
	// avoiding external deps for simplicity if possible, but fine to just use string for now
	return url
}
