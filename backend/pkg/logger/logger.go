package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	mu     sync.RWMutex
	Logger *slog.Logger
)

// Init configures the application-wide structured logger. Production uses
// JSON output (aggregator-friendly); development uses human-readable text.
func Init(env string) {
	level := slog.LevelInfo
	if env != "production" {
		level = slog.LevelDebug
	}

	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	mu.Lock()
	Logger = slog.New(handler)
	mu.Unlock()
	slog.SetDefault(Logger)
}

// Get returns the configured logger, falling back to the default slog logger
// if Init has not been called.
func Get() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if Logger == nil {
		return slog.Default()
	}
	return Logger
}
