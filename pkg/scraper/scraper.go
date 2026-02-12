package scraper

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/entreya/job-aggregation/pkg/models"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/proxy"
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

	// Only apply proxy rotation if NOT using a local file
	if !strings.HasPrefix(targetURL, "file://") {
		log.Println("Fetching proxy list...")
		// Fetch proxies from a free list (HTTP/S, Anonymous) - Force SSL=yes for HTTPS target
		resp, err := http.Get("https://api.proxyscrape.com/v2/?request=displayproxies&protocol=http&timeout=5000&country=all&ssl=yes&anonymity=all")
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			proxyList := strings.Split(string(body), "\n")

			var validProxies []string
			for _, p := range proxyList {
				p = strings.TrimSpace(p)
				if p != "" {
					validProxies = append(validProxies, "http://"+p)
				}
			}

			if len(validProxies) > 0 {
				log.Printf("Found %d HTTPS proxies. Setting up rotation.", len(validProxies))
				rp, err := proxy.RoundRobinProxySwitcher(validProxies...)
				if err != nil {
					log.Printf("Failed to set proxy switcher: %v", err)
				} else {
					c.SetProxyFunc(rp)
				}
			} else {
				log.Println("No proxies found. Falling back to direct connection.")
			}
		} else {
			log.Printf("Failed to fetch proxy list: %v", err)
		}

		// Keep the custom transport settings for robustness
		transport := &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second, // Fast fail for proxies
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives:     true,
		}
		c.WithTransport(transport)
		c.SetRequestTimeout(20 * time.Second) // Fast timeout for individual requests
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

	// Attempt counter to prevent infinite loops (simple context hack or just limit retries conceptually)
	// For simplicity, we'll just trust Colly's queue or rely on the fact that we have many proxies.

	// Register callbacks
	s.collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// ...
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
		log.Printf("Request URL: %s failed. Proxy: %s. Error: %v. Retrying...", r.Request.URL, r.Request.ProxyURL, err)
		// Retry with a different proxy (RoundRobin will pick next one)
		// We add a short delay to avoid rapid-fire looping if all fail
		time.Sleep(1 * time.Second)
		r.Request.Retry()
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
