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

	go Close(50 * time.Millisecond)
	go Close(50 * time.Millisecond)
}
