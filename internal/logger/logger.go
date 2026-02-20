package logger

import (
	"context"
	"log/slog"
	"os"
	"time"
)

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

var (
	logger *slog.Logger
)

type contextKey string

const loggerContextKey contextKey = "logger"

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
}

func Setup(level Level, jsonOutput bool) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: getLevel(level),
	}

	if jsonOutput {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger = slog.New(handler)
	slog.SetDefault(logger)
}

func getLevel(level Level) slog.Level {
	switch level {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func DebugContext(ctx context.Context, msg string, args ...any) {
	logger.DebugContext(ctx, msg, args...)
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func InfoContext(ctx context.Context, msg string, args ...any) {
	logger.InfoContext(ctx, msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}

func WarnContext(ctx context.Context, msg string, args ...any) {
	logger.WarnContext(ctx, msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func ErrorContext(ctx context.Context, msg string, args ...any) {
	logger.ErrorContext(ctx, msg, args...)
}

func With(args ...any) *slog.Logger {
	return logger.With(args...)
}

func WithContext(ctx context.Context, _ ...any) context.Context {
	return context.WithValue(ctx, loggerContextKey, logger)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerContextKey).(*slog.Logger); ok {
		return l
	}
	return logger
}

type LogTimer struct {
	start  time.Time
	logger *slog.Logger
	msg    string
	args   []any
}

func StartTimer(msg string, args ...any) *LogTimer {
	return &LogTimer{
		start:  time.Now(),
		logger: logger,
		msg:    msg,
		args:   args,
	}
}

func (t *LogTimer) End() {
	t.logger.Info(t.msg, append(t.args, "duration", time.Since(t.start))...)
}
