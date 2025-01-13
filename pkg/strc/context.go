package strc

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type key int

const (
	traceIDKey key = iota
	spanIDKey  key = iota

	traceLength = 15 // ojtlqPCGXEWytHg
	spanLength  = 7  // aCBzdka.NjPdyjv

	TraceHTTPHeaderName = "X-Strc-Trace-ID"
	SpanHTTPHeaderName  = "X-Strc-Span-ID"

	EmptyTraceID TraceID = TraceID("000000000000000")
	EmptySpanID  SpanID  = SpanID("0000000.0000000")
)

// TraceID is a unique identifier for a trace.
type TraceID string

func (t TraceID) String() string {
	if t == "" {
		return EmptyTraceID.String()
	}

	return string(t)
}

// NewTraceID generates a new random trace ID.
func NewTraceID() TraceID {
	return TraceID(randString(traceLength))
}

// SpanID is a unique identifier for a trace.
type SpanID string

func (s SpanID) String() string {
	if s == "" {
		return EmptySpanID.String()
	}

	return string(s)
}

func (s SpanID) ParentID() string {
	return string(s)[:spanLength]
}

func (s SpanID) ID() string {
	return string(s)[spanLength+1 : spanLength*2+1]
}

// NewSpanID generates a new span ID. Uses context to fetch its parent span ID.
func NewSpanID(ctx context.Context) SpanID {
	return SpanID(SpanIDFromContext(ctx).ID() + "." + randString(spanLength))
}

// Start returns trace ID from a context. It returns EmptyTraceID if trace ID is not found.
// Use NewTraceID to generate a new trace ID.
func TraceIDFromContext(ctx context.Context) TraceID {
	if ctx == nil {
		return EmptyTraceID
	}

	if v := ctx.Value(traceIDKey); v != nil {
		return v.(TraceID)
	}

	return EmptyTraceID
}

// Start returns a new context with trace ID.
func WithTraceID(ctx context.Context, traceID TraceID) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromRequest returns trace ID from a request. If trace ID is not found, it returns EmptyTraceID.
func TraceIDFromRequest(req *http.Request) TraceID {
	t := req.Header.Get(TraceHTTPHeaderName)
	if t == "" {
		return EmptyTraceID
	}
	return TraceID(t)
}

// AddTraceIDHeader adds trace ID from context to a request header. If trace ID is not found in the context or
// if the request already has a trace ID header, it does nothing.
func AddTraceIDHeader(ctx context.Context, req *http.Request) {
	traceID := TraceIDFromContext(ctx)
	if traceID != "" && req.Header.Get(TraceHTTPHeaderName) == "" {
		req.Header.Add(TraceHTTPHeaderName, traceID.String())
	}
}

// Start returns span ID from a context. It returns EmptySpanID if span ID is not found.
// Use NewSpanID to generate a new span ID.
func SpanIDFromContext(ctx context.Context) SpanID {
	if ctx == nil {
		return EmptySpanID
	}

	if v := ctx.Value(spanIDKey); v != nil {
		return v.(SpanID)
	}

	return EmptySpanID
}

// Start returns a new context with span ID.
func WithSpanID(ctx context.Context, spanID SpanID) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

// SpanIDFromRequest returns span ID from a request. If span ID is not found, it returns EmptySpanID.
func SpanIDFromRequest(req *http.Request) SpanID {
	s := req.Header.Get(SpanHTTPHeaderName)
	if s == "" {
		return EmptySpanID
	}
	return SpanID(s)
}

// AddSpanIDHeader adds span ID from context to a request header. If span ID is not found in the context or
// if the request already has a span ID header, it does nothing.
func AddSpanIDHeader(ctx context.Context, req *http.Request) {
	spanID := SpanIDFromContext(ctx)
	if spanID != "" && req.Header.Get(SpanHTTPHeaderName) == "" {
		req.Header.Add(SpanHTTPHeaderName, spanID.String())
	}
}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = rand.NewSource(time.Now().UnixNano())

func randString(n int) string {
	// decently fast random string generator
	sb := strings.Builder{}
	sb.Grow(n)
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return sb.String()
}
