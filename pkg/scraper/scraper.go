package scraper

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/entreya/job-aggregation/pkg/models"
	"github.com/entreya/job-aggregation/pkg/proxy"
)

// Scraper handles the job scraping logic with proxy rotation and retry support.
type Scraper struct {
	TargetURL  string
	Rotator    *proxy.ProxyRotator
	RetryCfg   RetryConfig
	Logger     *slog.Logger
	ChromePath string // Custom browser path (empty = auto-detect)
	Timeout    time.Duration
}

// Config holds initialization parameters for the Scraper.
type Config struct {
	TargetURL  string
	Rotator    *proxy.ProxyRotator
	RetryCfg   RetryConfig
	Logger     *slog.Logger
	ChromePath string
	Timeout    time.Duration
}

// NewScraper creates a Scraper with all dependencies injected.
func NewScraper(cfg Config) *Scraper {
	if cfg.Timeout == 0 {
		cfg.Timeout = 60 * time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &Scraper{
		TargetURL:  cfg.TargetURL,
		Rotator:    cfg.Rotator,
		RetryCfg:   cfg.RetryCfg,
		Logger:     cfg.Logger,
		ChromePath: cfg.ChromePath,
		Timeout:    cfg.Timeout,
	}
}

// Scrape fetches job postings from the target URL using chromedp with
// proxy rotation, retry logic, and anti-bot countermeasures.
func (s *Scraper) Scrape() (*models.JobList, error) {
	var htmlContent string
	var usedProxy string

	// Wrap the entire chromedp operation in the retry loop
	err := WithRetry(s.RetryCfg, s.TargetURL, "", s.Logger, func(attempt int) error {
		// Select proxy for this attempt (rotates on each retry)
		proxyURL := ""
		if s.Rotator != nil {
			proxyURL = s.Rotator.ProxyServerAddr()
		}
		usedProxy = proxyURL

		s.Logger.Info("scrape attempt starting",
			slog.String("url", s.TargetURL),
			slog.String("proxy_used", proxyURL),
			slog.Int("attempt", attempt+1),
		)

		// Human-like delay before request (1–3 seconds)
		HumanDelay(1, 3)

		// Build chromedp allocator with anti-bot options
		opts := ChromedpAllocatorOpts(proxyURL, s.ChromePath)

		allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), opts...)
		defer allocCancel()

		ctx, ctxCancel := chromedp.NewContext(allocCtx)
		defer ctxCancel()

		// Operation timeout
		ctx, timeoutCancel := context.WithTimeout(ctx, s.Timeout)
		defer timeoutCancel()

		var html string
		runErr := chromedp.Run(ctx,
			chromedp.Navigate(s.TargetURL),
			chromedp.WaitVisible("body", chromedp.ByQuery),
			// Small human-like delay before extraction
			chromedp.Sleep(time.Duration(1)*time.Second),
			chromedp.OuterHTML("html", &html),
		)
		if runErr != nil {
			return fmt.Errorf("chromedp navigation failed: %w", runErr)
		}

		// Validate we got meaningful HTML
		if strings.TrimSpace(html) == "" || len(html) < 100 {
			return fmt.Errorf("empty response from %s", s.TargetURL)
		}

		htmlContent = html
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scrape failed after retries: %w", err)
	}

	s.Logger.Info("page loaded successfully",
		slog.String("url", s.TargetURL),
		slog.String("proxy_used", usedProxy),
		slog.Int("html_length", len(htmlContent)),
	)

	// Parse HTML into structured job data
	jobs, err := ParseJobs(htmlContent, s.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	jobList := &models.JobList{
		LastUpdated: time.Now().Unix(),
		Jobs:        jobs,
	}

	s.Logger.Info("scrape complete",
		slog.Int("jobs_found", len(jobs)),
	)

	return jobList, nil
}
