package logger

import (
	"log/slog"
	"os"
	"strings"
)

// Logger wraps slog for structured, leveled logging.
type Logger struct {
	inner *slog.Logger
	level slog.Level
}

// New creates a new Logger at the given level ("debug", "info", "warn", "error").
func New(level string) *Logger {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "info":
		l = slog.LevelInfo
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: l}
	handler := slog.NewJSONHandler(os.Stdout, opts)

	return &Logger{
		inner: slog.New(handler),
		level: l,
	}
}

func (l *Logger) Debug(msg string, args ...interface{}) { l.inner.Debug(msg, args...) }
func (l *Logger) Info(msg string, args ...interface{})  { l.inner.Info(msg, args...) }
func (l *Logger) Warn(msg string, args ...interface{})  { l.inner.Warn(msg, args...) }
func (l *Logger) Error(msg string, args ...interface{}) { l.inner.Error(msg, args...) }

// With returns a child logger with added attributes.
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{inner: l.inner.With(args...), level: l.level}
}
