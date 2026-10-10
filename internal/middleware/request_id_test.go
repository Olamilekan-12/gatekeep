package middleware

import (
	"context"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDGeneratesWhenAbsent(t *testing.T) {
	var seenInHandler string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenInHandler = RequestIDFrom(r.Context())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	RequestID(next).ServeHTTP(rec, req)

	if _, err := hex.DecodeString(seenInHandler); err != nil || len(seenInHandler) != 32 {
		t.Errorf("request ID = %q, want 32 hex characters", seenInHandler)
	}
	if got := rec.Header().Get(HeaderRequestID); got != seenInHandler {
		t.Errorf("response header %s = %q, want %q", HeaderRequestID, got, seenInHandler)
	}
}

func TestRequestIDReusesInboundHeader(t *testing.T) {
	const want = "upstream-id-123"

	var seenInHandler string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenInHandler = RequestIDFrom(r.Context())
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set(HeaderRequestID, want)

	RequestID(next).ServeHTTP(rec, req)

	if seenInHandler != want {
		t.Errorf("request ID in context = %q, want %q", seenInHandler, want)
	}
	if got := rec.Header().Get(HeaderRequestID); got != want {
		t.Errorf("response header %s = %q, want %q", HeaderRequestID, got, want)
	}
}

func TestRequestIDsAreUnique(t *testing.T) {
	const n = 1000

	seen := make(map[string]struct{}, n)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[RequestIDFrom(r.Context())] = struct{}{}
	})

	handler := RequestID(next)
	for range n {
		handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	}

	if len(seen) != n {
		t.Errorf("got %d distinct IDs across %d requests, want %d", len(seen), n, n)
	}
}

func TestRequestIDFromReturnsEmptyWhenAbsent(t *testing.T) {
	if got := RequestIDFrom(context.Background()); got != "" {
		t.Errorf("RequestIDFrom(empty context) = %q, want %q", got, "")
	}
}
