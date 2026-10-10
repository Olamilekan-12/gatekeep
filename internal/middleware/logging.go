package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"time"
)

// responseRecorder wraps an http.ResponseWriter to capture the status code and
// the number of bytes written — neither of which the standard interface exposes.
type responseRecorder struct {
	http.ResponseWriter
	status      int
	bytes       int
	wroteHeader bool
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.wroteHeader {
		return
	}
	r.status = status
	r.wroteHeader = true
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Unwrap exposes the original writer so http.ResponseController can reach
// capabilities (flushing, hijacking, deadlines) that this wrapper does not
// implement itself.
func (r *responseRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

// Logger emits one structured line per request once the handler has returned.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		slog.Info("request",
			"request_id", RequestIDFrom(r.Context()),
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"bytes", rec.bytes,
			"duration_ms", float64(time.Since(start).Nanoseconds())/1e6,
			"remote_ip", clientIP(r),
			"user_agent", r.UserAgent(),
		)
	})
}

// clientIP returns the peer address of the TCP connection. It deliberately
// ignores X-Forwarded-For: that header is attacker-controlled unless the
// request provably came through a trusted proxy.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
