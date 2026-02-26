# 🔍 Scraping Strategy: `recruitment.nic.in` from GitHub Actions

> **Author**: Auto-generated research document
> **Date**: 2026-02-13
> **Status**: Research Complete

---

## Table of Contents
1. [Root Cause: Why GitHub Actions Gets Blocked](#1-root-cause)
2. [Solution Matrix](#2-solution-matrix)
3. [Approach 1: Self-Hosted Runner on Indian VPS (⭐ Recommended)](#3-approach-1-self-hosted-runner)
4. [Approach 2: Paid Residential Proxies](#4-approach-2-residential-proxies)
5. [Approach 3: Scraping APIs](#5-approach-3-scraping-apis)
6. [Approach 4: Cloudflare WARP (Free, Unreliable)](#6-approach-4-cloudflare-warp)
7. [Approach 5: Direct Access with Header Manipulation](#7-approach-5-direct-access)
8. [Go Code: Proxy Rotation with Chromedp](#8-go-code-examples)
9. [Final Recommendation](#9-final-recommendation)

---

## 1. Root Cause: Why GitHub Actions Gets Blocked {#1-root-cause}

### The Problem
GitHub-hosted runners use **Azure datacenter IPs** (ranges are [publicly listed](https://api.github.com/meta)). Indian government websites like `recruitment.nic.in` employ multiple layers of blocking:

| Detection Layer         | How It Works                                                          | Difficulty to Bypass |
|------------------------|-----------------------------------------------------------------------|---------------------|
| **Datacenter IP Blocklist** | NIC servers maintain blocklists of known cloud provider IP ranges (Azure, AWS, GCP). | Hard               |
| **GeoIP Filtering**     | Some `.nic.in` endpoints are geo-fenced to Indian IPs only.           | Medium              |
| **Rate Limiting**       | Aggressive rate limits on non-Indian IPs.                             | Easy                |
| **User-Agent Filtering**| Blocks known bot/library user-agents (`Go-http-client`, `python-requests`). | Easy                |
| **TLS Fingerprinting**  | JA3/JA4 fingerprints from Go's `net/http` differ from real browsers.  | Hard (without browser) |

### Why Free Proxies (ProxyScrape) Fail (~90% failure rate)
- Free proxy lists are overwhelmingly **datacenter IPs** — same problem as GitHub Actions.
- Free proxies have **high latency** (>5s), causing timeouts.
- Many free proxies are already **blacklisted** by government sites.
- Free proxy pools have **no Indian-origin IPs** available.
- Connection reliability is **< 10%** due to proxy churn and overuse.

---

## 2. Solution Matrix {#2-solution-matrix}

| Approach                           | Reliability | Monthly Cost     | Setup Complexity | Best For                     |
|-----------------------------------|-------------|-----------------|-----------------|-------------------------------|
| **🆓 Oracle Cloud Free VPS (Mumbai)** | ⭐⭐⭐⭐⭐  | **$0 forever**  | Medium (one-time) | **Your use case** ✅          |
| **🆓 ScraperAPI Free Tier**        | ⭐⭐⭐⭐    | **$0** (1K credits) | Very Easy     | Easiest zero-setup option     |
| Indian VPS (Paid)                  | ⭐⭐⭐⭐⭐  | $2.50 – $6/mo   | Medium (one-time) | If Oracle capacity is full    |
| Paid Residential Proxies           | ⭐⭐⭐⭐    | $7 – $49/mo     | Easy              | Multi-site, high-volume       |
| Cloudflare WARP                    | ⭐⭐        | Free             | Medium            | Non-geo-blocked sites         |
| Direct Headers Bypass              | ⭐          | Free             | Easy              | Non-protected sites only      |

---

## 3. Approach 1: Oracle Cloud Always Free VPS (⭐ RECOMMENDED — $0) {#3-approach-1-self-hosted-runner}

> [!IMPORTANT]
> **Oracle Cloud offers a PERMANENTLY FREE Indian VPS** in Mumbai. This is not a trial — it's their "Always Free" tier with no expiration. This is the best solution for your use case.

### What You Get (Free Forever)

| Resource         | Free Allocation                            |
|-----------------|-------------------------------------------|
| **Compute**      | ARM A1 Flex: up to **4 OCPUs + 24GB RAM** |
| **Storage**      | **200GB** block storage                    |
| **Network**      | 10TB/month outbound                       |
| **Location**     | **Mumbai, India** 🇮🇳                     |
| **OS**           | Ubuntu 22.04/24.04 (aarch64)              |

### Why This Works
- Traffic originates from an **Indian ISP-grade IP** (Oracle's Mumbai DC) — not Azure/AWS.
- The VPS is in India, **eliminating geo-blocking entirely**.
- No proxy overhead = **fastest execution** (~2-5s page loads).
- You already have the workflow (`.github/workflows/scraper-vps.yml`).

### Setup Steps

1. **Sign up** at [cloud.oracle.com](https://cloud.oracle.com) (credit card required for identity, **never charged**).
2. **Select Home Region → India South (Mumbai)**. ⚠️ Cannot be changed later!
3. **Create Compute Instance**:
   - Shape: `VM.Standard.A1.Flex` (Ampere ARM)
   - OCPUs: 1 (1 is plenty for a cron scraper)
   - RAM: 6GB
   - Image: Ubuntu 22.04 Minimal (aarch64)
   - Boot volume: 50GB
4. **Install dependencies** (SSH into the instance):
   ```bash
   # Update system
   sudo apt update && sudo apt upgrade -y
   sudo apt install -y curl wget git

   # Install Go (ARM64)
   wget https://go.dev/dl/go1.23.0.linux-arm64.tar.gz
   sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.23.0.linux-arm64.tar.gz
   echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
   source ~/.profile

   # Install Chromium (Chrome is NOT available for ARM Linux)
   sudo apt install -y chromium-browser
   # Verify
   chromium-browser --version
   ```
5. **Configure GitHub Actions Runner** (ARM64 version):
   ```bash
   mkdir actions-runner && cd actions-runner
   # Download ARM64 runner (get URL from GitHub Settings → Runners → New self-hosted runner → Linux → ARM64)
   curl -o actions-runner-linux-arm64-2.x.tar.gz -L <GITHUB_PROVIDED_URL>
   tar xzf ./actions-runner-linux-arm64-2.x.tar.gz
   ./config.sh --url https://github.com/entreya/job-aggregation --token <YOUR_TOKEN>
   sudo ./svc.sh install
   sudo ./svc.sh start
   ```

> [!WARNING]
> **ARM Caveat:** Since Oracle's free tier uses ARM (aarch64), you must use `chromium-browser` instead of `google-chrome`. Update your workflow's verify step to check `chromium-browser --version` instead.

### VPS Fallback Providers (If Oracle Capacity is Full)

Oracle's free ARM instances are popular. If you get "Out of Capacity" errors, retry later or use a paid fallback:

| Provider    | Location   | Cheapest Plan | RAM  | Storage | Notes                          |
|------------|-----------|---------------|------|---------|--------------------------------|
| **Vultr**   | Mumbai    | **$2.50/mo**  | 512MB | 10GB SSD | Cheapest paid option.          |
| **Linode**  | Mumbai    | **$5/mo**     | 1GB  | 25GB SSD | Reliable. Has Chennai too.     |
| **LordCloud** | Mumbai  | **₹299/mo (~$3.50)** | 1GB | 20GB | Indian company.                |

### Workflow (Already Done)
Your `scraper-vps.yml` is correct:
```yaml
runs-on: self-hosted  # Routes to your Indian VPS
```

### Cost Analysis
```
Oracle Cloud Always Free: $0/mo = $0/year  🎉
4 scrapes/day × 365 = 1,460 scrapes/year
Cost per scrape: $0.00
```

---

## 4. Approach 2: Paid Residential Proxies {#4-approach-2-residential-proxies}

If you don't want to maintain a VPS, residential proxies are the next best option.

### How They Work
Residential proxies route traffic through **real ISP-assigned IPs** (BSNL, Airtel, Jio).
Government sites see a genuine Indian home user, not a datacenter.

### Provider Comparison (India Residential Proxies)

| Provider        | India IPs    | Price (Pay-as-you-go) | Min Purchase | Geo-Targeting | Best Feature                  |
|----------------|-------------|----------------------|-------------|--------------|-------------------------------|
| **IPRoyal**     | 144K+       | **$7/GB** (1GB)       | $7           | City-level    | Non-expiring traffic 💰       |
| **Bright Data** | Millions    | $4–8/GB               | ~$4          | ASN-level     | Largest pool, AI unblocker    |
| **Oxylabs**     | 12M+        | $8/GB                 | $49 (5GB)    | City-level    | Highest success rates         |
| **Decodo**      | 9.4M+       | ~$4/GB                | Varies       | City-level    | Fastest connections           |
| **SOAX**        | 5M+         | ~$5/GB                | Varies       | ISP-level     | 99.5% success rate            |

### Bandwidth Estimation for Your Use Case
```
1 page load (recruitment.nic.in) ≈ 200KB–500KB
4 scrapes/day = 2MB/day = ~60MB/month
With overhead and retries = ~200MB/month max

→ At $7/GB (IPRoyal): < $1.40/month
→ At $4/GB (Bright Data with coupon): < $0.80/month
```

> [!TIP]
> For your low-bandwidth use case (scraping a single page 4x/day), **residential proxies cost less than $2/month**. IPRoyal's non-expiring traffic model is ideal — buy 1GB and it lasts months.

### Go Code: Using a Residential Proxy with Chromedp

```go
package scraper

import (
    "context"
    "github.com/chromedp/chromedp"
)

func ScrapeWithProxy(targetURL, proxyAddr string) (string, error) {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-gpu", true),
        // Route all Chrome traffic through the proxy
        chromedp.ProxyServer(proxyAddr), // e.g., "http://user:pass@gate.iproyal.com:12321"
        chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) ..."),
    )

    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    defer cancel()

    ctx, cancel := chromedp.NewContext(allocCtx)
    defer cancel()

    var htmlContent string
    err := chromedp.Run(ctx,
        chromedp.Navigate(targetURL),
        chromedp.WaitVisible("body"),
        chromedp.OuterHTML("html", &htmlContent),
    )
    return htmlContent, err
}
```

### Workflow: GitHub Actions + Residential Proxy
```yaml
- name: Run Scraper with Proxy
  env:
    PROXY_URL: ${{ secrets.RESIDENTIAL_PROXY_URL }}
    # e.g., "http://user:pass@gate.iproyal.com:12321"
  run: go run cmd/scraper/main.go --proxy "$PROXY_URL"
```

---

## 5. Approach 3: Scraping APIs {#5-approach-3-scraping-apis}

Managed services that handle proxies, CAPTCHAs, and browser rendering for you.

### Provider Comparison

| Provider       | Free Tier                    | Paid Plans      | India Support | Key Feature                  |
|---------------|-----------------------------|-----------------|--------------|-----------------------------|
| **ScraperAPI** | 1,000 credits/mo + 7-day trial | $49/mo (100K)  | Yes (`country_code=in`) | 3.5M rotating India IPs    |
| **Scrape.do**  | 1,000 credits/mo            | $29/mo          | Yes           | Built-in Cloudflare bypass  |
| **ZenRows**    | 1,000 credits/mo            | $49/mo          | Yes           | Rotating residential proxies |

### Example: ScraperAPI Integration (Go)
```go
package main

import (
    "fmt"
    "io"
    "net/http"
    "net/url"
)

const scraperAPIKey = "YOUR_API_KEY" // From GitHub Secrets

func scrapeViaAPI(targetURL string) (string, error) {
    apiURL := fmt.Sprintf(
        "http://api.scraperapi.com?api_key=%s&url=%s&country_code=in&render=true",
        scraperAPIKey,
        url.QueryEscape(targetURL),
    )

    resp, err := http.Get(apiURL)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    return string(body), err
}
```

> [!IMPORTANT]
> ScraperAPI's free tier (1,000 credits/mo) is **enough for your use case** (4 requests/day × 30 days = 120 credits). This is the **easiest zero-cost, zero-infrastructure option**. Sign up, get API key, add to GitHub Secrets, done.

---

## 6. Approach 4: Cloudflare WARP (Free, Unreliable) {#6-approach-4-cloudflare-warp}

### How It Works
Cloudflare WARP routes traffic through Cloudflare's network, but **does not guarantee an Indian IP**.

### Setup (GitHub Actions)
```yaml
- name: Setup Cloudflare WARP
  uses: Boostport/setup-cloudflare-warp@v1
  with:
    organization: ${{ secrets.CF_ORG }}
    auth_client_id: ${{ secrets.CF_AUTH_CLIENT_ID }}
    auth_client_secret: ${{ secrets.CF_AUTH_CLIENT_SECRET }}
```

### Requirements
- Cloudflare Zero Trust account (free tier available).
- Generate a Service Token in Zero Trust dashboard.
- Configure Device Enrollment policy with `SERVICE_AUTH` rule.

### Why It's Unreliable for This Use Case
| Issue                        | Impact |
|------------------------------|--------|
| WARP doesn't give Indian IPs | Geo-blocking still applies |
| WARP uses Cloudflare DCs     | May still be blocked as non-residential |
| Requires Zero Trust setup    | Complex configuration for uncertain results |
| Free tier limitations        | Rate limits apply |

> [!WARNING]
> Cloudflare WARP is designed for **secure access to internal resources**, not for geo-spoofing. It will not reliably bypass `recruitment.nic.in`'s geo-blocking.

---

## 7. Approach 5: Direct Access with Header Manipulation {#7-approach-5-direct-access}

### Can You Just Fix the Headers?
**Short answer: No.** Here's why:

The blocking is primarily at the **IP level**, not the header level. However, proper headers are still important as a secondary measure. Here's what helps and what doesn't:

| Technique                      | Helps? | Why                                                     |
|-------------------------------|--------|-------------------------------------------------------|
| Realistic User-Agent           | ✅ Minor | Blocks default Go/Python UAs, but not the main issue  |
| Accept/Accept-Language headers | ✅ Minor | Adds legitimacy to requests                            |
| Browser-like `Sec-*` headers   | ❌ No   | These are sent by the browser (chromedp), not settable |
| Referer header                 | ❌ No   | Government sites don't check referrer                  |
| TLS fingerprint spoofing       | ❌ No   | Chromedp uses real Chrome = already legitimate TLS     |
| Random delays between requests | ✅ Yes  | Avoids rate limiting, but doesn't help with IP blocks  |

### What Chromedp Already Gives You
Since you're using `chromedp` (real Chrome), you already get:
- ✅ Legitimate TLS fingerprint (JA3/JA4)
- ✅ JavaScript execution (bypasses JS challenges)
- ✅ Authentic browser headers (Sec-CH-UA, etc.)
- ✅ Cookie handling

**The only remaining issue is the IP address.**

---

## 8. Go Code: Robust Proxy Rotation with Chromedp {#8-go-code-examples}

### Full Implementation Pattern

```go
package scraper

import (
    "context"
    "fmt"
    "log"
    "math/rand"
    "strings"
    "time"

    "github.com/PuerkitoBio/goquery"
    "github.com/chromedp/chromedp"
)

// ProxyConfig holds proxy authentication details.
type ProxyConfig struct {
    URL      string // e.g., "http://gate.iproyal.com:12321"
    Username string
    Password string
}

// ScraperConfig holds all scraper settings.
type ScraperConfig struct {
    TargetURL string
    Proxies   []ProxyConfig
    Timeout   time.Duration
    MaxRetries int
}

// ScrapeWithRotation attempts to scrape using rotating proxies with retries.
func ScrapeWithRotation(cfg ScraperConfig) (string, error) {
    if len(cfg.Proxies) == 0 {
        // No proxies configured — try direct connection
        return scrapeOnce(cfg.TargetURL, "", cfg.Timeout)
    }

    var lastErr error
    for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
        // Pick a random proxy (or rotate sequentially)
        proxy := cfg.Proxies[rand.Intn(len(cfg.Proxies))]
        proxyURL := formatProxyURL(proxy)

        log.Printf("[Attempt %d/%d] Using proxy: %s", attempt+1, cfg.MaxRetries, proxy.URL)

        html, err := scrapeOnce(cfg.TargetURL, proxyURL, cfg.Timeout)
        if err != nil {
            lastErr = err
            log.Printf("Proxy %s failed: %v", proxy.URL, err)
            // Add jitter between retries to avoid patterns
            time.Sleep(time.Duration(2+rand.Intn(3)) * time.Second)
            continue
        }

        return html, nil
    }

    return "", fmt.Errorf("all %d proxy attempts exhausted: %w", cfg.MaxRetries, lastErr)
}

func scrapeOnce(targetURL, proxyURL string, timeout time.Duration) (string, error) {
    opts := append(chromedp.DefaultExecAllocatorOptions[:],
        chromedp.Flag("headless", true),
        chromedp.Flag("no-sandbox", true),
        chromedp.Flag("disable-gpu", true),
        chromedp.Flag("disable-dev-shm-usage", true),
        chromedp.UserAgent(randomUserAgent()),
    )

    // Add proxy if configured
    if proxyURL != "" {
        opts = append(opts, chromedp.ProxyServer(proxyURL))
    }

    allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
    defer cancel()

    ctx, cancel := chromedp.NewContext(allocCtx)
    defer cancel()

    ctx, cancel = context.WithTimeout(ctx, timeout)
    defer cancel()

    var htmlContent string
    err := chromedp.Run(ctx,
        chromedp.Navigate(targetURL),
        chromedp.WaitVisible("body", chromedp.ByQuery),
        // Small human-like delay
        chromedp.Sleep(time.Duration(1+rand.Intn(2)) * time.Second),
        chromedp.OuterHTML("html", &htmlContent),
    )

    return htmlContent, err
}

func formatProxyURL(p ProxyConfig) string {
    if p.Username != "" && p.Password != "" {
        // Format: http://user:pass@host:port
        parts := strings.TrimPrefix(p.URL, "http://")
        return fmt.Sprintf("http://%s:%s@%s", p.Username, p.Password, parts)
    }
    return p.URL
}

// randomUserAgent returns a random realistic browser User-Agent string.
func randomUserAgent() string {
    userAgents := []string{
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:121.0) Gecko/20100101 Firefox/121.0",
        "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.2 Safari/605.1.15",
    }
    return userAgents[rand.Intn(len(userAgents))]
}
```

### Usage in `cmd/scraper/main.go`

```go
// Option A: Direct (for VPS runner — no proxy needed)
s := scraper.NewScraper("https://recruitment.nic.in/index_new.php")
jobsList, err := s.Scrape()

// Option B: With proxy rotation (for GitHub-hosted runner)
proxyURL := os.Getenv("PROXY_URL") // Set via GitHub Secrets
cfg := scraper.ScraperConfig{
    TargetURL:  "https://recruitment.nic.in/index_new.php",
    Proxies:    []scraper.ProxyConfig{{URL: proxyURL}},
    Timeout:    60 * time.Second,
    MaxRetries: 3,
}
html, err := scraper.ScrapeWithRotation(cfg)
```

---

## 9. Final Recommendation {#9-final-recommendation}

### For Your Use Case: **FREE Solutions Only**

```
┌──────────────────────────────────────────────────────┐
│  RECOMMENDED STACK (100% FREE)                       │
│                                                      │
│  Primary:  Oracle Cloud Always Free (Mumbai)         │
│            + Self-Hosted GitHub Runner               │
│            + scraper-vps.yml (already created)       │
│                                                      │
│  Fallback: ScraperAPI free tier (1K credits/mo)      │
│            for when VPS is down or provisioning      │
│                                                      │
│  Total Cost: $0/mo ($0/year)  🎉                     │
└──────────────────────────────────────────────────────┘
```

### Quick-Start Path (Choose One)

#### Option A: ScraperAPI (5 minutes, zero infra)
1. Sign up at [scraperapi.com](https://scraperapi.com) (free, 1K credits/mo)
2. Get your API key
3. Add `SCRAPER_API_KEY` to GitHub Secrets
4. Update scraper to use the API → Done!

#### Option B: Oracle Cloud Free VPS (30 minutes, most reliable)
1. Sign up at [cloud.oracle.com](https://cloud.oracle.com)
2. Select **India South (Mumbai)** as home region
3. Create ARM A1 instance (free forever)
4. Install Go + Chromium + GitHub Runner
5. Commit & push `scraper-vps.yml` → Done!

### Action Plan

| # | Action                              | Status      | Notes                                   |
|---|-------------------------------------|-------------|----------------------------------------|
| 1 | Sign up for Oracle Cloud Free       | 🔵 To Do    | Select Mumbai region                   |
| 2 | Create ARM A1 VM (1 OCPU, 6GB)     | 🔵 To Do    | $0 forever                             |
| 3 | Install Go + Chromium on VM         | 🔵 To Do    | ARM64 builds                           |
| 4 | Configure GitHub Runner (ARM64)     | 🔵 To Do    | `config.sh --url ... --token ...`       |
| 5 | Install runner as service           | 🔵 To Do    | `svc.sh install && start`              |
| 6 | Commit `scraper-vps.yml`            | 🟡 Pending  | Already created                        |
| 7 | (Quick Alt) Sign up for ScraperAPI  | 🔵 Optional | If Oracle capacity unavailable         |

### Why NOT Other Approaches

| Approach              | Why Not for You                                        |
|----------------------|-------------------------------------------------------|
| Paid VPS              | Oracle Free does the same thing. No need to pay.       |
| Residential Proxies   | Costs money. Overkill for 4 requests/day.              |
| Cloudflare WARP       | Doesn't give Indian IPs. Won't bypass geo-blocking.   |
| Free Proxies          | < 10% success rate. Already proven unreliable.         |
| Direct Headers        | IP-level blocking can't be bypassed with headers.      |

---

## Appendix: Legal Considerations

- `recruitment.nic.in` is a **public government portal** publishing job vacancies.
- The data is **publicly available** and intended for citizen access.
- Scraping at 4x/day (every 6 hours) is **well within reasonable usage**.
- Always check `robots.txt` for any restrictions.
- Be mindful of India's **Digital Personal Data Protection Act (DPDPA)** — the scraped data (job titles, departments, URLs) is **not personal data**.
