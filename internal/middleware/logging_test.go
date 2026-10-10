package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggerRecordsRequestDetails(t *testing.T) {
	logs := captureLogs(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("hello"))
	})

	req := httptest.NewRequest(http.MethodPost, "/orders", nil)
	req.Header.Set("User-Agent", "gatekeep-test")

	Chain(next, RequestID, Logger).ServeHTTP(httptest.NewRecorder(), req)

	var entry map[string]any
	if err := json.Unmarshal(logs.Bytes(), &entry); err != nil {
		t.Fatalf("log line is not valid JSON: %v\n%s", err, logs.String())
	}

	want := map[string]any{
		"msg":        "request",
		"method":     http.MethodPost,
		"path":       "/orders",
		"status":     float64(http.StatusTeapot), // JSON numbers decode as float64
		"bytes":      float64(5),
		"user_agent": "gatekeep-test",
	}
	for k, v := range want {
		if entry[k] != v {
			t.Errorf("log field %q = %v, want %v", k, entry[k], v)
		}
	}

	if id, _ := entry["request_id"].(string); len(id) != 32 {
		t.Errorf("log field request_id = %q, want 32 hex characters", id)
	}
	if d, _ := entry["duration_ms"].(float64); d < 0 {
		t.Errorf("log field duration_ms = %v, want >= 0", d)
	}
}

func TestResponseRecorderDefaultsToStatus200(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// No WriteHeader call: net/http implies 200.
		_, _ = w.Write([]byte("ok"))
	})

	rec := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	next.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.status != http.StatusOK {
		t.Errorf("recorded status = %d, want %d", rec.status, http.StatusOK)
	}
	if rec.bytes != 2 {
		t.Errorf("recorded bytes = %d, want 2", rec.bytes)
	}
}

func TestResponseRecorderKeepsFirstStatus(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.WriteHeader(http.StatusInternalServerError) // superfluous; net/http ignores it
	})

	inner := httptest.NewRecorder()
	rec := &responseRecorder{ResponseWriter: inner}
	next.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.status != http.StatusNotFound {
		t.Errorf("recorded status = %d, want %d (the first one written)", rec.status, http.StatusNotFound)
	}
	if inner.Code != http.StatusNotFound {
		t.Errorf("client saw status %d, want %d", inner.Code, http.StatusNotFound)
	}
}

func TestResponseRecorderStaysFlushable(t *testing.T) {
	var flushable bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The wrapper does not implement http.Flusher itself; ResponseController
		// must reach the real writer through Unwrap.
		flushable = http.NewResponseController(w).Flush() == nil
	})

	rec := &responseRecorder{ResponseWriter: httptest.NewRecorder()}
	next.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if !flushable {
		t.Error("Flush through the wrapper failed; streaming responses would buffer")
	}
}
