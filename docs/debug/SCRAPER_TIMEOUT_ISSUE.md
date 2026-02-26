# Debug Log: Scraper Timeout on GitHub Actions

## Issue
The scraper fails with `chromedp navigation failed: context deadline exceeded` on all 3 retry attempts when running on GitHub Actions without a proxy.

## Timeline
| Attempt | Start Time | Fail Time | Duration | Backoff |
|---------|-----------|-----------|----------|---------|
| 1       | 12:19:31  | 12:20:33  | ~62s     | 2s      |
| 2       | 12:20:35  | 12:21:38  | ~63s     | 4s      |
| 3       | 12:21:42  | (pending) | ~63s     | —       |

## Root Cause Analysis

### Confirmed: Site IS reachable
A direct HTTP fetch of `https://recruitment.nic.in/index_new.php` returns valid HTML with job listings. The site is **not down**.

### Confirmed: The issue is chromedp + GitHub Actions
- The 60s timeout at `scraper.go:88` is consumed entirely by Chrome trying to load the page.
- `chromedp.WaitVisible("body", chromedp.ByQuery)` waits for the DOM body — but this government site loads slowly (external resources, possible JS rendering) and may block datacenter IPs at the network level.
- GitHub Actions runners use shared datacenter IPs that are commonly blocked by anti-bot/firewall rules on government sites.
- Without `PROXY_URLS`, traffic goes direct — hitting these blocks.

### Contributing Factors
1. **No proxy configured**: `PROXY_URLS` is empty in CI, so no residential/rotating proxy is used.
2. **60s timeout too tight**: Chrome startup + navigation + WaitVisible + external resource loading can exceed 60s on a slow/blocked connection.
3. **`chromedp.WaitVisible("body")`**: Waits for full DOM rendering. On slow sites, even `body` can take abnormally long if Chrome is stalled on DNS/TLS to a blocked IP.
4. **No fallback strategy**: If chromedp fails, there's no fallback to a simple HTTP fetch (which works fine, as proven above).

## Recommendations
See `implementation_plan.md` for the proposed fix.
