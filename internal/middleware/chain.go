package middleware

import "net/http"

// Middleware wraps an http.Handler with additional behaviour.
type Middleware func(http.Handler) http.Handler

// Chain composes middleware around h so that the first argument is the
// outermost layer. Chain(h, A, B, C) is equivalent to A(B(C(h))), which means
// the list reads in the order a request travels through it.
func Chain(h http.Handler, mw ...Middleware) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}
