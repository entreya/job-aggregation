package scraper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/entreya/job-aggregation/pkg/models"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
)

const (
	MaxRetries = 3
)

// Scraper handles the job scraping logic.
type Scraper struct {
	collector *colly.Collector
	TargetURL string
}

func NewScraper(targetURL string) *Scraper {
	c := colly.NewCollector(
		colly.AllowedDomains("recruitment.nic.in"),
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		colly.AllowURLRevisit(),
	)

	// Set headers to mimic real browser
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
		r.Headers.Set("Referer", "https://www.google.com/")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	})

	// Only apply proxy rotation if NOT using a local file
	if !strings.HasPrefix(targetURL, "file://") {
		// 1. Check for manual proxy (e.g. from GitHub Secrets)
		envProxy := os.Getenv("HTTP_PROXY")
		if envProxy != "" {
			log.Printf("Using configured HTTP_PROXY: %s", envProxy)
			if err := c.SetProxy(envProxy); err != nil {
				log.Printf("Failed to set HTTP_PROXY: %v", err)
			}
		} else {
			// 2. Fetch proxies (Prioritize India for gov sites)
			log.Println("Fetching proxy list (India)...")
			// Try fetching Indian proxies first
			proxies := fetchProxies("IN")
			if len(proxies) == 0 {
				log.Println("No Indian proxies found, trying all countries...")
				proxies = fetchProxies("all")
			}

			if len(proxies) > 0 {
				log.Printf("Found %d HTTPS proxies. Setting up rotation.", len(proxies))
				rp, err := proxy.RoundRobinProxySwitcher(proxies...)
				if err != nil {
					log.Printf("Failed to set proxy switcher: %v", err)
				} else {
					// CRITICAL: Assign the proxy switcher directly to the Transport via Colly's mechanism
					c.SetProxyFunc(rp)
				}
			} else {
				log.Println("No proxies found. Falling back to direct connection.")
			}
		}

		// Use standard transport (Colly's default) to ensure SetProxyFunc works,
		// OR configured strict transport if proxies are failing.
		// For now, let's rely on Colly's default handling + our Headers.
		c.SetRequestTimeout(30 * time.Second)
	}

	return &Scraper{
		collector: c,
		TargetURL: targetURL,
	}
}

func fetchProxies(country string) []string {
	apiURL := fmt.Sprintf("https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=5000&country=%s&ssl=yes&anonymity=all", country)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Printf("Failed to fetch proxies: %v", err)
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	lines := strings.Split(string(body), "\n")
	var valid []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			valid = append(valid, "http://"+line)
		}
	}
	return valid
}

// Scrape fetches job postings from the recruitment site.
func (s *Scraper) Scrape() (*models.JobList, error) {
	jobList := &models.JobList{
		LastUpdated: time.Now().Unix(),
		Jobs:        make([]*models.JobPosting, 0),
	}

	// Register callbacks
	s.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
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

	// Retry logic
	s.collector.OnError(func(r *colly.Response, err error) {
		log.Printf("Request URL: %s failed. Proxy: %s. Error: %v", r.Request.URL, r.Request.ProxyURL, err)

		if r.Request.Ctx.GetAny("retries") == nil {
			r.Request.Ctx.Put("retries", 0)
		}
		retries := r.Request.Ctx.GetAny("retries").(int)

		if retries < MaxRetries {
			log.Printf("Retrying... (%d/%d)", retries+1, MaxRetries)
			r.Request.Ctx.Put("retries", retries+1)
			time.Sleep(2 * time.Second) // Backoff
			r.Request.Retry()
		} else {
			log.Println("Max retries reached. Moving on.")
		}
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
