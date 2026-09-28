package logger

import (
	"log/slog"
	"os"
	"shared/config"
)

func NewLogger(cfg config.Config, service, version string) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	if cfg.Env == "development" {
		opts.Level = slog.LevelDebug
		opts.AddSource = true
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(
		"service", service,
		"version", version,
	)
}
