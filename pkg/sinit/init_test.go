package sinit

import (
	"context"
	"errors"
	"testing"
	"time"
)

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
