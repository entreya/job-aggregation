package scraper

import (
	"math/rand"
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
