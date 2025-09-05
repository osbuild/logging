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
}
