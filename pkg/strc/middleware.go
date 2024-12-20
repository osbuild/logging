package strc

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	TraceIDKey = "trace_id"
	SpanIDKey  = "span_id"

	RequestBodyMaxSize  = 64 * 1024 // 64KB
	ResponseBodyMaxSize = 64 * 1024 // 64KB

	HiddenRequestHeaders = map[string]struct{}{
		"authorization": {},
		"cookie":        {},
		"set-cookie":    {},
		"x-auth-token":  {},
		"x-csrf-token":  {},
		"x-xsrf-token":  {},
	}
	HiddenResponseHeaders = map[string]struct{}{
		"set-cookie": {},
	}
)

type MiddlewareConfig struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level

	SpanName           string
	WithUserAgent      bool
	WithRequestBody    bool
	WithRequestHeader  bool
	WithResponseBody   bool
	WithResponseHeader bool

	Filters []Filter
}

// NewMiddleware returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func NewMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return NewMiddlewareWithConfig(logger, MiddlewareConfig{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		SpanName:           "http request",
		WithUserAgent:      false,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,

		Filters: []Filter{},
	})
}

// NewMiddlewareWithFilters returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func NewMiddlewareWithFilters(logger *slog.Logger, filters ...Filter) func(http.Handler) http.Handler {
	return NewMiddlewareWithConfig(logger, MiddlewareConfig{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		SpanName:           "http request",
		WithUserAgent:      false,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,

		Filters: filters,
	})
}

// NewMiddlewareWithConfig returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
func NewMiddlewareWithConfig(logger *slog.Logger, config MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			query := r.URL.RawQuery
			method := r.Method
			host := r.Host
			userAgent := r.UserAgent()
			ip := r.RemoteAddr
			referer := r.Referer()

			// dump request body
			br := newBodyReader(r.Body, RequestBodyMaxSize, config.WithRequestBody)
			r.Body = br

			// dump response body
			bw := newBodyWriter(w, ResponseBodyMaxSize, config.WithResponseBody)

			// trace id
			ctx := r.Context()
			traceID := TraceIDFromRequest(r)
			if traceID == EmptyTraceID {
				traceID = NewTraceID()
			}
			ctx = WithTraceID(ctx, traceID)

			// span id
			spanID := SpanIDFromRequest(r)
			ctx = WithSpanID(ctx, spanID)
			r = r.WithContext(ctx)

			defer func() {
				status := bw.Status()
				end := time.Now()
				latency := end.Sub(start)

				baseAttributes := []slog.Attr{
					slog.String(TraceIDKey, traceID.String()),
					slog.String(SpanIDKey, spanID.String()),
				}

				requestAttributes := []slog.Attr{
					slog.Time("time", start.UTC()),
					slog.String("method", method),
					slog.String("host", host),
					slog.String("path", path),
					slog.String("query", query),
					slog.String("ip", ip),
					slog.String("referer", referer),
				}

				responseAttributes := []slog.Attr{
					slog.Time("time", end.UTC()),
					slog.Duration("latency", latency),
					slog.Int("status", status),
				}

				// request body
				requestAttributes = append(requestAttributes, slog.Int("length", br.bytes))
				if config.WithRequestBody {
					requestAttributes = append(requestAttributes, slog.String("body", br.body.String()))
				}

				// request headers
				if config.WithRequestHeader {
					kv := []any{}

					for k, v := range r.Header {
						if _, found := HiddenRequestHeaders[strings.ToLower(k)]; found {
							continue
						}
						kv = append(kv, slog.Any(k, v))
					}

					requestAttributes = append(requestAttributes, slog.Group("header", kv...))
				}

				if config.WithUserAgent {
					requestAttributes = append(requestAttributes, slog.String("user-agent", userAgent))
				}

				// response body
				responseAttributes = append(responseAttributes, slog.Int("length", bw.bytes))
				if config.WithResponseBody {
					responseAttributes = append(responseAttributes, slog.String("body", bw.body.String()))
				}

				// response headers
				if config.WithResponseHeader {
					kv := []any{}

					for k, v := range w.Header() {
						if _, found := HiddenResponseHeaders[strings.ToLower(k)]; found {
							continue
						}
						kv = append(kv, slog.Any(k, v))
					}

					responseAttributes = append(responseAttributes, slog.Group("header", kv...))
				}

				attributes := append(
					[]slog.Attr{
						{
							Key:   "request",
							Value: slog.GroupValue(requestAttributes...),
						},
						{
							Key:   "response",
							Value: slog.GroupValue(responseAttributes...),
						},
					},
					baseAttributes...,
				)

				for _, filter := range config.Filters {
					if !filter(bw, r) {
						return
					}
				}

				level := config.DefaultLevel
				if status >= http.StatusInternalServerError {
					level = config.ServerErrorLevel
				} else if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
					level = config.ClientErrorLevel
				}

				logger.LogAttrs(r.Context(), level, strconv.Itoa(status)+": "+http.StatusText(status), attributes...)
			}()

			next.ServeHTTP(bw, r)
		})
	}
}
