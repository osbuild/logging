package strc

import (
	"log/slog"
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
}
