package sinit

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func init() {
	hostnameFunc = func() (string, error) {
		return "default-hostname", nil
	}
}

func TestValidationSplunkWithHostname(t *testing.T) {
	ch := make(chan string)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		ch <- string(body)
	}))
	defer srv.Close()

	ctx := context.Background()
	cfg := LoggingConfig{
		SplunkConfig: SplunkConfig{
			Enabled:  true,
			URL:      srv.URL,
			Hostname: "some-test-hostname",
		},
	}

	err := InitializeLogging(ctx, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	defer func() {
		// XXX: uncomment once https://github.com/osbuild/logging/pull/31 is merged
		// if err := Close(100 * time.Millisecond); err != nil {
		// 	t.Fatalf("expected no error on close, got %v", err)
		// }
	}()

	slog.Warn("foo")
	Flush()

	select {
	case splunkBody := <-ch:
		if !strings.Contains(splunkBody, `"host":"some-test-hostname"`) {
			t.Fatalf("expected hostname to be set in splunk body, got %s", splunkBody)
		}
	case <-time.After(6 * time.Second):
		panic("no splunk record in 6s")
	}
}

func TestValidationSplunkWithoutHostname(t *testing.T) {
	ch := make(chan string)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		ch <- string(body)
	}))
	defer srv.Close()

	ctx := context.Background()
	cfg := LoggingConfig{
		SplunkConfig: SplunkConfig{
			Enabled: true,
			URL:     srv.URL,
		},
	}

	err := InitializeLogging(ctx, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	defer func() {
		// XXX: uncomment once https://github.com/osbuild/logging/pull/31 is merged
		// if err := Close(100 * time.Millisecond); err != nil {
		// 	t.Fatalf("expected no error on close, got %v", err)
		// }
	}()

	slog.Warn("foo")
	Flush()

	select {
	case splunkBody := <-ch:
		if !strings.Contains(splunkBody, `"host":"default-hostname"`) {
			t.Fatalf("expected hostname to be set in splunk body, got %s", splunkBody)
		}
	case <-time.After(6 * time.Second):
		panic("no splunk record in 6s")
	}
}
