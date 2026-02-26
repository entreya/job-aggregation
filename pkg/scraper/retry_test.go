package scraper

import (
	"errors"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func testRetryLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

func TestWithRetry_SucceedsOnFirstAttempt(t *testing.T) {
	cfg := RetryConfig{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	var calls int32

	err := WithRetry(cfg, "http://test.com", "proxy1", testRetryLogger(), func(attempt int) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}

func TestWithRetry_RetriesOnRetryableError(t *testing.T) {
	cfg := RetryConfig{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	var calls int32

	err := WithRetry(cfg, "http://test.com", "proxy1", testRetryLogger(), func(attempt int) error {
		c := atomic.AddInt32(&calls, 1)
		if c < 3 {
			return errors.New("connection timeout")
		}
		return nil
	})

	if err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_StopsOnNonRetryableError(t *testing.T) {
	cfg := RetryConfig{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	var calls int32

	err := WithRetry(cfg, "http://test.com", "proxy1", testRetryLogger(), func(attempt int) error {
		atomic.AddInt32(&calls, 1)
		return errors.New("authentication failed - forbidden")
	})

	if err == nil {
		t.Fatal("expected error for non-retryable failure")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call (no retry), got %d", calls)
	}
}

func TestWithRetry_ExhaustsMaxRetries(t *testing.T) {
	cfg := RetryConfig{MaxRetries: 3, BaseDelay: 10 * time.Millisecond}
	var calls int32

	err := WithRetry(cfg, "http://test.com", "proxy1", testRetryLogger(), func(attempt int) error {
		atomic.AddInt32(&calls, 1)
		return errors.New("connection refused")
	})

	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestWithRetry_ExponentialBackoffTiming(t *testing.T) {
	cfg := RetryConfig{MaxRetries: 3, BaseDelay: 50 * time.Millisecond}
	var calls int32

	start := time.Now()
	_ = WithRetry(cfg, "http://test.com", "proxy1", testRetryLogger(), func(attempt int) error {
		atomic.AddInt32(&calls, 1)
		return errors.New("i/o timeout")
	})
	elapsed := time.Since(start)

	// Expected backoff: 50ms (2^0 * 50ms) + 100ms (2^1 * 50ms) = 150ms minimum
	// Allow generous margin for test reliability
	if elapsed < 100*time.Millisecond {
		t.Errorf("expected at least 100ms of backoff, got %v", elapsed)
	}
}

func TestIsRetryable_RetryableErrors(t *testing.T) {
	retryableMsgs := []string{
		"connection timeout",
		"deadline exceeded",
		"connection refused",
		"connection reset by peer",
		"unexpected eof",
		"broken pipe",
		"i/o timeout",
		"tls handshake failed",
		"chromedp: target closed",
		"page crashed",
	}

	for _, msg := range retryableMsgs {
		if !isRetryable(errors.New(msg)) {
			t.Errorf("expected %q to be retryable", msg)
		}
	}
}

func TestIsRetryable_NonRetryableErrors(t *testing.T) {
	if isRetryable(nil) {
		t.Error("nil error should not be retryable")
	}

	nonRetryableMsgs := []string{
		"authentication failed",
		"permission denied",
		"invalid argument",
	}

	for _, msg := range nonRetryableMsgs {
		if isRetryable(errors.New(msg)) {
			t.Errorf("expected %q to be non-retryable", msg)
		}
	}
}

func TestRetryableError_Unwrap(t *testing.T) {
	inner := errors.New("connection refused")
	re := &RetryableError{
		URL:     "http://test.com",
		Proxy:   "proxy1",
		Attempt: 2,
		Err:     inner,
	}

	if !errors.Is(re, inner) {
		t.Error("expected Unwrap to return inner error")
	}

	expected := "attempt 2 for http://test.com (proxy=proxy1): connection refused"
	if re.Error() != expected {
		t.Errorf("expected error string %q, got %q", expected, re.Error())
	}
}

func TestWithRetry_ZeroMaxRetries_DefaultsToOne(t *testing.T) {
	cfg := RetryConfig{MaxRetries: 0, BaseDelay: 10 * time.Millisecond}
	var calls int32

	_ = WithRetry(cfg, "http://test.com", "", testRetryLogger(), func(attempt int) error {
		atomic.AddInt32(&calls, 1)
		return errors.New("connection timeout")
	})

	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("expected 1 call with zero MaxRetries, got %d", calls)
	}
}
