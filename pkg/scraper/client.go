package scraper

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

// userAgents contains 10 real browser User-Agent strings for rotation.
// These represent the most common desktop browsers as of 2025–2026.
var userAgents = []string{
	// Chrome on Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	// Chrome on macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	// Firefox on Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:123.0) Gecko/20100101 Firefox/123.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
	// Safari on macOS
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.3 Safari/605.1.15",
	// Edge on Windows
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36 Edg/122.0.0.0",
	// Chrome on Linux
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36",
	// Firefox on Linux
	"Mozilla/5.0 (X11; Linux x86_64; rv:123.0) Gecko/20100101 Firefox/123.0",
}

// RandomUA returns a randomly selected User-Agent string from the pool.
func RandomUA() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// HumanDelay sleeps for a random duration between minSec and maxSec seconds,
// simulating human-like browsing behavior to reduce detection risk.
func HumanDelay(minSec, maxSec int) {
	if minSec < 0 {
		minSec = 0
	}
	if maxSec <= minSec {
		maxSec = minSec + 1
	}
	delay := time.Duration(minSec+rand.Intn(maxSec-minSec+1)) * time.Second
	time.Sleep(delay)
}

// ChromedpAllocatorOpts builds the full set of chromedp allocator options
// with anti-bot countermeasures: User-Agent rotation, proxy support,
// and headless Chrome flags optimized for scraping.
//
// proxyURL: proxy address for chromedp.ProxyServer() (empty = direct)
// chromePath: custom Chrome/Chromium path (empty = auto-detect)
func ChromedpAllocatorOpts(proxyURL string, chromePath string) []chromedp.ExecAllocatorOption {
	ua := RandomUA()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true), // Prevents /dev/shm OOM on low-memory VPS
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.UserAgent(ua),
	)

	// Proxy support: chromedp routes all browser traffic through this proxy
	if proxyURL != "" {
		opts = append(opts, chromedp.ProxyServer(proxyURL))
	}

	// Custom browser path (e.g., /usr/bin/chromium-browser on ARM)
	if chromePath == "" {
		chromePath = os.Getenv("CHROME_PATH")
	}
	if chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}

	return opts
}

// FetchHTML performs a simple net/http GET to the given URL and returns the
// full response body as a string. It uses User-Agent rotation to appear as a
// real browser. This is used as a lightweight fallback when chromedp fails.
//
// proxyURL: optional HTTP/HTTPS proxy address (e.g. "http://user:pass@host:8080").
//
//	Pass empty string for a direct connection.
//
// timeout:  maximum time allowed for the full request/response cycle.
func FetchHTML(ctx context.Context, targetURL string, proxyURL string, timeout time.Duration) (string, error) {
	transport := &http.Transport{
		DisableKeepAlives:   false,
		IdleConnTimeout:     30 * time.Second,
		TLSHandshakeTimeout: 15 * time.Second,
	}

	// Route through the proxy when one is configured, mirroring the chromedp path.
	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err != nil {
			return "", fmt.Errorf("invalid proxy URL %q: %w", proxyURL, err)
		}
		transport.Proxy = http.ProxyURL(parsed)
	}

	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to build HTTP request: %w", err)
	}

	// Rotate User-Agent to appear as a real browser
	req.Header.Set("User-Agent", RandomUA())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP GET failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected HTTP status: %d %s", resp.StatusCode, resp.Status)
	}

	// Read with a reasonable size cap (10 MB) to avoid memory exhaustion
	const maxBodySize = 10 << 20 // 10 MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBodySize))
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return "", fmt.Errorf("empty response body from %s", targetURL)
	}

	return string(body), nil
}
