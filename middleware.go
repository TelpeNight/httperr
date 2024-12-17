package httperr

import (
	"net/http"
)

type M = func(http.Handler) http.Handler

// Middleware is the simplest middleware that wraps incoming [http.ResponseWriter] and passes it to the next handler
func Middleware() M {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(ResponseWriterWithErrors(w), r)
		})
	}
}

// EnsureMiddleware is a helper for debugging.
// Since not all possible middlewares in the stack may support 1.22 unwrapping.
// Will log such situation with logger and optional stacktrace.
func EnsureMiddleware(ops ...EnsureOp) M {
	options := newEnsureOp(ops)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if unwrap(w) == nil {
				options.warn(w)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// WithMiddleware is a helper function to apply middlewares to handler
func WithMiddleware(h http.Handler, m ...M) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}
