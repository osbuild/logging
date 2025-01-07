package strc

import (
	"log/slog"
	"net/http"
	"strings"
)

// DoerErr is a simple wrapped error without any message. Additional message would
// stack for each request as multiple doers are called leading to:
//
// "error in doer1: error in doer2: error in doer3: something happened"
type DoerErr struct {
	Err error
}

func NewDoerErr(err error) *DoerErr {
	return &DoerErr{Err: err}
}

func (e *DoerErr) Error() string {
	return e.Err.Error()
}

func (e *DoerErr) Unwrap() error {
	return e.Err
}

type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// TracingDoer is a http client doer that adds tracing to the request and response.
type TracingDoer struct {
	doer   HttpRequestDoer
	config TracingDoerConfig
}

type TracingDoerConfig struct {
	WithRequestBody     bool
	WithResponseBody    bool
	WithRequestHeaders  bool
	WithResponseHeaders bool
}

// NewTracingDoer returns a new TracingDoer.
func NewTracingDoer(doer HttpRequestDoer) *TracingDoer {
	client := TracingDoer{
		doer: doer,
	}
	return &client
}

func NewTracingDoerWithConfig(doer HttpRequestDoer, config TracingDoerConfig) *TracingDoer {
	client := TracingDoer{
		doer:   doer,
		config: config,
	}
	return &client
}

func (td *TracingDoer) Do(req *http.Request) (*http.Response, error) {
	span, ctx := Start(req.Context(), "http client request")
	defer span.End()

	logger := slog.Default().WithGroup("client").With(
		slog.String("method", req.Method),
		slog.String("url", req.URL.RequestURI()),
	)

	// add tracing
	AddTraceIDHeader(ctx, req)
	AddSpanIDHeader(ctx, req)

	// dump request body if enabled
	if td.config.WithRequestBody && req.Body != nil {
		br := newBodyReader(req.Body, RequestBodyMaxSize, td.config.WithRequestBody)
		req.Body = br

		logger.WithGroup("request").With(
			slog.Int64("length", req.ContentLength),
			slog.String("body", br.body.String()),
		)
	}

	if td.config.WithRequestHeaders {
		kv := []any{}
		for k, v := range req.Header {
			if _, found := HiddenRequestHeaders[strings.ToLower(k)]; found {
				continue
			}
			kv = append(kv, slog.Any(k, v))
		}
		logger = logger.WithGroup("request").With(kv...)
	}

	if td.config.WithRequestBody || td.config.WithRequestHeaders {
		logger.DebugContext(ctx, "http client request")
	}

	// delegate the request
	res, err := td.doer.Do(req)
	if err != nil {
		return nil, NewDoerErr(err)
	}

	// dump request body if enabled
	if td.config.WithResponseBody && res.Body != nil {
		br := newBodyReader(req.Body, ResponseBodyMaxSize, td.config.WithRequestBody)
		req.Body = br

		logger.WithGroup("response").With(
			slog.Int64("length", res.ContentLength),
			slog.String("body", br.body.String()),
			slog.Int("status", res.StatusCode),
		)
	}

	if td.config.WithResponseHeaders {
		kv := []any{}
		for k, v := range res.Header {
			if _, found := HiddenResponseHeaders[strings.ToLower(k)]; found {
				continue
			}
			kv = append(kv, slog.Any(k, v))
		}
		logger = logger.WithGroup("response").With(kv...)
	}

	if td.config.WithResponseBody || td.config.WithResponseHeaders {
		logger.DebugContext(ctx, "http client response")
	}

	return res, nil
}
