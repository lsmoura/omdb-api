package handler

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slog"
	"math/rand"
	"strings"
	"testing"
)

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func TestLoggingBasic(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf))

	ctx := context.Background()

	logger.InfoCtx(ctx, "hello world")

	resp := buf.String()

	assert.Contains(t, resp, "level=INFO")
	assert.Contains(t, resp, "msg=\"hello world\"")
}

func TestLoggingWith(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf))

	ctx := context.Background()

	logger = logger.With("foo", "bar")

	logger.InfoCtx(ctx, "hello world")

	resp := buf.String()

	assert.Contains(t, resp, "level=INFO")
	assert.Contains(t, resp, "msg=\"hello world\"")
	assert.Contains(t, resp, "foo=bar")
}

func TestLoggingCtx(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf)).With("foo", "bar")

	ctx := context.Background()
	ctx = withLogger(ctx, logger)

	ctxLogger := loggerFromContext(ctx)

	ctxLogger.InfoCtx(ctx, "hello world")

	resp := buf.String()

	assert.Contains(t, resp, "level=INFO")
	assert.Contains(t, resp, "msg=\"hello world\"")
	assert.Contains(t, resp, "foo=bar")
}

func TestLoggingCtxUpdate(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf))

	ctx := context.Background()
	ctx = withLogger(ctx, logger)

	key := randomString(10)
	value := randomString(10)

	loggerUpdateWith(ctx, key, value)

	ctxLogger := loggerFromContext(ctx)

	ctxLogger.InfoCtx(ctx, "hello world")

	resp := buf.String()

	assert.Contains(t, resp, "level=INFO")
	assert.Contains(t, resp, "msg=\"hello world\"")
	assert.Contains(t, resp, fmt.Sprintf("%s=%s", key, value))

	// confirms that the initial logger remains untouched
	buf.Reset()
	logger.InfoCtx(ctx, "hello world")
	resp = buf.String()

	assert.Contains(t, resp, "level=INFO")
	assert.Contains(t, resp, "msg=\"hello world\"")
	assert.NotContains(t, resp, key)
	assert.NotContains(t, resp, value)
}
