package sinit

import (
	"context"
	"testing"
	"time"
)

func TestValidationSplunkFlushRacy(t *testing.T) {
	ctx := context.Background()
	cfg := LoggingConfig{
		SplunkConfig: SplunkConfig{
			Enabled: true,
			URL:     "http://example.com/splunk",
		},
		SentryConfig: SentryConfig{
			Enabled: true,
			DSN:     "https://user:pass@example.com/sentry",
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

	go Flush()
	go Flush()
}

func TestValidationSplunkCloseRacy(t *testing.T) {
	ctx := context.Background()
	cfg := LoggingConfig{
		SplunkConfig: SplunkConfig{
			Enabled: true,
			URL:     "http://example.com/splunk",
		},
		SentryConfig: SentryConfig{
			Enabled: true,
			DSN:     "https://user:pass@example.com/sentry",
		},
	}
	err := InitializeLogging(ctx, cfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := Close(100 * time.Millisecond); err != nil {
		t.Fatalf("expected no error on close, got %v", err)
	}

	if err := Close(100 * time.Millisecond); err != ErrNotInitialized {
		t.Fatalf("expected error on second close, got nil")
	}
}
