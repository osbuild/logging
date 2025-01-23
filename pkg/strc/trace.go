package strc

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"sync/atomic"
	"time"
)

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

func SetLogger(lg *slog.Logger) {
	destination.Store(lg.WithGroup(SpanGroupName))
}

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

	logger().LogAttrs(ctx, Level, "span "+name+" started", attrs...)

	return span, ctx
}

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

	logger().LogAttrs(s.ctx, Level, "span "+s.name+" event "+name, attrs...)
}

func (s *Span) End() {
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

	logger().LogAttrs(s.ctx, Level, fmt.Sprintf("span %s finished in %v", s.name, dur), attrs...)
}
