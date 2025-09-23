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

var tracer atomic.Pointer[Tracer]

func init() {
	SetNoopLogger()
}

// SetLogger sets the logger for the package.
func SetLogger(logger *slog.Logger) {
	tracer.Store(NewTracer(logger))
}

// SetNoopLogger sets a no-op logger for the package.
func SetNoopLogger() {
	tracer.Store(NewTracer(slog.New(&NoopHandler{})))
}

// Tracer is a wrapper for slog.Logger which logs into the initialized slog. Use strc.Start
// and End package functions to use slog.Default() logger.
type Tracer struct {
	logger *slog.Logger
}

// NewTracer creates a new Tracer with the given logger. Use strc.Start and End package functions
// to use slog.Default() logger.
func NewTracer(logger *slog.Logger) *Tracer {
	return &Tracer{logger: logger.WithGroup(SpanGroupName)}
}

// Span represents a span of a trace. It is used to log events and end the span.
// It is a lightweight object and can be passed around in contexts.
type Span struct {
	ctx     context.Context
	tracer  *Tracer
	name    string
	tid     TraceID
	sid     SpanID
	args    []any
	started time.Time
}

// Start starts a new span with the given name and optional arguments. All arguments are present in
// subsequent Event and End calls. Must call End to finish the span.
//
//	span, ctx := strc.Start(ctx, "calculating something big")
//	defer span.End()
//
// Avoid adding complex arguments to spans as they are added to the context.
//
// It immediately logs a message with the span name, span information in SpanGroupName and optional
// arguments.
//
// Special argument named "started" of type time.Time can be used to set the start time of the span.
func Start(ctx context.Context, name string, args ...any) (*Span, context.Context) {
	return tracer.Load().Start(ctx, name, args...)
}

// Start starts a new span, see strc.Start for more information.
func (t *Tracer) Start(ctx context.Context, name string, args ...any) (*Span, context.Context) {
	tid := TraceIDFromContext(ctx)
	if tid == EmptyTraceID {
		tid = NewTraceID()
		ctx = WithTraceID(ctx, tid)
	}

	sid := NewSpanID(ctx)
	ctx = WithSpanID(ctx, sid)

	started := time.Now()
	if p := findArgs[time.Time](args, "started"); p != nil {
		started = *p
	}

	span := &Span{
		ctx:     ctx,
		tracer:  t,
		name:    name,
		tid:     tid,
		sid:     sid,
		args:    args,
		started: started,
	}

	if !t.logger.Enabled(ctx, Level) {
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
		attrs = append(attrs, slog.String(slog.SourceKey, callerPtr(3)))
	}

	logger := t.logger
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
//
// Special argument named "at" of type time.Time can be used to set the event time.
func (s *Span) Event(name string, args ...any) {
	if !s.tracer.logger.Enabled(s.ctx, Level) {
		return
	}

	at := time.Now()
	if p := findArgs[time.Time](args, "at"); p != nil {
		at = *p
	}

	// keep the order and capacity correct
	attrs := make([]slog.Attr, 0, 6+1)
	attrs = append(attrs,
		slog.String("name", s.name),
		slog.String(SpanIDName, s.sid.ID()),
		slog.String(ParentIDName, s.sid.ParentID()),
		slog.String(TraceIDName, s.tid.String()),
		slog.String("event", name),
		slog.Duration("at", at.Sub(s.started)),
	)

	if !SkipSource {
		attrs = append(attrs, slog.String(slog.SourceKey, callerPtr(2)))
	}

	logger := s.tracer.logger
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
//
// Special argument named "finished" of type time.Time can be used to set the finish time of the span.
func (s *Span) End(args ...any) {
	if !s.tracer.logger.Enabled(s.ctx, Level) {
		return
	}

	finished := time.Now()
	if p := findArgs[time.Time](args, "finished"); p != nil {
		finished = *p
	}
	dur := finished.Sub(s.started)

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

	logger := s.tracer.logger
	if len(s.args) > 0 {
		logger = logger.With(s.args...)
	}
	if len(args) > 0 {
		logger = logger.With(args...)
	}
	logger.LogAttrs(s.ctx, Level, fmt.Sprintf("span %s finished in %v", s.name, dur), attrs...)
}

// TraceID returns the trace ID of the span.
func (s *Span) TraceID() TraceID {
	return s.tid
}

func callerPtr(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return ""
	}

	return file + ":" + fmt.Sprint(line)
}

func findArgs[T any](args []any, key string) *T {
	// find the key in argument key-value pairs
	for i := 0; i < len(args)-1; i += 2 {
		if k, ok := args[i].(string); ok && k == key {
			if result, ok := args[i+1].(T); ok {
				return &result
			}
		}
	}

	// find the key in slog.Attrs
	for i := range args {
		if attr, ok := args[i].(slog.Attr); ok && attr.Key == key {
			if result, ok := attr.Value.Any().(T); ok {
				return &result
			}
		}
	}

	// not found
	return nil
}
