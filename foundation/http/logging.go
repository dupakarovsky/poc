package foundationhttp

import (
	"log/slog"
	"net/http"
	"time"
)

// responseWriterWrapper intercepts WriteHeader to capture the status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriterWrapper(w http.ResponseWriter) *responseWriterWrapper {
	// Default to 200, since WriteHeader is only called explicitly for
	// non-200 responses.
	return &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}
}

// WriteHeader first captures the statusCode in an internal field then passes it
// to internal response writer
func (rww *responseWriterWrapper) WriteHeader(code int) {
	rww.statusCode = code
	rww.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware wraps an http.Handler, logging every request and
// response at DEBUG level with method, path, status, and duration.
func LoggingMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			logger.Debug("request received",
				"protocol", r.Proto,
				"method", r.Method,
				"path", r.URL.Path,
			)

			wrapped := newResponseWriterWrapper(w)
			next.ServeHTTP(wrapped, r)

			logger.Debug("response sent",
				"protocol", r.Proto,
				"method", r.Method,
				"path", r.URL.Path,
				"status", wrapped.statusCode,
				"duration", time.Since(start).String(),
			)
		})
	}
}
