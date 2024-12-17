package strc

import (
	"context"
	"math/rand"
	"strings"
	"time"
)

type key int

const (
	traceIDKey key = iota
	spanIDKey  key = iota
)

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func TraceID(ctx context.Context) string {
	if v := ctx.Value(traceIDKey); v != nil {
		return v.(string)
	}
	return ""
}

func WithSpanID(ctx context.Context, spanID string) context.Context {
	return context.WithValue(ctx, spanIDKey, spanID)
}

func SpanID(ctx context.Context) string {
	if v := ctx.Value(spanIDKey); v != nil {
		return v.(string)
	}
	return ""
}

func NewTraceID() string {
	return randString(16)
}

func NewSpanID(ctx context.Context) string {
	parent := SpanID(ctx)
	if parent == "" {
		return randString(12)
	}
	return parent + "." + randString(12)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
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
