package collect

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"runtime"
	"sort"
	"sync"
)

// CollectorHandler is a test handler that buffers log entries. This is useful for testing
// of log output.
type CollectorHandler struct {
	level     slog.Level
	addTime   bool
	addSource bool
	addLevel  bool
	goas      []groupOrAttrs
	data      *data
}

type groupOrAttrs struct {
	group string
	attrs []slog.Attr
}

// prevent mutex copying
type data struct {
	fields []map[string]any
	mu     sync.RWMutex
}

var _ slog.Handler = (*CollectorHandler)(nil)

// NewTestHandler creates a new BufferHandler.
func NewTestHandler(level slog.Level, addTime, addSource, addLevel bool) *CollectorHandler {
	h := &CollectorHandler{
		level:     level,
		addTime:   addTime,
		addSource: addSource,
		addLevel:  addLevel,
		data: &data{
			fields: make([]map[string]any, 0),
			mu:     sync.RWMutex{},
		},
	}
	return h
}

func (h *CollectorHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *CollectorHandler) add(m map[string]any) {
	h.data.mu.Lock()
	defer h.data.mu.Unlock()

	h.data.fields = append(h.data.fields, m)
}

func (h *CollectorHandler) Handle(ctx context.Context, r slog.Record) error {
	m := make(map[string]any)

	if TestID(ctx) != "" {
		m["test_id"] = TestID(ctx)
	}

	m[slog.MessageKey] = r.Message

	if h.addLevel {
		m[slog.LevelKey] = r.Level.String()
	}

	if h.addTime && !r.Time.IsZero() {
		m[slog.TimeKey] = r.Time
	}

	if r.PC != 0 && h.addSource {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		m[slog.SourceKey] = fmt.Sprintf("%s:%d", f.File, f.Line)
	}

	var groups []string
	for _, goa := range h.goas {
		if goa.group != "" {
			groups = append(groups, goa.group)
		} else {
			for _, a := range goa.attrs {
				h.appendAttr(a, m, groups)
			}
		}
	}

	r.Attrs(func(a slog.Attr) bool {
		h.appendAttr(a, m, groups)
		return true
	})

	h.add(m)
	return nil
}

func (h *CollectorHandler) appendAttr(a slog.Attr, m map[string]any, g []string) {
	a.Value = a.Value.Resolve()
	if a.Equal(slog.Attr{}) {
		return
	}

	switch a.Value.Kind() {
	case slog.KindGroup:
		attrs := a.Value.Group()
		if len(attrs) == 0 {
			return
		}
		if a.Key == "" {
			for _, ga := range attrs {
				h.appendAttr(ga, m, g)
			}
		}
		m[a.Key] = make(map[string]any)
		for _, ga := range attrs {
			h.appendAttr(ga, m[a.Key].(map[string]any), g)
		}
	default:
		add_p(a, m, g)
	}
}

func add_p(a slog.Attr, m map[string]any, g []string) {
	if len(g) == 0 {
		m[a.Key] = a.Value.Any()
		return
	}
	group := g[0]
	if _, ok := m[group]; !ok {
		m[group] = make(map[string]any)
	}

	add_p(a, m[group].(map[string]any), g[1:])
}

func (h *CollectorHandler) withGroupOrAttrs(goa groupOrAttrs) *CollectorHandler {
	h2 := *h
	h2.goas = make([]groupOrAttrs, len(h.goas)+1)
	copy(h2.goas, h.goas)
	h2.goas[len(h2.goas)-1] = goa
	return &h2
}

func (h *CollectorHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}

	return h.withGroupOrAttrs(groupOrAttrs{attrs: attrs})
}

func (h *CollectorHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return h.withGroupOrAttrs(groupOrAttrs{group: name})
}

// Last returns the last entry that was logged or an empty map
func (h *CollectorHandler) Last() map[string]any {
	h.data.mu.RLock()
	defer h.data.mu.RUnlock()

	i := len(h.data.fields) - 1
	if i < 0 {
		return make(map[string]any)
	}
	return h.data.fields[i]
}

// All returns all entries that were logged.
func (h *CollectorHandler) All() []map[string]any {
	h.data.mu.RLock()
	defer h.data.mu.RUnlock()

	return h.data.fields
}

// Count returns number of all entries that were logged.
func (h *CollectorHandler) Count() int {
	h.data.mu.RLock()
	defer h.data.mu.RUnlock()

	return len(h.data.fields)
}

// Can be replaced by slices.SortedKeys after Go 1.23+ upgrade
func sortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

// String returns all records formatted as a string.
func (h *CollectorHandler) String() string {
	h.data.mu.RLock()
	defer h.data.mu.RUnlock()

	var s string
	for _, m := range h.data.fields {
		var keys []string
		for _, k := range sortedKeys(m) {
			keys = append(keys, fmt.Sprintf("%s=%v", k, m[k]))
		}
		s += fmt.Sprintf("%s\n", keys)
	}
	return s
}

// Reset removes all Entries from this test hook.
func (h *CollectorHandler) Reset() {
	h.data.mu.Lock()
	defer h.data.mu.Unlock()

	h.data.fields = make([]map[string]any, 0)
}

// Contains searches through all collected logs for expected value. It supports nested
// maps of fields (slog groups), and returns true if the value matches expected value.
// Native slog fields like message, time or level are also supported.
//
// Example: a record created with logger.WithGroup("g").Debug("test", "key", "value") will
// be a match for call Contains("value", "g", "key").
func (h *CollectorHandler) Contains(expected any, names ...string) bool {
	h.data.mu.RLock()
	defer h.data.mu.RUnlock()

	if len(names) == 0 {
		panic("at least one name must be provided")
	}

	for _, record := range h.data.fields {
		value, found := h.findNested(record, names)
		if found && reflect.DeepEqual(value, expected) {
			return true
		}
	}

	return false
}

func (h *CollectorHandler) findNested(data map[string]any, names []string) (any, bool) {
	current, ok := data[names[0]]
	if !ok {
		return nil, false
	}

	if len(names) == 1 {
		return current, true
	}

	if nextData, ok := current.(map[string]any); ok {
		return h.findNested(nextData, names[1:])
	}

	return nil, false
}
