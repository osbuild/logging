package strc

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

var Level slog.Level = slog.LevelDebug

var SpanGroupName string = "span"
var TraceIDName string = "trace"

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

	logger().With(
		slog.Group(SpanGroupName,
			slog.String("name", name),
			slog.String(TraceIDName, tid),
			slog.String("id", sid),
		),
	).Log(ctx, Level, "span "+name+" started", args...)

	return span, ctx
}

func (s *Span) Event(name string, args ...any) {
	if !logger().Enabled(s.ctx, Level) {
		return
	}

	logger().With(
		slog.Group(SpanGroupName,
			slog.String("event", name),
			slog.String("name", s.name),
			slog.String("id", s.sid),
			slog.Duration("at", time.Since(s.started)),
			slog.String(TraceIDName, s.tid),
		),
	).Log(s.ctx, Level, "span "+s.name+" event "+name, args...)
}

func (s *Span) End() {
	if !logger().Enabled(s.ctx, Level) {
		return
	}

	logger().With(
		slog.Group(SpanGroupName,
			slog.String("name", s.name),
			slog.String("id", s.sid),
			slog.String(TraceIDName, s.tid),
			slog.Duration("dur", time.Since(s.started)),
		),
	).Log(s.ctx, Level, fmt.Sprintf("span %s finished in %v", s.name, time.Since(s.started)))
}
