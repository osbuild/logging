package journal

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"

	"github.com/coreos/go-systemd/v22/journal"
)

var enabled bool

func init() {
	// This operation is pretty slow so let's do it only once.
	enabled = journal.Enabled()
}

// Handler for systemd journal
type Handler struct {
	level slog.Level
	group string
	attrs []slog.Attr
}

var _ slog.Handler = (*Handler)(nil)

func NewHandler(_ context.Context, level slog.Level) *Handler {
	h := &Handler{
		level: level,
	}
	return h
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return enabled && level >= h.level
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	if !enabled {
		return nil
	}

	var extra int
	if r.PC != 0 {
		extra++
	}
	values := make(map[string]string, len(h.attrs)+r.NumAttrs()+extra)

	for _, a := range h.attrs {
		values[stringifyPair(a.Key, h.group)] = a.Value.String()
	}

	r.Attrs(func(a slog.Attr) bool {
		values[stringifyPair(a.Key, h.group)] = a.Value.String()
		return true
	})

	if r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		values[stringify(slog.SourceKey)] = fmt.Sprintf("%s:%d", f.File, f.Line)
	}

	journal.Send(r.Message, levelToPriority(r.Level), values)
	return nil
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		level: h.level,
		attrs: attrs,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &Handler{
		level: h.level,
		group: stringify(name) + "_",
	}
}

func levelToPriority(level slog.Level) journal.Priority {
	switch level {
	case slog.LevelDebug:
		return journal.PriDebug
	case slog.LevelInfo:
		return journal.PriInfo
	case slog.LevelWarn:
		return journal.PriWarning
	case slog.LevelError:
		return journal.PriErr
	default:
		return journal.PriInfo
	}
}

func stringifyRune(r rune) rune {
	switch {
	case r >= 'A' && r <= 'Z':
		return r
	case r >= '0' && r <= '9':
		return r
	case r == '_':
		return r
	case r >= 'a' && r <= 'z':
		return r - 32
	default:
		return rune('_')
	}
}

func stringify(key string) string {
	key = strings.Map(stringifyRune, key)
	key = strings.TrimPrefix(key, "_")
	return key
}

func stringifyPair(key, group string) string {
	if group == "" {
		return stringify(key)
	}

	return group + stringify(key)
}
