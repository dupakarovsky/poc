package foundationhttp

import (
	"context"
	"net/http"
	"time"
)

// TimeoutMiddleware wraps next so that every request downstream — the
// handler itself, and anything it calls (the application-layer use cases,
// and through them, the repository's database calls) — receives a context
// bounded by timeout.
func TimeoutMiddleware(timeout time.Duration) Middleware {

	m := func(next http.Handler) http.Handler {

		h := func(w http.ResponseWriter, r *http.Request) {

			// create the timed context
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			// creates a new http request object with the timed context
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(h)
	}

	return m
}
