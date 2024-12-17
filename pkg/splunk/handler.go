package splunk

import (
	"context"
	"io"
	"log/slog"
)

var _ slog.Handler = (*SplunkHandler)(nil)
var _ io.Writer = (*SplunkHandler)(nil)

// SplunkHandler sends records to a Splunk instance as events.
type SplunkHandler struct {
	level  slog.Level
	splunk *splunkLogger
	jh     slog.Handler
}

// NewSplunkHandler creates a new SplunkHandler. It uses highly-optimized JSON handler from
// the standard library to format the log records. The handler implements io.Writer interface
// which is then used to stream JSON data into the Splunk client.
func NewSplunkHandler(ctx context.Context, level slog.Level, url, token, source, hostname string) slog.Handler {
	h := &SplunkHandler{
		level: level,
		splunk: newSplunkLogger(ctx, url, token, source, hostname),
	}
	h.jh = slog.NewJSONHandler(h, &slog.HandlerOptions{Level: level, AddSource: true})
	return h
}

// Close flushes all pending payloads and closes the Splunk client. All new log records
// will be discarded after the handler is closed.
func (h *SplunkHandler) Close() {
	h.splunk.close()
}

func (h *SplunkHandler) Write(buf []byte) (int, error) {
	h.splunk.event(buf)

	return len(buf), nil
}

func (h *SplunkHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *SplunkHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.jh.Handle(ctx, r)
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
