package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestFetchHTML_Success verifies that FetchHTML correctly fetches HTML from a
// local test server and returns the full body without error.
func TestFetchHTML_Success(t *testing.T) {
	const wantHTML = "<html><body><h1>Test Page</h1></body></html>"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify User-Agent is set (anti-bot header rotation)
		if r.Header.Get("User-Agent") == "" {
			t.Error("expected User-Agent header to be set, got empty string")
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(wantHTML))
	}))
	defer srv.Close()

	ctx := context.Background()
	// proxyURL = "" → direct connection (no proxy needed for local httptest)
	got, err := FetchHTML(ctx, srv.URL, "", 5*time.Second)
	if err != nil {
		t.Fatalf("FetchHTML returned unexpected error: %v", err)
	}
	if got != wantHTML {
		t.Errorf("body mismatch\ngot:  %q\nwant: %q", got, wantHTML)
	}
}

// TestFetchHTML_Non2xxStatus verifies that non-2xx HTTP responses are returned
// as errors with a descriptive message.
func TestFetchHTML_Non2xxStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := FetchHTML(context.Background(), srv.URL, "", 5*time.Second)
	if err == nil {
		t.Fatal("expected an error for 403 response, got nil")
	}
	if !strings.Contains(err.Error(), "403") {
		t.Errorf("expected error message to contain '403', got: %v", err)
	}
}

// TestFetchHTML_EmptyBody verifies that an empty response body is treated as an error.
func TestFetchHTML_EmptyBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write nothing — empty body
	}))
	defer srv.Close()

	_, err := FetchHTML(context.Background(), srv.URL, "", 5*time.Second)
	if err == nil {
		t.Fatal("expected an error for empty body, got nil")
	}
	if !strings.Contains(err.Error(), "empty response body") {
		t.Errorf("expected 'empty response body' in error, got: %v", err)
	}
}

// TestFetchHTML_ContextCancellation verifies that FetchHTML respects context
// cancellation and returns an appropriate error.
func TestFetchHTML_ContextCancellation(t *testing.T) {
	// Server that delays responding — context should cancel before it replies
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Block until the client disconnects
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := FetchHTML(ctx, srv.URL, "", 5*time.Second)
	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}
}

// TestFetchHTML_InvalidURL verifies that a malformed URL returns an error.
func TestFetchHTML_InvalidURL(t *testing.T) {
	_, err := FetchHTML(context.Background(), "://not-a-valid-url", "", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

// TestFetchHTML_InvalidProxyURL verifies that a malformed proxy URL returns a
// descriptive error before any network request is made.
func TestFetchHTML_InvalidProxyURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html>ok</html>"))
	}))
	defer srv.Close()

	_, err := FetchHTML(context.Background(), srv.URL, "://bad-proxy", 5*time.Second)
	if err == nil {
		t.Fatal("expected error for invalid proxy URL, got nil")
	}
	if !strings.Contains(err.Error(), "invalid proxy URL") {
		t.Errorf("expected 'invalid proxy URL' in error, got: %v", err)
	}
}
