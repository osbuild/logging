package sinit

import (
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
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
// Do not use this function during application exit, use Close() instead to
// ensure all logs are flushed properly.
func Flush() {
	if handlerSplunk != nil {
		handlerSplunk.Flush()
	}

	if handlerCloudWatch != nil {
		handlerCloudWatch.Flush()
	}

	sentry.Flush(2 * time.Second)
}

var onceClose sync.Once

// Close flushes all pending logs to the configured outputs and closes all
// destinations. Blocks not more than specified timeout. This is useful in
// environments like Kubernetes where the application might be terminated and we
// want to ensure logs are flushed.
//
// Returns true if timeout was not reached and all payloads were sent or if it
// was already closed, false if the timeout was reached.
func Close(timeout time.Duration) bool {
	result := true

	onceClose.Do(func() {
		each := timeout / 3

		if handlerSplunk != nil {
			result = handlerSplunk.CloseWithTimeout(each) && result
		}

		if handlerCloudWatch != nil {
			result = handlerCloudWatch.CloseWithTimeout(each) && result
		}

		result = sentry.Flush(each) && result
	})

	return result
}
