package strc

import (
	"context"
	"log/slog"
)

// ExportHandler is an slog.Handler which provides a callback function that allows for exporting
// attributes (slog.Attr) to a different system. This handler is not optimized for performance,
// it is recommended to write a dedicated slog.Handler. For more information on the topic, see
// https://github.com/golang/example/blob/master/slog-handler-guide/README.md
type ExportHandler struct {
	callback    func(context.Context, []slog.Attr)
	attrs       []slog.Attr
	groups      []string
	includeTime bool
}

var _ slog.Handler = (*ExportHandler)(nil)

type ExporterOption func(*ExportHandler)

// IncludeTime is an ExporterOption that includes the time in the exported attributes.
func IncludeTime() ExporterOption {
	return func(h *ExportHandler) {
		h.includeTime = true
	}
}

// NewExportHandler creates a new Exporter with the given callback function and options.
func NewExportHandler(export func(context.Context, []slog.Attr), opts ...ExporterOption) slog.Handler {
	h := &ExportHandler{callback: export}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *ExportHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= Level
}

func (h *ExportHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ExportHandler{
		callback:    h.callback,
		attrs:       appendAttrsToGroup(h.groups, h.attrs, attrs...),
		groups:      h.groups,
		includeTime: h.includeTime,
	}
}

func (h *ExportHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &ExportHandler{
		callback:    h.callback,
		attrs:       h.attrs,
		groups:      append(h.groups, name),
		includeTime: h.includeTime,
	}
}

func (h *ExportHandler) Handle(ctx context.Context, r slog.Record) error {
	// extra capacity is allocated for the new attributes
	attrs := appendRecordAttrsToAttrs(h.attrs, h.groups, &r)

	if h.includeTime && !r.Time.IsZero() {
		attrs = append(attrs, slog.Group(SpanGroupName, slog.Time(slog.TimeKey, r.Time)))
	}

	h.callback(ctx, attrs)

	return nil
}
