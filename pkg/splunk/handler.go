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

// NewSplunkHandler creates a new SplunkHandler. It uses highly-optimized JSON handler from
// the standard library to format the log records. The handler implements io.Writer interface
// which is then used to stream JSON data into the Splunk client.
func NewSplunkHandler(ctx context.Context, level slog.Level, url, token, source, hostname string) *SplunkHandler {
	h := &SplunkHandler{
		level:  level,
		splunk: newSplunkLogger(ctx, url, token, source, hostname, DefaultPayloadsChannelSize, DefaultMaximumSize, DefaultSendFrequency),
	}

	h.jh = slog.NewJSONHandler(h, &slog.HandlerOptions{Level: level, AddSource: true, ReplaceAttr: replaceAttr})
	return h
}

func newSplunkHandlerWithParams(
	ctx context.Context,
	level slog.Level,
	url, token, source, hostname string,
	maximumSize int,
	payloadsChannelSize int,
	sendFrequency time.Duration) *SplunkHandler {
	h := &SplunkHandler{
		level:  level,
		splunk: newSplunkLogger(ctx, url, token, source, hostname, payloadsChannelSize, maximumSize, sendFrequency),
	}

	h.jh = slog.NewJSONHandler(h, &slog.HandlerOptions{Level: level, AddSource: true, ReplaceAttr: replaceAttr})
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
	return h.splunk.Statistics()
}
