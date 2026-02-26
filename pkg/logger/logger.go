package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Init creates and returns a configured *slog.Logger based on the environment.
//
//   - env="production"  → JSON handler (machine-readable, for CI/CD)
//   - env="development" → Text handler (human-readable, for local dev)
//
// The logger includes default attributes for component identification.
func Init(env string) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	switch strings.ToLower(strings.TrimSpace(env)) {
	case "production":
		handler = slog.NewJSONHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler).With(
		slog.String("component", "job-scraper"),
	)

	// Set as the global default so slog.Info(), slog.Error() etc. use it.
	slog.SetDefault(logger)

	return logger
}
