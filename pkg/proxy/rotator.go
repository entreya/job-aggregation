package proxy

import (
	"crypto/tls"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

// Strategy defines how the next proxy is selected.
type Strategy string

const (
	StrategyRandom     Strategy = "random"
	StrategyRoundRobin Strategy = "round-robin"
)

// ProxyRotator manages a pool of proxy URLs and selects the next one
// based on a configurable strategy. Thread-safe via atomic operations.
type ProxyRotator struct {
	proxies  []string     // Validated proxy URLs
	strategy Strategy     // Selection strategy
	index    atomic.Int64 // Round-robin counter (atomic for concurrency safety)
	logger   *slog.Logger
}

// NewRotator creates a ProxyRotator from a comma-separated list of proxy URLs
// and a selection strategy. Validates each URL format.
//
// proxyURLs format: "http://user:pass@host:port,http://user:pass@host2:port2"
// strategy: "random" or "round-robin" (defaults to "round-robin" if invalid)
//
// Returns a rotator with zero proxies (direct fallback) if proxyURLs is empty.
func NewRotator(proxyURLs string, strategy string, logger *slog.Logger) (*ProxyRotator, error) {
	r := &ProxyRotator{
		proxies: make([]string, 0),
		logger:  logger,
	}

	// Parse strategy
	switch Strategy(strings.ToLower(strings.TrimSpace(strategy))) {
	case StrategyRandom:
		r.strategy = StrategyRandom
	default:
		r.strategy = StrategyRoundRobin
	}

	// Parse and validate proxy URLs
	if strings.TrimSpace(proxyURLs) == "" {
		r.logger.Warn("PROXY_URLS is empty — falling back to direct connection")
		return r, nil
	}

	parts := strings.Split(proxyURLs, ",")
	for _, raw := range parts {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}

		if err := validateProxyURL(raw); err != nil {
			return nil, fmt.Errorf("invalid proxy URL %q: %w", raw, err)
		}

		r.proxies = append(r.proxies, raw)
	}

	r.logger.Info("proxy rotator initialized",
		slog.Int("proxy_count", len(r.proxies)),
		slog.String("strategy", string(r.strategy)),
	)

	return r, nil
}

// Next returns the next proxy URL based on the configured strategy.
// Returns an empty string if no proxies are configured (direct connection fallback).
func (r *ProxyRotator) Next() string {
	if len(r.proxies) == 0 {
		return ""
	}

	if len(r.proxies) == 1 {
		return r.proxies[0]
	}

	switch r.strategy {
	case StrategyRandom:
		return r.proxies[rand.Intn(len(r.proxies))]
	default: // round-robin
		idx := r.index.Add(1) - 1
		return r.proxies[idx%int64(len(r.proxies))]
	}
}

// HasProxies reports whether the rotator has any configured proxies.
func (r *ProxyRotator) HasProxies() bool {
	return len(r.proxies) > 0
}

// Count returns the number of configured proxies.
func (r *ProxyRotator) Count() int {
	return len(r.proxies)
}

// HTTPClient returns a configured *http.Client using the next proxy in rotation.
// Includes a 30-second timeout and TLS configuration.
// If no proxies are configured, returns a client with direct connection.
func (r *ProxyRotator) HTTPClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	}

	proxyURL := r.Next()
	if proxyURL != "" {
		parsed, err := url.Parse(proxyURL)
		if err == nil {
			transport.Proxy = http.ProxyURL(parsed)
			r.logger.Info("http client using proxy",
				slog.String("proxy", sanitizeProxyLog(proxyURL)),
			)
		}
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}

// ProxyServerAddr returns the proxy URL formatted for chromedp.ProxyServer().
// chromedp expects the format: "http://host:port" (auth handled separately).
// Returns empty string for direct connection.
func (r *ProxyRotator) ProxyServerAddr() string {
	return r.Next()
}

// validateProxyURL checks that a proxy URL is well-formed.
func validateProxyURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" && parsed.Scheme != "socks5" {
		return fmt.Errorf("unsupported scheme %q (expected http, https, or socks5)", parsed.Scheme)
	}

	if parsed.Host == "" {
		return fmt.Errorf("missing host")
	}

	return nil
}

// sanitizeProxyLog masks the password in a proxy URL for safe logging.
func sanitizeProxyLog(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "<invalid-url>"
	}

	if parsed.User != nil {
		username := parsed.User.Username()
		parsed.User = url.UserPassword(username, "****")
	}

	return parsed.String()
}
