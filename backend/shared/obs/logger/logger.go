package logger

import (
	"log/slog"
	"os"
	"shared/config"
	"time"
)

func NewLogger(cfg config.Config, service, version string) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.TimeValue(t.UTC())
				}
			}
			return a
		},
	}

	// if cfg.Env == "development" {
	// 	opts.Level = slog.LevelDebug
	// 	handler = slog.NewTextHandler(os.Stdout, opts)
	// } else {
	// 	handler = slog.NewJSONHandler(os.Stdout, opts)
	// }

	handler = slog.NewJSONHandler(os.Stdout, opts)

	return slog.New(handler).With(
		"service", service,
		"version", version,
	)
}
