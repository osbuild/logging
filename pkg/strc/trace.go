package strc

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync/atomic"
	"time"
)

// Level is the log level used for trace logging.
var Level slog.Level = slog.LevelDebug

// SpanGroupName is the group name used for span attributes.
var SpanGroupName string = "span"

// TraceIDName is the key name used for trace ID.
var TraceIDName string = "trace"

// SpanIDName is the key name used for trace ID.
var SpanIDName string = "id"

// ParentIDName is the key name used for trace ID.
var ParentIDName string = "parent"

// SkipSource is a flag that disables source logging.
var SkipSource bool

var destination atomic.Pointer[slog.Logger]

func init() {
	destination.Store(slog.New(&NoopHandler{}))
}

// SetLogger sets the logger for the package.
func SetLogger(lg *slog.Logger) {
	destination.Store(lg.WithGroup(SpanGroupName))
}

// SetNoopLogger sets a no-op logger for the package.
func SetNoopLogger() {
	destination.Store(slog.New(&NoopHandler{}))
}

func logger() *slog.Logger {
	return destination.Load()
}

type Span struct {
	ctx     context.Context
	name    string
	tid     TraceID
	sid     SpanID
	args    []any
	started time.Time
}

func callerPtr(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	return file + ":" + fmt.Sprint(line)
}

// Start starts a new span with the given name and optional arguments. All arguments are present in
// subsequent Event and End calls. Must call End to finish the span.
//
//	span, ctx := strc.Start(ctx, "calculating something big")
//	defer span.End()
//
// It immediately logs a message with the span name, span information in SpanGroupName and optional
// arguments.
func Start(ctx context.Context, name string, args ...any) (*Span, context.Context) {
	tid := TraceIDFromContext(ctx)
	if tid == EmptyTraceID {
		tid = NewTraceID()
		ctx = WithTraceID(ctx, tid)
	}

	sid := NewSpanID(ctx)
	ctx = WithSpanID(ctx, sid)

	span := &Span{
		ctx:     ctx,
		name:    name,
		tid:     tid,
		sid:     sid,
		args:    args,
		started: time.Now(),
	}

	if !logger().Enabled(ctx, Level) {
		// Return early if logging is disabled with all arguments in case
		// level changes during span lifetime. But we still need to return
		// the span and context.
		return span, ctx
	}

	// keep the order and capacity correct
	attrs := make([]slog.Attr, 0, 4+1)
	attrs = append(attrs,
		slog.String("name", name),
		slog.String(SpanIDName, sid.ID()),
		slog.String(ParentIDName, sid.ParentID()),
		slog.String(TraceIDName, tid.String()),
	)

	if !SkipSource {
		attrs = append(attrs, slog.String(slog.SourceKey, callerPtr(2)))
	}

	logger := logger()
	if len(args) > 0 {
		logger = logger.With(args...)
	}
	logger.LogAttrs(ctx, Level, "span "+name+" started", attrs...)

	return span, ctx
}

// Event logs a new event in the span with the given name and optional arguments.
//
// It immediately logs a message with the span name, span information in SpanGroupName and
// optional arguments.
func (s *Span) Event(name string, args ...any) {
	if !logger().Enabled(s.ctx, Level) {
		return
	}

	// keep the order and capacity correct
	attrs := make([]slog.Attr, 0, 6+1)
	attrs = append(attrs,
		slog.String("name", s.name),
		slog.String(SpanIDName, s.sid.ID()),
		slog.String(ParentIDName, s.sid.ParentID()),
		slog.String(TraceIDName, s.tid.String()),
		slog.String("event", name),
		slog.Duration("at", time.Since(s.started)),
	)

	if !SkipSource {
		attrs = append(attrs, slog.String(slog.SourceKey, callerPtr(2)))
	}

	logger := logger()
	if len(s.args) > 0 {
		logger = logger.With(s.args...)
	}
	if len(args) > 0 {
		logger = logger.With(args...)
	}
	logger.LogAttrs(s.ctx, Level, "span "+s.name+" event "+name, attrs...)
}

// End finishes the span and logs the duration of the span. Optional arguments can be provided
// to log additional information.
//
// It immediately logs a message with the span name, span information in SpanGroupName and
// optional arguments.
func (s *Span) End(args ...any) {
	if !logger().Enabled(s.ctx, Level) {
		return
	}
	dur := time.Since(s.started)

	// keep the order and capacity correct
	attrs := make([]slog.Attr, 0, 5+1)
	attrs = append(attrs,
		slog.String("name", s.name),
		slog.String(SpanIDName, s.sid.ID()),
		slog.String(ParentIDName, s.sid.ParentID()),
		slog.String(TraceIDName, s.tid.String()),
		slog.Duration("dur", dur),
	)

	if !SkipSource {
		attrs = append(attrs, slog.String(slog.SourceKey, callerPtr(2)))
	}

	logger := logger()
	if len(s.args) > 0 {
		logger = logger.With(s.args...)
	}
	if len(args) > 0 {
		logger = logger.With(args...)
	}
	logger.LogAttrs(s.ctx, Level, fmt.Sprintf("span %s finished in %v", s.name, dur), attrs...)
}
