package strc

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
)

// ExportHandler is an slog.Handler which provides a callback function that allows for exporting
// attributes (slog.Attr) to a different system. For more information on how to write efficient
// slog handlers: https://github.com/golang/example/blob/master/slog-handler-guide/README.md
type ExportHandler struct {
	callback      func([]slog.Attr)
	goas          []groupOrAttrs
	includeTime   bool
	includeSource bool
}

var _ slog.Handler = (*ExportHandler)(nil)

type ExporterOption func(*ExportHandler)

// IncludeTime is an ExporterOption that includes the time in the exported attributes.
func IncludeTime() ExporterOption {
	return func(h *ExportHandler) {
		h.includeTime = true
	}
}

// IncludeSource is an ExporterOption that includes the source in the exported attributes.
func IncludeSource() ExporterOption {
	return func(h *ExportHandler) {
		h.includeSource = true
	}
}

// NewExportHandler creates a new Exporter with the given callback function and options.
func NewExportHandler(export func([]slog.Attr), opts ...ExporterOption) slog.Handler {
	h := &ExportHandler{callback: export}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *ExportHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true
}

// groupOrAttrs holds either a group name or a list of slog.Attrs.
type groupOrAttrs struct {
	group string      // group name if non-empty
	attrs []slog.Attr // attrs if non-empty
}

func (h *ExportHandler) withGroupOrAttrs(goa groupOrAttrs) *ExportHandler {
	h2 := *h
	h2.goas = make([]groupOrAttrs, len(h.goas)+1)
	copy(h2.goas, h.goas)
	h2.goas[len(h2.goas)-1] = goa
	return &h2
}

func (h *ExportHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return h.withGroupOrAttrs(groupOrAttrs{group: name})
}

func (h *ExportHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	return h.withGroupOrAttrs(groupOrAttrs{attrs: attrs})
}

func (h *ExportHandler) Handle(ctx context.Context, r slog.Record) error {
	result := make([]slog.Attr, 0, 8)

	if h.includeTime && !r.Time.IsZero() {
		result = append(result, slog.Time(slog.TimeKey, r.Time))
	}

	if h.includeSource && r.PC != 0 {
		// TODO remove one frame
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		result = append(result, slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", f.File, f.Line)))
	}

	goas := h.goas
	if r.NumAttrs() == 0 {
		// If the record has no Attrs, remove groups at the end of the list; they are empty.
		for len(goas) > 0 && goas[len(goas)-1].group != "" {
			goas = goas[:len(goas)-1]
		}
	}
	group := make([]string, 0, len(goas))
	for _, goa := range goas {
		if goa.group != "" {
			group = append(group, goa.group)
		} else {
			result = append(result, makeGroup(group, goa.attrs)...)
		}
	}
	
	recAttrs := make([]slog.Attr, 0, r.NumAttrs())
	r.Attrs(func(attrs slog.Attr) bool {
		recAttrs = append(recAttrs, attrs)
		return true
	})
	result = append(result, makeGroup(group, recAttrs)...)

	h.callback(result)
	return nil
}

func makeGroup(groups []string, attrs []slog.Attr) []slog.Attr {
	if len(groups) == 0 {
		return attrs
	}

	var result slog.Attr = slog.Attr{Key: groups[len(groups)-1], Value: slog.GroupValue(attrs...)}
	if len(groups) > 1 {
		for i := len(groups) - 2; i >= 0; i-- {
			result = slog.Group(groups[i], result)
		}
	}

	return []slog.Attr{result}
}
