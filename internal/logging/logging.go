// Package logging configures the application's structured logger.
package logging

import (
	"log/slog"
	"os"
)

// New builds a structured JSON logger at the given level ("debug", "info", "warn", "error").
// An unrecognized level falls back to info.
func New(level string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(level),
	})

	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
