package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New initializes the structured logger based on environment and log level.
// It configures JSON handling for production and Text handling for development.
func New(appEnv string, level string) *slog.Logger {
	var handler slog.Handler

	var slogLevel slog.Level
	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "info":
		slogLevel = slog.LevelInfo
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	if appEnv == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)

	// Set as global default so slog.Info(), slog.Error() use our configuration natively
	slog.SetDefault(logger)

	return logger
}
