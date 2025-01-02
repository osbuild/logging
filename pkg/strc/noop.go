package strc

import (
	"context"
	"log/slog"
)

var _ slog.Handler = (*NoopHandler)(nil)

// NoopHandler does nothing.
type NoopHandler struct {
}

func (h *NoopHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return false
}

func (h *NoopHandler) Handle(ctx context.Context, r slog.Record) error {
	return nil
}

func (h *NoopHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *NoopHandler) WithGroup(name string) slog.Handler {
	return h
}
