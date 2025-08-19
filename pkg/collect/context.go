package collect

import (
	"context"
)

type key int

const (
	testIDKey key = iota
)

func WithTestID(ctx context.Context, str string) context.Context {
	return context.WithValue(ctx, testIDKey, str)
}

func TestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if v := ctx.Value(testIDKey); v != nil {
		return v.(string)
	}

	return ""
}
