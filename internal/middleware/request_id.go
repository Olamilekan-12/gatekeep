package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const HeaderRequestID = "X-Request-ID"

// ctxKey is unexported so no other package can collide with our context keys.
type ctxKey int

const reqquestIDKey ctxKey = iota

// RequestID attaches a unique ID to every request, reusing an inbound
// X-Request-ID if the caller supplied one.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(HeaderRequestID)
		if id == "" {
			id = newRequestID()
		}
		w.Header().Set(HeaderRequestID, id)
		ctx := context.WithValue(r.Context(), reqquestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequestIDFrom returns the request ID stored in ctx, or "" if there is none.
func RequestIDFrom(ctx context.Context) string {
	id, _ := ctx.Value(reqquestIDKey).(string)
	return id
}

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
