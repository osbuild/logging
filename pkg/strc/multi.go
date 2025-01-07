package strc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"github.com/osbuild/logging"
)

// Ported from https://github.com/samber/slog-multi to avoid a dependency

var _ slog.Handler = (*MultiHandler)(nil)

var (
	// TraceIDFieldKey is the key used to store the trace ID in the log record by MultiHandler.
	// Set to empty string to disable this feature.
	TraceIDFieldKey = "trace_id"

	// BuildIDFieldKey is the key used to store the git commit in the log record by MultiHandler.
	// Set to empty string to disable this feature.
	BuildIDFieldKey = "build_id"
)

// MultiHandler distributes records to multiple slog.Handler
type MultiHandler struct {
	handlers []slog.Handler
	inGroup  bool
	callback MultiCallback
}

type MultiCallback func(context.Context, []slog.Attr) error

// NewMultiHandler distributes records to multiple slog.Handler
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	return NewMultiHandlerCustom(nil, nil, handlers...)
}

// NewMultiHandlerCustom distributes records to multiple slog.Handler
// with custom attributes and callback. Pass static slice of attributes added
// to the every record, and a callback that can add dynamic attributes from the context.
func NewMultiHandlerCustom(attrs []slog.Attr, callback MultiCallback, handlers ...slog.Handler) *MultiHandler {
	a := make([]slog.Attr, 0, len(attrs)+1)
	a = append(a, attrs...)

	if BuildIDFieldKey != "" {
		a = append(a, slog.Attr{
			Key:   BuildIDFieldKey,
			Value: slog.StringValue(logging.BuildID()),
		})
	}

	for i := range handlers {
		handlers[i] = handlers[i].WithAttrs(a)
	}

	return &MultiHandler{
		handlers: handlers,
		callback: callback,
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

func (h *MultiHandler) Handle(ctx context.Context, recOrig slog.Record) error {
	r := recOrig.Clone()

	if !h.inGroup {
		attrs := make([]slog.Attr, 0, 2)

		// add optional trace_id attribute
		if id := TraceIDFromContext(ctx); id != EmptyTraceID && TraceIDFieldKey != "" {
			attrs = append(attrs, slog.Attr{
				Key:   TraceIDFieldKey,
				Value: slog.StringValue(id.String()),
			})
		}

		// add zero or more optional attributes
		if h.callback != nil {
			if err := h.callback(ctx, attrs); err != nil {
				return err
			}
		}

		r.AddAttrs(attrs...)
	}

	var errs []error
	for i := range h.handlers {
		if h.handlers[i].Enabled(ctx, r.Level) {
			err := try(func() error {
				return h.handlers[i].Handle(ctx, r)
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

	return &MultiHandler{
		handlers: handlers,
		inGroup:  true,
	}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	handlers := make([]slog.Handler, len(h.handlers))
	for i := range h.handlers {
		handlers[i] = h.handlers[i].WithGroup(name)
	}

	return &MultiHandler{
		handlers: handlers,
		inGroup:  true,
	}
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
