package sinit

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestInitAgain(t *testing.T) {
	ctx := context.Background()
	cfg := LoggingConfig{
		StdoutConfig: StdoutConfig{
			Enabled: true,
		},
	}
	// XX: this needs to become a real test that capture
	// stdout and compares what is written, its just for
	// demo purposes
	err := InitializeLogging(ctx, cfg)
	assert.NoError(t, err)
	slog.Info("stdout enabled")
	cfg.StdoutConfig.Enabled = false
	err = InitializeLogging(ctx, cfg)
	assert.NoError(t, err)
	slog.Info("stdout no longer enabled")
	cfg.StdoutConfig.Enabled = true
	err = InitializeLogging(ctx, cfg)
	assert.NoError(t, err)
	slog.Info("stdout and enabled again")
}
