package scraper

import (
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"
)

// RetryConfig controls the retry behavior for scrape operations.
type RetryConfig struct {
	MaxRetries int           // Maximum number of retry attempts (e.g., 3)
	BaseDelay  time.Duration // Base delay for exponential backoff (e.g., 2s → 2s, 4s, 8s)
}

// DefaultRetryConfig returns a sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  2 * time.Second,
	}
}

// RetryableError wraps an error with metadata for structured logging.
type RetryableError struct {
	URL     string
	Proxy   string
	Attempt int
	Err     error
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("attempt %d for %s (proxy=%s): %s", e.Attempt, e.URL, e.Proxy, e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// isRetryable determines if an error warrants a retry.
// Retries on: network/timeout errors, chromedp context errors, empty responses.
// Does NOT retry on: nil errors, context cancellation by caller.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	errMsg := strings.ToLower(err.Error())

	// Non-retryable: intentional context cancellation
	if errors.Is(err, errors.New("context canceled")) {
		return false
	}

	// Retryable patterns: network failures, timeouts, chromedp crashes
	retryablePatterns := []string{
		"timeout",
		"deadline exceeded",
		"connection refused",
		"connection reset",
		"eof",
		"broken pipe",
		"no such host",
		"i/o timeout",
		"tls handshake",
		"chromedp",
		"devtools",
		"target closed",
		"page crashed",
		"empty response",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// WithRetry executes fn with exponential backoff retry logic.
// It logs each attempt with structured context (URL, proxy, attempt number, error).
//
// The fn receives the current attempt number (0-indexed).
// If fn returns nil, WithRetry returns immediately.
// If fn returns a non-retryable error, WithRetry returns immediately with the error.
// If all retries are exhausted, returns the last error.
func WithRetry(cfg RetryConfig, url string, proxy string, logger *slog.Logger, fn func(attempt int) error) error {
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 1
	}

	var lastErr error
	for attempt := 0; attempt < cfg.MaxRetries; attempt++ {
		err := fn(attempt)
		if err == nil {
			if attempt > 0 {
				logger.Info("retry succeeded",
					slog.String("url", url),
					slog.String("proxy_used", proxy),
					slog.Int("attempt", attempt+1),
				)
			}
			return nil
		}

		lastErr = err

		logger.Warn("scrape attempt failed",
			slog.String("url", url),
			slog.String("proxy_used", proxy),
			slog.Int("attempt", attempt+1),
			slog.Int("max_retries", cfg.MaxRetries),
			slog.String("error", err.Error()),
		)

		// Do not retry if the error is non-retryable
		if !isRetryable(err) {
			logger.Error("non-retryable error — aborting",
				slog.String("url", url),
				slog.String("error", err.Error()),
			)
			return &RetryableError{
				URL:     url,
				Proxy:   proxy,
				Attempt: attempt + 1,
				Err:     err,
			}
		}

		// Exponential backoff: baseDelay * 2^attempt
		// attempt 0 → 2s, attempt 1 → 4s, attempt 2 → 8s
		if attempt < cfg.MaxRetries-1 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * cfg.BaseDelay
			logger.Info("backing off before retry",
				slog.String("url", url),
				slog.Duration("backoff", backoff),
				slog.Int("next_attempt", attempt+2),
			)
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("all %d retry attempts exhausted for %s: %w", cfg.MaxRetries, url, lastErr)
}
