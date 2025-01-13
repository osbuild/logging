package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/osbuild/logging/pkg/sinit"
	"github.com/osbuild/logging/pkg/strc"
)

func main() {
	cfg := sinit.LoggingConfig{
		StdoutConfig: sinit.StdoutConfig{
			Enabled: true,
			Level:   "debug",
			Format:  "text",
		},
		JournalConfig: sinit.JournalConfig{
			Enabled: true,
			Level:   "debug",
		},
		SplunkConfig: sinit.SplunkConfig{
			Enabled:  os.Getenv("SPLUNK_URL") != "",
			Level:    "debug",
			URL:      os.Getenv("SPLUNK_URL"),
			Token:    os.Getenv("SPLUNK_TOKEN"),
			Source:   "test-source",
			Hostname: "test-hostname",
		},
		CloudWatchConfig: sinit.CloudWatchConfig{
			Enabled:      os.Getenv("AWS_REGION") != "",
			Level:        "debug",
			AWSRegion:    os.Getenv("AWS_REGION"),
			AWSSecret:    os.Getenv("AWS_SECRET"),
			AWSKey:       os.Getenv("AWS_KEY"),
			AWSLogGroup:  "test-group",
			AWSLogStream: "test-stream",
		},
		SentryConfig: sinit.SentryConfig{
			Enabled: true,
			DSN:     os.Getenv("SENTRY_DSN"),
		},
		TracingConfig: sinit.TracingConfig{
			Enabled: true,
		},
	}
	err := sinit.InitializeLogging(context.Background(), cfg)
	if err != nil {
		panic(err)
	}
	defer sinit.Flush()

	span, ctx := strc.Start(context.Background(), "main")

	slog.DebugContext(ctx, "message",
		slog.Bool("b1", true),
		slog.Int("i1", 1),
		slog.Float64("f1", 1.1),
		slog.String("s1", "v1"),
		slog.Time("t1", time.Now()),
		slog.Duration("d1", 1*time.Second),
	)

	slog.ErrorContext(ctx, "an error occured", "err", "this is an error")

	span.End()
}
