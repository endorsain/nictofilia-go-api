// Package logging configures the application-wide structured logger.
package logging

import (
	"log/slog"
	"os"
	"strings"

	"github.com/endorsain/nictofilia-go-api/internal/config"
)

// New builds the application's structured logger (JSON) from configuration.
func New(cfg config.LoggingConfig) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(cfg.Level),
	})
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
