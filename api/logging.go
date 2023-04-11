package handler

import (
	"context"
	"golang.org/x/exp/slog"
)

type logCtxType struct{}

var logCtxKey = logCtxType{}

type logCtxHolder struct {
	logger *slog.Logger
}

func withLogger(ctx context.Context, logger *slog.Logger, args ...any) context.Context {
	v := &logCtxHolder{logger.With(args...)}
	return context.WithValue(ctx, logCtxKey, v)
}

func loggerFromContext(ctx context.Context) *slog.Logger {
	v, ok := ctx.Value(logCtxKey).(*logCtxHolder)
	if !ok {
		return nil
	}

	return v.logger
}

func loggerUpdateWith(ctx context.Context, args ...any) {
	v, ok := ctx.Value(logCtxKey).(*logCtxHolder)
	if !ok {
		return
	}

	v.logger = v.logger.With(args...)
}
