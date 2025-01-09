package splunk

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"
)

var _ slog.Handler = (*SplunkHandler)(nil)
var _ io.Writer = (*SplunkHandler)(nil)

const (
	// EventKey is the key used to group the event attributes.
	EventKey = "event"
)

// SplunkHandler sends records to a Splunk instance as events.
type SplunkHandler struct {
	level  slog.Level
	splunk *splunkLogger
	jh     slog.Handler
}

func replaceAttr(groups []string, a slog.Attr) slog.Attr {
	if groups == nil && a.Key == slog.TimeKey {
		return slog.Int64(slog.TimeKey, time.Now().Unix())
	}

	return a
}

// SplunkConfig is the configuration for the Splunk handler.
type SplunkConfig struct {
	// Level is the minimum level of logs that will be sent to Splunk.
	Level slog.Level

	// URL is the Splunk HEC endpoint.
	URL string

	// Token is the Splunk HEC token.
	Token string

	// Source is the source of the logs.
	Source string

	// Hostname is the hostname of the logs.
	Hostname string

	// DefaultMaximumSize is the initialized capacity of the event buffer before it is flushed, default is 1MB.
	DefaultMaximumSize int
}

// NewSplunkHandler creates a new SplunkHandler. It uses highly-optimized JSON handler from
// the standard library to format the log records. The handler implements io.Writer interface
// which is then used to stream JSON data into the Splunk client.
func NewSplunkHandler(ctx context.Context, config SplunkConfig) *SplunkHandler {
	h := &SplunkHandler{
		level:  config.Level,
		splunk: newSplunkLogger(ctx, config.URL, config.Token, config.Source, config.Hostname, config.DefaultMaximumSize),
	}

	h.jh = slog.NewJSONHandler(h, &slog.HandlerOptions{Level: config.Level, AddSource: true, ReplaceAttr: replaceAttr})
	return h
}

// Flush flushes all pending payloads to the Splunk client. This is done automatically and it is not necessary
// to call this method unless you want to force the flush manually (e.g. in an unit test). Calling this method
// does not guarantee immediate delivery of the payloads to the Splunk instance.
func (h *SplunkHandler) Flush() {
	h.splunk.flush()
}

// Close flushes all pending payloads and closes the Splunk client. Sending new logs after
// closing the handler will return ErrFullOrClosed. The call can block but not longer than 2 seconds.
func (h *SplunkHandler) Close() {
	h.splunk.close()
}

// Write is called by the JSON handler to write the JSON payload to the Splunk client.
func (h *SplunkHandler) Write(buf []byte) (int, error) {
	return h.splunk.event(buf)
}

func (h *SplunkHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *SplunkHandler) Handle(ctx context.Context, r slog.Record) error {
	err := h.jh.Handle(ctx, r)

	// Since errors are silently ignored in slog, let's make an good will attempt.
	if err != nil {
		fmt.Fprintf(os.Stderr, "splunk handler error: %v\n", err)
	}

	return err
}

func (h *SplunkHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SplunkHandler{
		level:  h.level,
		splunk: h.splunk,
		jh:     h.jh.WithAttrs(attrs),
	}
}

func (h *SplunkHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &SplunkHandler{
		level:  h.level,
		splunk: h.splunk,
		jh:     h.jh.WithGroup(name),
	}
}

// Statistics returns the statistics of the Splunk client.
func (h *SplunkHandler) Statistics() Stats {
	return h.splunk.statistics()
}
