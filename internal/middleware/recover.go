package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recover turns a panic in a downstream handler into a 500 response and a
// structured log line, instead of a silently dropped connection.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			// net/http uses this sentinel to abort a response on purpose.
			// Swallowing it would break that contract.
			if err, ok := rec.(error); ok && errors.Is(err, http.ErrAbortHandler) {
				panic(rec)
			}

			slog.Error("panic recovered",
				"err", rec,
				"request_id", RequestIDFrom(r.Context()),
				"method", r.Method,
				"path", r.URL.Path,
				"stack", string(debug.Stack()),
			)

			w.WriteHeader(http.StatusInternalServerError)
		}()

		next.ServeHTTP(w, r)
	})
}
