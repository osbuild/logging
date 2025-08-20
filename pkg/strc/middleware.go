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

	// NoExtractTraceID disables extracting trace id from incoming requests. Defaults to false.
	NoTraceContext bool

	// Filters is a list of filters to apply before logging. Optional.
	Filters []Filter
}

// NewMiddleware returns a `func(http.Handler) http.Handler` (middleware) that logs requests using slog.
//
// In addition to that, it also extracts span id and trace id from incoming requests and puts them into
// the context. If the request does not have a trace id, a new one is generated. This feature can be
// disabled.
//
// Finally, it can be configured to dump request and response bodies, request and response headers,
// and user agent.
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
			m := middleware{
				logger: logger,
				config: config,
				w:      w,
				r:      r,
				callFunc: func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
				},
			}

			m.before()
			defer m.after()

			next.ServeHTTP(m.bw, m.r)
		})
	}
}

type middleware struct {
	logger *slog.Logger
	config MiddlewareConfig
	w      http.ResponseWriter
	r      *http.Request
	bw     *bodyWriter
	br     *bodyReader

	callFunc func(w http.ResponseWriter, r *http.Request)

	start         time.Time
	path          string
	method        string
	host          string
	userAgent     string
	ip            string
	returnTraceID bool
	traceID       TraceID
	spanID        SpanID

	status   int
	internal error
	message  any
}

func (m *middleware) before() {
	m.start = time.Now()
	m.path = m.r.URL.Path
	m.method = m.r.Method
	m.host = m.r.Host
	m.userAgent = m.r.UserAgent()
	m.ip = m.r.RemoteAddr

	// dump request body
	m.br = newBodyReader(m.r.Body, RequestBodyMaxSize, m.config.WithRequestBody)
	m.r.Body = m.br

	// dump response body
	m.bw = newBodyWriter(m.w, ResponseBodyMaxSize, m.config.WithResponseBody)

	// apply filters early
	for _, filter := range m.config.Filters {
		if !filter(m.bw, m.r) {
			m.callFunc(m.bw, m.r)
			return
		}
	}

	// trace id
	ctx := m.r.Context()
	m.traceID = TraceIDFromRequest(m.r)
	if m.traceID == EmptyTraceID {
		m.traceID = NewTraceID()
		m.returnTraceID = true
	}

	// span id
	m.spanID = SpanIDFromRequest(m.r)

	if !m.config.NoTraceContext {
		ctx = WithTraceID(ctx, m.traceID)
		ctx = WithSpanID(ctx, m.spanID)
		m.r = m.r.WithContext(ctx)
	}
}

func (m *middleware) after() {
	errorSet := true
	if m.status == 0 {
		m.status = m.bw.Status()
		errorSet = false
	}
	end := time.Now()
	latency := end.Sub(m.start)

	// add trace id to response header
	if m.returnTraceID && !m.config.NoTraceContext {
		m.w.Header().Add(TraceHTTPHeaderName, m.traceID.String())
	}

	// build attributes
	baseAttributes := []slog.Attr{}

	if m.config.WithTraceID && !m.config.NoTraceContext {
		baseAttributes = append(baseAttributes, slog.String(TraceIDKey, m.traceID.String()))
	}

	if m.config.WithSpanID && !m.config.NoTraceContext {
		baseAttributes = append(baseAttributes, slog.String(SpanIDKey, m.spanID.String()))
	}

	requestAttributes := []slog.Attr{
		slog.Time("time", m.start.UTC()),
		slog.String("method", m.method),
		slog.String("host", m.host),
		slog.String("path", m.path),
		slog.String("ip", m.ip),
	}

	responseAttributes := []slog.Attr{
		slog.Time("time", end.UTC()),
		slog.Duration("latency", latency),
		slog.Int("status", m.status),
	}

	// request body
	requestAttributes = append(requestAttributes, slog.Int("length", m.br.bytes))
	if m.config.WithRequestBody {
		requestAttributes = append(requestAttributes, slog.String("body", m.br.body.String()))
	}

	// request headers
	if m.config.WithRequestHeader {
		kv := []any{}

		for k, v := range m.r.Header {
			if _, found := HiddenRequestHeaders[strings.ToLower(k)]; found {
				continue
			}
			kv = append(kv, slog.Any(k, v))
		}

		requestAttributes = append(requestAttributes, slog.Group("header", kv...))
	}

	if m.config.WithUserAgent {
		requestAttributes = append(requestAttributes, slog.String("user-agent", m.userAgent))
	}

	// response body
	if !errorSet {
		// only append length if status was not set externally (e.g. via echo error)
		responseAttributes = append(responseAttributes, slog.Int("length", m.bw.bytes))
		if m.config.WithResponseBody {
			responseAttributes = append(responseAttributes, slog.String("body", m.bw.body.String()))
		}
	}

	// response headers
	if m.config.WithResponseHeader {
		kv := []any{}

		for k, v := range m.w.Header() {
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

	level := m.config.DefaultLevel
	if m.status >= http.StatusInternalServerError {
		level = m.config.ServerErrorLevel
	} else if m.status >= http.StatusBadRequest && m.status < http.StatusInternalServerError {
		level = m.config.ClientErrorLevel
	}

	m.logger.LogAttrs(m.r.Context(), level, strconv.Itoa(m.status)+": "+http.StatusText(m.status), attributes...)
}
