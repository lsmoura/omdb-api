package handler

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggerMiddleware(t *testing.T) {
	successHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf))

	req, err := http.NewRequest("GET", "http://localhost:8080", nil)
	require.NoError(t, err)

	var recorder httptest.ResponseRecorder
	loggerMiddleware(http.HandlerFunc(successHandler), logger).ServeHTTP(&recorder, req)

	resp := buf.String()

	assert.Contains(t, resp, "level=INFO")
	assert.Contains(t, resp, "url=http://localhost:8080")
	assert.Contains(t, resp, "msg=request")
	assert.Contains(t, resp, "status=200")
	assert.Contains(t, resp, "size=2")
}

func TestLoggerMiddlewarePanic(t *testing.T) {
	panicHandler := func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("oh no"))
	}

	panicWrapper := func(fn http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()

			fn.ServeHTTP(w, r)
		})
	}

	var buf strings.Builder
	logger := slog.New(slog.NewTextHandler(&buf))

	req, err := http.NewRequest("GET", "http://localhost:8080", nil)
	require.NoError(t, err)

	var recorder httptest.ResponseRecorder
	panicWrapper(loggerMiddleware(http.HandlerFunc(panicHandler), logger)).ServeHTTP(&recorder, req)

	resp := buf.String()

	assert.Contains(t, resp, "level=ERROR")
	assert.Contains(t, resp, "url=http://localhost:8080")
	assert.Contains(t, resp, "msg=\"request panic\"")
	assert.Contains(t, resp, "err=\"oh no\"")
}
