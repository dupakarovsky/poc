package foundationhttp

import (
	"net/http"
	"slices"
)

// Middleware wraps an http.Handler with additional behavior — logging,
// panic recovery, auth, rate-limiting, etc. — and returns a new
// http.Handler that runs the added behavior around (or instead of) the
// original. LoggingMiddleware already had exactly this shape before this
// file existed; naming it as a type just lets us store a slice of them
// instead of hardcoding one-off wrapping by hand.
type Middleware func(http.Handler) http.Handler

// Router wraps http.ServeMux with two things ServeMux doesn't offer on its
// own: a global middleware stack applied to every request regardless of
// which pattern matched (or whether any pattern matched at all), and the
// ability to attach per-route middleware at the point a specific route is
// registered. Every application service's handler (ReservationHandler
// today, others later) registers its own routes onto a shared *Router
// instance built once in main, rather than main knowing the URL patterns
// itself.
type Router struct {
	mux    *http.ServeMux
	global []Middleware
}

// NewRouter constructs an empty Router ready for routes and global
// middleware to be attached, in either order — Use and Handle can be
// called in any sequence before Build is finally called once, at startup.
func NewRouter() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// Use appends mw to the global middleware stack. Global middleware wraps
// every request handled by this Router, including requests that match no
// registered pattern at all (ServeMux's own built-in 404) — since it wraps
// the whole mux, not any individual route.
//
// Order matters: middleware passed to Use earlier ends up outermost in the
// final chain — the first thing that runs on the way in, the last thing
// that runs on the way out. Call Use(Logging, Recovery) — in that order —
// if you want Logging to observe (and log) even a request that Recovery
// had to step in for.
func (r *Router) Use(mw ...Middleware) {
	r.global = append(r.global, mw...)
}

// Handle registers handler for pattern, optionally wrapped in its own
// route-specific middleware before the global stack ever gets involved.
// Same outermost-first ordering convention as Use, kept consistent on
// purpose so there's only one rule to remember for both cases.
func (r *Router) Handle(pattern string, handler http.HandlerFunc, mw ...Middleware) {

	var h http.Handler = handler

	for _, m := range slices.Backward(mw) {
		if m != nil {
			h = m(h)
		}
	}

	r.mux.Handle(pattern, h)
}

// Build finalizes the Router into a single http.Handler ready to hand to
// http.Server, wrapping the underlying mux (with every route's own
// middleware already applied) in the global middleware stack. Called
// exactly once, after every RegisterRoutes call has finished attaching its
// routes — any Use or Handle call made after Build won't be reflected in
// the handler Build already returned.
func (r *Router) Build() http.Handler {
	var h http.Handler = r.mux

	for _, m := range slices.Backward(r.global) {
		if m != nil {
			h = m(h)
		}
	}

	return h
}
