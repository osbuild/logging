package strc

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// This code is coming from https://github.com/samber/slog-http

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
	// DefaultLevel is the default log level for requests. Defaults to Info.
	DefaultLevel slog.Level

	// ClientErrorLevel is the log level for requests with client errors (4xx). Defaults to Warn.
	ClientErrorLevel slog.Level

	// ServerErrorLevel is the log level for requests with server errors (5xx). Defaults to Error.
	ServerErrorLevel slog.Level

	// SpanName is the name of the span. Defaults to "http request".
	SpanName string

	// WithUserAgent enables logging of the User-Agent header. Defaults to false.
	WithUserAgent bool

	// WithRequestBody enables logging of the request body. Defaults to false.
	WithRequestBody bool

	// WithRequestHeader enables logging of the request headers. Defaults to false.
	WithRequestHeader bool

	// WithResponseBody enables logging of the response body. Defaults to false.
	WithResponseBody bool

	// WithResponseHeader enables logging of the response headers. Defaults to false.
	WithResponseHeader bool

	// WithSpanID enables logging of the span ID. Defaults to false.
	WithTraceID bool

	// WithSpanID enables logging of the span ID. Defaults to false.
	WithSpanID bool

	// Filters is a list of filters to apply before logging. Optional.
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
		SpanName:         "http request",
		Filters:          []Filter{},
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
		SpanName:         "http request",
		Filters:          filters,
	})
}

// NewMiddlewareWithConfig returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
func NewMiddlewareWithConfig(logger *slog.Logger, config MiddlewareConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			path := r.URL.Path
			method := r.Method
			host := r.Host
			userAgent := r.UserAgent()
			ip := r.RemoteAddr

			// dump request body
			br := newBodyReader(r.Body, RequestBodyMaxSize, config.WithRequestBody)
			r.Body = br

			// dump response body
			bw := newBodyWriter(w, ResponseBodyMaxSize, config.WithResponseBody)

			// apply filters early
			for _, filter := range config.Filters {
				if !filter(bw, r) {
					next.ServeHTTP(bw, r)
					return
				}
			}

			// trace id
			ctx := r.Context()
			traceID := TraceIDFromRequest(r)
			var returnTraceID bool
			if traceID == EmptyTraceID {
				traceID = NewTraceID()
				returnTraceID = true
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

				// add trace id to response header
				if returnTraceID {
					w.Header().Add(TraceHTTPHeaderName, traceID.String())
				}

				// build attributes
				baseAttributes := []slog.Attr{}

				if config.WithTraceID {
					baseAttributes = append(baseAttributes, slog.String(TraceIDKey, traceID.String()))
				}

				if config.WithSpanID {
					baseAttributes = append(baseAttributes, slog.String(SpanIDKey, spanID.String()))
				}

				requestAttributes := []slog.Attr{
					slog.Time("time", start.UTC()),
					slog.String("method", method),
					slog.String("host", host),
					slog.String("path", path),
					slog.String("ip", ip),
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
