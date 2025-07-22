package sinit

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/lzap/cloudwatchwriter2"
	"github.com/osbuild/logging/pkg/logrus"
	"github.com/osbuild/logging/pkg/splunk"
	"github.com/osbuild/logging/pkg/strc"
	slogsentry "github.com/samber/slog-sentry/v2"
	journal "github.com/systemd/slog-journal"
)

// LoggingConfig is the configuration for the logging system.
type LoggingConfig struct {
	StdoutConfig StdoutConfig

	JournalConfig JournalConfig

	SplunkConfig SplunkConfig

	CloudWatchConfig CloudWatchConfig

	SentryConfig SentryConfig

	TracingConfig TracingConfig

	LogrusConfig LogrusConfig
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

// JournalConfig is the configuration for the system journal.
type JournalConfig struct {
	// Enabled is a flag to enable this output.
	Enabled bool

	// Logging level for this output. Strings "debug", "info", "warn", "error", "fatal", "panic" are accepted.
	// Keep in mind that log/slog has only 4 levels: Debug, Info, Warn, Error. Default value is "debug".
	Level string
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

// TracingConfig is the configuration for strc.
type TracingConfig struct {
	// Enabled is a flag to enable tracing
	Enabled bool

	// CustomAttrs is a list of custom static attributes to add to every log entry. To add
	// dynamic attributes, use ContextCallback that can access context.
	CustomAttrs []slog.Attr

	// ContextCallback is an optional callback function that is called for each log entry
	// to add additional attributes to the log entry.
	ContextCallback strc.MultiCallback
}

// LogrusConfig is the configuration for the logrus proxy.
type LogrusConfig struct {
	// Enabled is a flag to enable logrus proxy.
	Enabled bool

	// ExitOnFatal is a flag to enable exiting the process on fatal log entries.
	// If set to true, the process will exit with status code 1 on fatal log entries as
	// wel as panic log entries.
	ExitOnFatal bool
}

type resources struct {
	handlerMulti      *strc.MultiHandler
	handlerSplunk     *splunk.SplunkHandler
	handlerCloudWatch *cloudwatchwriter2.Handler
	sentryEnabled     bool
	prevSlogger       *slog.Logger
}

var (
	res   *resources
	resMu sync.Mutex

	ErrAlreadyInitialized   = errors.New("logging already initialized, call Close clean")
	ErrInvalidURL           = errors.New("invalid URL")
	ErrMissingURL           = errors.New("missing URL")
	ErrSentryInitialization = errors.New("sentry initialization error")
)

var osHostname = os.Hostname

// InitializeLogging initializes the logging system with the provided
// configuration. Use Close to ensure all logs are written before exiting.
// Subsequent calls to InitializeLogging will lead to ErrAlreadyInitialized.
func InitializeLogging(ctx context.Context, config LoggingConfig) error {
	resMu.Lock()
	defer resMu.Unlock()

	if res != nil {
		return ErrAlreadyInitialized
	}
	res = &resources{}

	var handlers []slog.Handler

	if err := validate(config); err != nil {
		return fmt.Errorf("logging configuration validation error: %w", err)
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

	if config.JournalConfig.Enabled {
		h, err := journal.NewHandler(&journal.Options{
			Level: parseLevel(config.JournalConfig.Level),
			ReplaceGroup: func(k string) string {
				return strings.ReplaceAll(strings.ToUpper(k), "-", "_")
			},
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				a.Key = strings.ReplaceAll(strings.ToUpper(a.Key), "-", "_")
				return a
			},
		})
		if err != nil {
			return fmt.Errorf("journal initialization error: %w", err)

		}
		handlers = append(handlers, h)
	}

	if config.SplunkConfig.Enabled {
		if config.SplunkConfig.Hostname == "" {
			hostname, err := osHostname()
			if err != nil {
				return fmt.Errorf("failed to get hostname: %w", err)
			}

			config.SplunkConfig.Hostname = hostname
		}

		c := splunk.SplunkConfig{
			Level:    parseLevel(config.SplunkConfig.Level),
			URL:      config.SplunkConfig.URL,
			Token:    config.SplunkConfig.Token,
			Source:   config.SplunkConfig.Source,
			Hostname: config.SplunkConfig.Hostname,
		}
		res.handlerSplunk = splunk.NewSplunkHandler(ctx, c)
		handlers = append(handlers, res.handlerSplunk)
	}

	if config.CloudWatchConfig.Enabled {
		var err error
		res.handlerCloudWatch, err = cloudwatchwriter2.NewHandler(cloudwatchwriter2.HandlerConfig{
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
			return fmt.Errorf("cloudwatch initialization error: %w", err)

		}
		handlers = append(handlers, res.handlerCloudWatch)
	}

	if config.SentryConfig.Enabled {
		res.sentryEnabled = true

		err := sentry.Init(sentry.ClientOptions{
			Dsn:           config.SentryConfig.DSN,
			EnableTracing: false,
		})
		if err != nil {
			return fmt.Errorf("%w: %w", ErrSentryInitialization, err)

		}

		h := slogsentry.Option{
			Level:     slog.LevelError,
			AddSource: true,
		}.NewSentryHandler()
		handlers = append(handlers, h)
	}

	// create the combined handler
	res.handlerMulti = strc.NewMultiHandlerCustom(
		config.TracingConfig.CustomAttrs,
		config.TracingConfig.ContextCallback,
		handlers...,
	)

	// configure slog
	res.prevSlogger = slog.Default()
	logger := slog.New(res.handlerMulti)
	slog.SetDefault(logger)

	// configure tracing
	if config.TracingConfig.Enabled {
		strc.SetLogger(logger)
	}

	// configure logrus proxy
	if config.LogrusConfig.Enabled {
		logrus.SetDefault(logrus.NewProxyFor(logger, logrus.Options{
			NoExit: !config.LogrusConfig.ExitOnFatal,
		}))
	}

	return nil
}

// StdLogger returns a standard library legacy logger that writes to configured
// outputs. This is only useful for passing to libraries that require a legacy
// Go standard logger.
//
// If logging was not initialized, it returns a logger that writes to
// io.Discard.
func StdLogger() *log.Logger {
	resMu.Lock()
	defer resMu.Unlock()

	if res == nil {
		return slog.NewLogLogger(slog.NewJSONHandler(io.Discard, nil), slog.LevelDebug)
	}

	// write via debug level, handlers will filter out messages below their level
	return slog.NewLogLogger(res.handlerMulti, slog.LevelDebug)
}

func validate(config LoggingConfig) error {
	if config.SplunkConfig.Enabled {
		if config.SplunkConfig.URL == "" {
			return fmt.Errorf("%w: splunk URL is required", ErrMissingURL)
		}

		_, err := url.Parse(config.SplunkConfig.URL)
		if err != nil {
			return fmt.Errorf("%w: '%s': %w", ErrInvalidURL, config.SplunkConfig.URL, err)
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
