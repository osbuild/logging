package sinit

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func init() {
	osHostname = func() (string, error) {
		return "default-hostname", nil
	}
}

func TestValidationEmpty(t *testing.T) {
	cfg := LoggingConfig{}

	if err := validate(cfg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidationSplunkURLValid(t *testing.T) {
	cfg := LoggingConfig{
		SplunkConfig: SplunkConfig{
			Enabled: true,
			URL:     "https://splunk.example.com",
		},
	}

	if err := validate(cfg); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidationSplunkURLInvalid(t *testing.T) {
	cfg := LoggingConfig{
		SplunkConfig: SplunkConfig{
			Enabled: true,
			URL:     "%zzzzz",
		},
	}

	if err := validate(cfg); !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL error, got %v", err)
	}
}

func TestEmptyConfiguration(t *testing.T) {
	ctx := context.Background()
	cfg := LoggingConfig{}

	err := InitializeLogging(ctx, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err = Close(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidationSplunkEmptyURL(t *testing.T) {
	cfg := LoggingConfig{
		SplunkConfig: SplunkConfig{
			Enabled: true,
		},
	}

	if err := validate(cfg); !errors.Is(err, ErrMissingURL) {
		t.Fatalf("expected ErrMissingURL error, got %v", err)
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
		if err := Close(100 * time.Millisecond); err != nil {
			t.Fatalf("expected no error on close, got %v", err)
		}
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
		if err := Close(100 * time.Millisecond); err != nil {
			t.Fatalf("expected no error on close, got %v", err)
		}
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
