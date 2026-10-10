package middleware

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// captureLogs redirects the default slog logger into a buffer for the duration
// of the test, and restores the previous logger afterwards. It keeps expected
// error logs out of the test output and lets us assert on what was logged.
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var buf bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	return &buf
}

func TestRecoverConvertsPanicTo500(t *testing.T) {
	logs := captureLogs(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)

	// Wrapped in RequestID so the recovered log line has an ID to report.
	Chain(next, RequestID, Recover).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	logged := logs.String()
	for _, want := range []string{"panic recovered", "boom", "/panic", "request_id"} {
		if !strings.Contains(logged, want) {
			t.Errorf("log output missing %q; got:\n%s", want, logged)
		}
	}
}

func TestRecoverLeavesNormalResponsesAlone(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("fine"))
	})

	rec := httptest.NewRecorder()
	Recover(next).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusTeapot)
	}
	if got := rec.Body.String(); got != "fine" {
		t.Errorf("body = %q, want %q", got, "fine")
	}
}

func TestRecoverRepanicsErrAbortHandler(t *testing.T) {
	defer func() {
		got := recover()
		if got == nil {
			t.Fatal("ErrAbortHandler was swallowed, want it to propagate")
		}
		if err, ok := got.(error); !ok || !errors.Is(err, http.ErrAbortHandler) {
			t.Fatalf("recovered %v, want http.ErrAbortHandler", got)
		}
	}()

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	})

	Recover(next).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))

	t.Fatal("ServeHTTP returned normally, want a panic")
}
