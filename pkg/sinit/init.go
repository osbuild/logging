package sinit

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/osbuild/logging/pkg/cloudwatch"
	"github.com/osbuild/logging/pkg/logrus"
	"github.com/osbuild/logging/pkg/splunk"
	"github.com/osbuild/logging/pkg/strc"

	"github.com/getsentry/sentry-go"
	slogsentry "github.com/samber/slog-sentry/v2"
)

// LoggingConfig is the configuration for the logging system.
type LoggingConfig struct {
	StdoutConfig StdoutConfig

	SplunkConfig SplunkConfig

	CloudWatchConfig CloudWatchConfig

	SentryConfig SentryConfig
}

// StdoutConfig is the configuration for the standard output.
type StdoutConfig struct {
	// Enabled is a flag to enable this output.
	Enabled bool

	// Logging level for this output. Strings "debug", "info", "warn", "error", "fatal", "panic" are accepted.
	// Keep in mind that log/slog has only 4 levels: Debug, Info, Warn, Error. Default value is "debug".
	Level string

	// Format is the log format to use for stdout logging. Possible values are "json" and "text".
	Format string
}

// SplunkConfig is the configuration for the Splunk output.
type SplunkConfig struct {
	// Enabled is a flag to enable this output.
	Enabled bool

	// Logging level for this output. Strings "debug", "info", "warn", "error", "fatal", "panic" are accepted.
	// Keep in mind that log/slog has only 4 levels: Debug, Info, Warn, Error. Default value is "debug".
	Level string

	// URL is the Splunk HEC URL.
	URL string

	// Token is the Splunk HEC token.
	Token string

	// Source is the Splunk HEC source.
	Source string

	// Hostname is the Splunk HEC hostname.
	Hostname string
}

// SentryConfig is the configuration for the Sentry output. Only log entries with error level are sent to Sentry.
type SentryConfig struct {
	// Enabled is a flag to enable Sentry.
	Enabled bool

	// DSN is the Sentry DSN.
	DSN string
}

// CloudWatchConfig is the configuration for the CloudWatch output.
type CloudWatchConfig struct {
	// Enabled is a flag to enable this output.
	Enabled bool

	// Logging level for this output. Strings "debug", "info", "warn", "error", "fatal", "panic" are accepted.
	// Keep in mind that log/slog has only 4 levels: Debug, Info, Warn, Error. Default value is "debug".
	Level string

	// AWSRegion is the AWS region.
	AWSRegion string

	// AWSKey is the AWS access key.
	AWSKey string

	// AWSSecret is the AWS secret key.
	AWSSecret string

	// AWSSession is an optional AWS session token.
	AWSSession string

	// AWSLogGroup is the AWS CloudWatch log group.
	AWSLogGroup string

	// AWSLogStream is the AWS CloudWatch log stream.
	AWSLogStream string
}

var initOnce sync.Once
var handlerMulti *strc.MultiHandler
var handlerSplunk *splunk.SplunkHandler
var handlerCloudWatch *cloudwatch.CloudwatchHandler

// InitializeLogging initializes the logging system with the provided configuration. Use Flush to ensure all logs are written before exiting.
// Subsequent calls to InitializeLogging will have no effect and will not return any error.
func InitializeLogging(ctx context.Context, config LoggingConfig) error {
	var outerError error

	initOnce.Do(func() {
		var handlers []slog.Handler

		if err := validate(config); err != nil {
			outerError = fmt.Errorf("logging configuration validation error: %w", err)
			return
		}

		if config.StdoutConfig.Enabled {
			var h slog.Handler
			opts := &slog.HandlerOptions{
				Level: parseLevel(config.StdoutConfig.Level),
			}
			if strings.EqualFold(config.StdoutConfig.Format, "json") {
				h = slog.NewJSONHandler(os.Stdout, opts)
			} else {
				h = slog.NewTextHandler(os.Stdout, opts)
			}
			handlers = append(handlers, h)
		}

		if config.SplunkConfig.Enabled {
			handlerSplunk = splunk.NewSplunkHandler(ctx,
				parseLevel(config.SplunkConfig.Level),
				config.SplunkConfig.URL,
				config.SplunkConfig.Token,
				config.SplunkConfig.Source,
				config.SplunkConfig.Hostname,
			)
			handlers = append(handlers, handlerSplunk)
		}

		if config.CloudWatchConfig.Enabled {
			var err error
			handlerCloudWatch, err = cloudwatch.New(cloudwatch.CloudwatchConfig{
				Level:        parseLevel(config.CloudWatchConfig.Level),
				AddSource:    true,
				AWSRegion:    config.CloudWatchConfig.AWSRegion,
				AWSKey:       config.CloudWatchConfig.AWSKey,
				AWSSecret:    config.CloudWatchConfig.AWSSecret,
				AWSSession:   config.CloudWatchConfig.AWSSession,
				AWSLogGroup:  config.CloudWatchConfig.AWSLogGroup,
				AWSLogStream: config.CloudWatchConfig.AWSLogStream,
			})
			if err != nil {
				outerError = fmt.Errorf("cloudwatch initialization error: %w", err)
				return
			}
			handlers = append(handlers, handlerCloudWatch)
		}

		if config.SentryConfig.Enabled {
			err := sentry.Init(sentry.ClientOptions{
				Dsn:           config.SentryConfig.DSN,
				EnableTracing: false,
			})
			if err != nil {
				outerError = fmt.Errorf("%w: %w", ErrSentryInitialization, err)
				return
			}

			h := slogsentry.Option{
				Level:     slog.LevelError,
				AddSource: true,
			}.NewSentryHandler()
			handlers = append(handlers, h)
		}

		// create the combined handler
		handlerMulti = strc.NewMultiHandler(handlers...)

		// configure slog
		logger := slog.New(handlerMulti)
		slog.SetDefault(logger)

		// configure tracing
		strc.SetLogger(logger)

		// configure logrus proxy
		logrus.SetDefault(logrus.NewProxyFor(logger))
	})

	return outerError
}

// Flush flushes all pending logs to the configured outputs. Blocks until all logs are written.
func Flush() {
	handlerSplunk.Close()
	handlerCloudWatch.Close()
	sentry.Flush(2 * time.Second)
}

// StdLogger returns a standard library legacy logger that writes to configured outputs.
// This is only useful for passing to libraries that require a legacy Go standard logger.
func StdLogger() *log.Logger {
	// write via debug level, handlers will filter out messages below their level
	return slog.NewLogLogger(handlerMulti, slog.LevelDebug)
}

var ErrInvalidURL = errors.New("invalid URL")

var ErrSentryInitialization = errors.New("sentry initialization error")

func validate(config LoggingConfig) error {
	if config.SplunkConfig.Enabled {
		_, err := url.Parse(config.SplunkConfig.URL)
		if err != nil {
			return fmt.Errorf("splunk URL '%s' is invalid: %w", config.SplunkConfig.URL, ErrInvalidURL)
		}
	}

	return nil
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug", "trace":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "fatal", "panic":
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}
