package strc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
)

// Ported from https://github.com/samber/slog-multi to avoid a dependency

var _ slog.Handler = (*MultiHandler)(nil)

// MultiHandler distributes records to multiple slog.Handler
type MultiHandler struct {
	handlers []slog.Handler
}

// NewMultiHandler distributes records to multiple slog.Handler
func NewMultiHandler(handlers ...slog.Handler) slog.Handler {
	return &MultiHandler{
		handlers: handlers,
	}
}

func (h *MultiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, level) {
			return true
		}
	}

	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err := try(func() error {
				return h.handlers[i].Handle(ctx, r.Clone())
			})
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlers[i] = h.handlers[i].WithAttrs(slices.Clone(attrs))
	}

	return NewMultiHandler(handlers...)
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	handlers := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlers[i] = h.handlers[i].WithGroup(name)
	}

	return NewMultiHandler(handlers...)
}

func try(callback func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("multi handler error: %+v", r)
			}
		}
	}()

	err = callback()

	return
}
