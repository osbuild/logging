package sinit

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/lzap/cloudwatchwriter2"
	"github.com/osbuild/logging/pkg/logrus"
	"github.com/osbuild/logging/pkg/splunk"
	"github.com/osbuild/logging/pkg/strc"
)

// Flush flushes all pending logs to the configured outputs. Depending on the
// logging configuration, it issues flush commands to various systems which
// behave differently:
//
// CloudWatch and Splunk handlers issue a flush command that has no guarantee of
// completion, meaning logs may not be flushed immediately. No blocking is
// performed.
//
// Sentry SDK flushes logs with blocking up to 2 seconds.
//
// Calling Flush without previously calling InitializeLogging will return
// ErrNotInitialized.
//
// Do not use this function during application exit, use Close() instead to
// ensure all logs are flushed properly.
func Flush() error {
	resMu.Lock()
	defer resMu.Unlock()

	if res == nil {
		return ErrNotInitialized
	}

	if res.handlerSplunk != nil {
		res.handlerSplunk.Flush()
	}

	if res.handlerCloudWatch != nil {
		res.handlerCloudWatch.Flush()
	}

	sentry.Flush(2 * time.Second)

	return nil
}

var (
	ErrTimeoutDuringClose = errors.New("timeout during close")
	ErrSentryTimeout      = errors.New("sentry timeout during close")
	ErrNotInitialized     = errors.New("logging not initialized, call InitializeLogging first")
)

// Close flushes all pending logs to the configured outputs and closes all
// destinations so InitializeLogging can be called again. Blocks not more than
// specified timeout. This is useful in environments like Kubernetes where the
// application might be terminated and we want to ensure logs are flushed.
//
// Returns ErrTimeoutDuringClose if timeout was reached together with multiple
// errors from the handlers that failed to close within the timeout, or any other
// errors that occurred during the close process.
func Close(timeout time.Duration) error {
	resMu.Lock()
	defer resMu.Unlock()

	if res == nil {
		return ErrNotInitialized
	}

	errs := make(chan error, 3)
	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		defer wg.Done()

		if res.handlerSplunk != nil {
			if err := res.handlerSplunk.CloseWithTimeout(timeout); err != nil {
				if errors.Is(err, splunk.ErrCloseTimeout) {
					errs <- fmt.Errorf("%w: %w", ErrTimeoutDuringClose, err)
				} else {
					errs <- err
				}
			}
		}
	}()

	go func() {
		defer wg.Done()

		if res.handlerCloudWatch != nil {
			if err := res.handlerCloudWatch.CloseWithTimeout(timeout); err != nil {
				if errors.Is(err, cloudwatchwriter2.ErrCloseTimeout) {
					errs <- fmt.Errorf("%w: %w", ErrTimeoutDuringClose, err)
				} else {
					errs <- err
				}
			}
		}
	}()

	go func() {
		defer wg.Done()

		if res.sentryEnabled {
			if !sentry.Flush(timeout) {
				errs <- fmt.Errorf("%w: %w", ErrTimeoutDuringClose, ErrSentryTimeout)
			}
		}
	}()

	wg.Wait()
	close(errs)

	// Close slog logger
	if res.prevSlogger != nil {
		slog.SetDefault(res.prevSlogger)
		res.prevSlogger = nil
	}

	// Close stdlib logger
	if res.prevStdLogger != nil {
		log.SetOutput(res.prevStdLogger)
		log.SetFlags(res.prevStdFlags)
		res.prevStdLogger = nil
		res.prevStdFlags = 0
	}

	// Close strc logger
	strc.SetNoopLogger()

	// Close logrus logger
	logrus.SetDefault(logrus.NewDiscardProxy())

	// Allow re-initialization
	res = nil

	// Collect all errors from the channel as well
	var result []error
	for err := range errs {
		result = append(result, err)
	}

	return errors.Join(result...)
}
