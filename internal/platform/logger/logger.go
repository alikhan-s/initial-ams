package logger

import (
	"log/slog"
	"os"
)

// New initializes a configured slog.Logger.
// env: "prod" (JSON) or "dev" (Text, colored/human readable)
// level: "DEBUG", "INFO", "WARN", "ERROR"
func New(env string, levelStr string) *slog.Logger {
	var handler slog.Handler

	// Parse level string to slog.Level
	var level slog.Level
	switch levelStr {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: level == slog.LevelDebug, // Only show file:line in DEBUG
	}

	if env == "prod" {
		// JSON for machine parsing (Production)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text for human readability (Development)
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)

	// Set as global default so we can use slog.Info() directly if needed
	slog.SetDefault(logger)

	return logger
}
