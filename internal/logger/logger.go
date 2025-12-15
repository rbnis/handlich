package logger

import (
	"log/slog"
	"os"
)

var log *slog.Logger

func init() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	log = slog.New(handler)
}

func Get() *slog.Logger {
	return log
}

func SetLevel(level slog.Level) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	log = slog.New(handler)
}
