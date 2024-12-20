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
)

const (
	traceHTTPHeader = "X-Strc-Trace-ID"
	spanHTTPHeader  = "X-Strc-Span-ID"
)

// StartContext returns trace ID from a context. It returns an empty string if trace ID is not found.
// Use NewTraceID to generate a new trace ID.
func TraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if v := ctx.Value(traceIDKey); v != nil {
		return v.(string)
	}
	
	return ""
}

// TraceIDFromContext is an alias for TraceID.
func TraceIDFromContext(ctx context.Context) string {
	return TraceID(ctx)
}

// StartContext returns a new context with trace ID.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// TraceIDFromRequest returns trace ID from a request. If trace ID is not found, it returns an empty string.
func TraceIDFromRequest(req *http.Request) string {
	return req.Header.Get(traceHTTPHeader)
}

// AddTraceIDHeader adds trace ID from context to a request header. If trace ID is not found in the context or
// if the request already has a trace ID header, it does nothing.
func AddTraceIDHeader(ctx context.Context, req *http.Request) {
	traceID := TraceID(ctx)
	if traceID != "" && req.Header.Get(traceHTTPHeader) == "" {
		req.Header.Add(traceHTTPHeader, traceID)
	}
}

// StartContext returns span ID from a context. It returns an empty string if span ID is not found.
// Use NewSpanID to generate a new span ID.
func SpanID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if v := ctx.Value(spanIDKey); v != nil {
		return v.(string)
	}

	return ""
}

// SpanIDFromContext is an alias for SpanID.
func SpanIDFromContext(ctx context.Context) string {
	return SpanID(ctx)
}

// StartContext returns a new context with span ID.
func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

// SpanIDFromRequest returns span ID from a request. If span ID is not found, it returns an empty string.
func SpanIDFromRequest(req *http.Request) string {
	return req.Header.Get(spanHTTPHeader)
}

// AddSpanIDHeader adds span ID from context to a request header. If span ID is not found in the context or
// if the request already has a span ID header, it does nothing.
func AddSpanIDHeader(ctx context.Context, req *http.Request) {
	spanID := SpanID(ctx)
	if spanID != "" && req.Header.Get(spanHTTPHeader) == "" {
		req.Header.Add(spanHTTPHeader, spanID)
	}
}

// NewTraceID generates a new trace ID.
func NewTraceID() string {
	return randString(13)
}

// NewSpanID generates a new span ID.
func NewSpanID(ctx context.Context) string {
	parent := SpanID(ctx)
	if parent == "" {
		return randString(7)
	}
	return parent + "." + randString(7)
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
