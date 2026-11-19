package foundationhttp

import (
	"log/slog"
	"net/http"
)

// RecoveryMiddleware wraps next so that a panic anywhere inside it — the
// handler itself, or any middleware layered between this one and the
// handler — is caught, logged with a stack trace, and turned into a
// 500 response, instead of crashing the goroutine handling this request.
// net/http's own server already recovers per-connection panics on its own,
// but silently: no response body reaches the client, and nothing surfaces
// in our structured logger. This middleware exists specifically to make
// that failure mode visible and give the client a real (if generic)
// response instead of a reset connection.
func RecoveryMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						"error", rec,
						"method", r.Method,
						"path", r.URL.Path,
					)

					// force this connection closed rather than returned to
					// the keep-alive pool: a panic mid-handler means we can
					// no longer be fully confident in this connection's
					// state, so the safer default is a fresh connection for
					// whatever request comes next. Must be set before
					// http.Error, since http.Error calls WriteHeader
					// internally — and WriteHeader freezes the header map,
					// silently ignoring any Set call made after it.
					w.Header().Set("Connection", "close")
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
