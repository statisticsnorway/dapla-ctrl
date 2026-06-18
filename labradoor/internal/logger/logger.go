package logger

import (
	"context"
	"log/slog"
)

type contextKey string // context.WithValue want us to use custom type
const loggerKey contextKey = "slog_logger"

func LoggerFromCtx(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func CtxWithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, l)
}
