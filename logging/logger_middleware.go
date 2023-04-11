package logging

import (
	"golang.org/x/exp/slog"
	"net/http"
	"os"
	"time"
)

type responseData struct {
	status int
	size   int
}

// our http.ResponseWriter implementation
type loggingResponseWriter struct {
	http.ResponseWriter // compose original http.ResponseWriter
	responseData        *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b) // write response using original http.ResponseWriter
	r.responseData.size += size            // capture size
	if r.responseData.status == 0 {
		r.responseData.status = http.StatusOK // if status code was not set, set it to 200
	}
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode) // write status code using original http.ResponseWriter
	r.responseData.status = statusCode       // capture status code
}

// GetIP gets a requests IP address by reading off the forwarded-for
// header (for proxies) and falls back to use the remote address.
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func LoggerMiddleware(fn http.Handler, logger *slog.Logger) http.Handler {
	var baseLogger *slog.Logger
	if logger == nil {
		baseLogger = slog.New(slog.NewTextHandler(os.Stdout))
	} else {
		baseLogger = logger
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ctx := WithLogger(
			r.Context(),
			baseLogger,
			"method", r.Method,
			"url", r.URL.String(),
			"ip", GetIP(r),
		)

		if ref := r.Header.Get("Referer"); ref != "" {
			LoggerUpdateWith(ctx, "referer", ref)
		}

		rw := &loggingResponseWriter{w, &responseData{}}

		defer func() {
			// log panics
			if err := recover(); err != nil {
				LoggerFromContext(ctx).ErrorCtx(
					ctx,
					"request panic",
					"err", err,
				)

				// keep panicking
				panic(err)
			}
			responseData := rw.responseData
			duration := time.Since(start)

			LoggerFromContext(ctx).InfoCtx(
				ctx,
				"request",
				"status", responseData.status,
				"size", responseData.size,
				"duration", duration,
			)
		}()

		fn.ServeHTTP(rw, r.WithContext(ctx))
	})
}
