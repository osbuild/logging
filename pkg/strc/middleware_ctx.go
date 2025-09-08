package strc

import (
	"context"
	"log/slog"
	"net/http"
)

// HeadfieldPair is a pair of header name and field name.
type HeadfieldPair struct {
	HeaderName string
	FieldName  string
}

// HeadfieldPairMiddleware is a middleware that extracts values from headers and stores them in the context.
func HeadfieldPairMiddleware(pairs []HeadfieldPair) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, p := range pairs {
				if value := r.Header.Get(p.HeaderName); value != "" {
					r = r.WithContext(context.WithValue(r.Context(), p, value))
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// HeadfieldPairCallback is a slog callback that extracts values from the context and adds them to the attributes.
func HeadfieldPairCallback(pairs []HeadfieldPair) MultiCallback {
	return func(ctx context.Context, a []slog.Attr) ([]slog.Attr, error) {
		for _, p := range pairs {
			if value, ok := ctx.Value(p).(string); ok {
				a = append(a, slog.String(p.FieldName, value))
			}
		}

		return a, nil
	}
}
