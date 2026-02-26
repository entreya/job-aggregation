package proxy

import (
	"log/slog"
	"os"
	"sync"
	"testing"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

func TestNewRotator_EmptyURLs_FallbackDirect(t *testing.T) {
	r, err := NewRotator("", "round-robin", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.HasProxies() {
		t.Error("expected no proxies when PROXY_URLS is empty")
	}
	if r.Count() != 0 {
		t.Errorf("expected 0 proxies, got %d", r.Count())
	}
	if got := r.Next(); got != "" {
		t.Errorf("expected empty string for direct fallback, got %q", got)
	}
}

func TestNewRotator_MalformedURL_ReturnsError(t *testing.T) {
	_, err := NewRotator("not-a-url", "random", testLogger())
	if err == nil {
		t.Error("expected error for malformed proxy URL")
	}
}

func TestNewRotator_InvalidScheme_ReturnsError(t *testing.T) {
	_, err := NewRotator("ftp://proxy.example.com:8080", "random", testLogger())
	if err == nil {
		t.Error("expected error for unsupported scheme ftp://")
	}
}

func TestNewRotator_ValidURLs(t *testing.T) {
	urls := "http://user:pass@proxy1.example.com:8080,http://user:pass@proxy2.example.com:8080"
	r, err := NewRotator(urls, "round-robin", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Count() != 2 {
		t.Errorf("expected 2 proxies, got %d", r.Count())
	}
}

func TestNewRotator_TrimsWhitespace(t *testing.T) {
	urls := "  http://proxy1.example.com:8080 , http://proxy2.example.com:8080  "
	r, err := NewRotator(urls, "random", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Count() != 2 {
		t.Errorf("expected 2 proxies, got %d", r.Count())
	}
}

func TestNewRotator_SkipsEmptyParts(t *testing.T) {
	urls := "http://proxy1.example.com:8080,,,http://proxy2.example.com:8080,"
	r, err := NewRotator(urls, "random", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Count() != 2 {
		t.Errorf("expected 2 proxies, got %d", r.Count())
	}
}

func TestRoundRobin_ReturnsCorrectOrder(t *testing.T) {
	urls := "http://p1.example.com:8080,http://p2.example.com:8080,http://p3.example.com:8080"
	r, err := NewRotator(urls, "round-robin", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{
		"http://p1.example.com:8080",
		"http://p2.example.com:8080",
		"http://p3.example.com:8080",
		"http://p1.example.com:8080", // Wraps around
		"http://p2.example.com:8080",
	}

	for i, want := range expected {
		got := r.Next()
		if got != want {
			t.Errorf("call %d: expected %q, got %q", i+1, want, got)
		}
	}
}

func TestRoundRobin_ConcurrencySafe(t *testing.T) {
	urls := "http://p1.example.com:8080,http://p2.example.com:8080"
	r, err := NewRotator(urls, "round-robin", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Hammer it concurrently to detect race conditions
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := r.Next()
			if result != "http://p1.example.com:8080" && result != "http://p2.example.com:8080" {
				t.Errorf("unexpected proxy: %q", result)
			}
		}()
	}
	wg.Wait()
}

func TestRandomSelection_StaysWithinBounds(t *testing.T) {
	urls := "http://p1.example.com:8080,http://p2.example.com:8080,http://p3.example.com:8080"
	r, err := NewRotator(urls, "random", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validProxies := map[string]bool{
		"http://p1.example.com:8080": true,
		"http://p2.example.com:8080": true,
		"http://p3.example.com:8080": true,
	}

	for i := 0; i < 50; i++ {
		got := r.Next()
		if !validProxies[got] {
			t.Errorf("call %d: proxy %q is not in the valid set", i+1, got)
		}
	}
}

func TestSingleProxy_AlwaysReturnsSame(t *testing.T) {
	r, err := NewRotator("http://only.example.com:8080", "round-robin", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < 5; i++ {
		got := r.Next()
		if got != "http://only.example.com:8080" {
			t.Errorf("expected the single proxy, got %q", got)
		}
	}
}

func TestHTTPClient_ReturnsClient(t *testing.T) {
	r, err := NewRotator("http://proxy.example.com:8080", "random", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client := r.HTTPClient()
	if client == nil {
		t.Fatal("expected non-nil http.Client")
	}
	if client.Timeout == 0 {
		t.Error("expected non-zero timeout on client")
	}
}

func TestHTTPClient_DirectConnection(t *testing.T) {
	r, err := NewRotator("", "random", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	client := r.HTTPClient()
	if client == nil {
		t.Fatal("expected non-nil http.Client for direct connection")
	}
}

func TestSocks5ProxyValid(t *testing.T) {
	r, err := NewRotator("socks5://user:pass@proxy.example.com:1080", "random", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Count() != 1 {
		t.Errorf("expected 1 proxy, got %d", r.Count())
	}
}

func TestDefaultStrategy(t *testing.T) {
	r, err := NewRotator("http://proxy.example.com:8080", "invalid-strategy", testLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.strategy != StrategyRoundRobin {
		t.Errorf("expected default round-robin strategy, got %q", r.strategy)
	}
}
