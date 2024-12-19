package strc

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"
)

var Level slog.Level = slog.LevelDebug

// SpanGroupName is the group name used for span attributes.
var SpanGroupName string = "span"

// TraceIDName is the key name used for trace ID.
var TraceIDName string = "trace"

// SkipSource is a flag that disables source logging.
var SkipSource bool

var customLogger = slog.Default()

func SetLogger(lg *slog.Logger) {
	customLogger = lg
}

func logger() *slog.Logger {
	if customLogger == nil {
		return slog.Default()
	}

	return customLogger
}

type Span struct {
	ctx     context.Context
	name    string
	tid     string
	sid     string
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

func StartContext(ctx context.Context, name string, args ...any) (*Span, context.Context) {
	tid := TraceID(ctx)
	if tid == "" {
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

	log := logger().With(
		slog.Group(SpanGroupName,
			// keep the order of name, id, trace_id
			slog.String("name", name),
			slog.String("id", sid),
			slog.String(TraceIDName, tid),
		),
	)
	if !SkipSource {
		log.With(slog.Group(slog.SourceKey, slog.String(slog.SourceKey, callerPtr(2))))
	}
	log.Log(ctx, Level, "span "+name+" started", args...)

	return span, ctx
}

func (s *Span) Event(name string, args ...any) {
	if !logger().Enabled(s.ctx, Level) {
		return
	}

	log := logger().With(
		slog.Group(SpanGroupName,
			// keep the order of name, id, trace_id
			slog.String("name", s.name),
			slog.String("id", s.sid),
			slog.String(TraceIDName, s.tid),
			slog.String("event", name),
			slog.Duration("at", time.Since(s.started)),
		),
	)
	if !SkipSource {
		log.With(slog.Group(slog.SourceKey, slog.String(slog.SourceKey, callerPtr(2))))
	}
	log.Log(s.ctx, Level, "span "+s.name+" event "+name, args...)
}

func (s *Span) End() {
	if !logger().Enabled(s.ctx, Level) {
		return
	}
	dur := time.Since(s.started)

	log := logger().With(
		slog.Group(SpanGroupName,
			// keep the order of name, id, trace_id
			slog.String("name", s.name),
			slog.String("id", s.sid),
			slog.String(TraceIDName, s.tid),
			slog.Duration("dur", dur),
		),
	)
	if !SkipSource {
		log.With(slog.Group(slog.SourceKey, slog.String(slog.SourceKey, callerPtr(2))))
	}
	log.Log(s.ctx, Level, fmt.Sprintf("span %s finished in %v", s.name, dur))
}
