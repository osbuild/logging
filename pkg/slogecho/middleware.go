package slogecho

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/osbuild/logging/pkg/strc"
)

var (
	TraceIDKey         = "trace_id"
	SpanIDKey          = "span_id"
	RequestIDKey       = "request_id"
	CustomRequestIDKey = "custom_id"

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

type Config struct {
	DefaultLevel     slog.Level
	ClientErrorLevel slog.Level
	ServerErrorLevel slog.Level

	WithUserAgent      bool
	WithRequestID      bool
	WithRequestBody    bool
	WithRequestHeader  bool
	WithResponseBody   bool
	WithResponseHeader bool

	CustomHeaderXRequestID string

	Filters []Filter
}

// New returns a echo.MiddlewareFunc (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func New(logger *slog.Logger) echo.MiddlewareFunc {
	return NewWithConfig(logger, Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithUserAgent:      true,
		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,

		CustomHeaderXRequestID: "X-Rh-Edge-Request-Id",

		Filters: []Filter{},
	})
}

// NewWithFilters returns a echo.MiddlewareFunc (middleware) that logs requests using slog.
//
// Requests with errors are logged using slog.Error().
// Requests without errors are logged using slog.Info().
func NewWithFilters(logger *slog.Logger, filters ...Filter) echo.MiddlewareFunc {
	return NewWithConfig(logger, Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelWarn,
		ServerErrorLevel: slog.LevelError,

		WithUserAgent:      false,
		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  false,
		WithResponseBody:   false,
		WithResponseHeader: false,

		Filters: filters,
	})
}

// NewWithConfig returns a echo.HandlerFunc (middleware) that logs requests using slog.
func NewWithConfig(logger *slog.Logger, config Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			req := c.Request()
			res := c.Response()
			start := time.Now()
			path := req.URL.Path
			query := req.URL.RawQuery
			method := req.Method
			host := req.Host
			msg := method + " " + "request"
			level := config.DefaultLevel

			params := map[string]string{}
			for i, k := range c.ParamNames() {
				params[k] = c.ParamValues()[i]
			}

			// dump request body if enabled
			br := newBodyReader(req.Body, RequestBodyMaxSize, config.WithRequestBody)
			req.Body = br

			// dump response body if enabled
			bw := newBodyWriter(res.Writer, ResponseBodyMaxSize, config.WithResponseBody)
			res.Writer = bw

			// tracing
			traceID := strc.TraceIDFromRequest(req)
			if traceID == "" {
				traceID = strc.NewTraceID()
			}
			c.SetRequest(c.Request().WithContext(strc.WithTraceID(c.Request().Context(), traceID)))

			spanID := strc.SpanIDFromRequest(req)
			span, ctxWithSpan := strc.StartContext(c.Request().Context(), msg)
			c.SetRequest(c.Request().WithContext(ctxWithSpan))
			defer span.End()

			err = next(c)

			if err != nil {
				if _, ok := err.(*echo.HTTPError); !ok {
					err = echo.
						NewHTTPError(http.StatusInternalServerError).
						WithInternal(err)
					c.Error(err)
				}
			}

			status := res.Status
			route := c.Path()
			end := time.Now()
			latency := end.Sub(start)
			userAgent := req.UserAgent()
			ip := c.RealIP()
			referer := c.Request().Referer()

			errMsg := err

			var httpErr *echo.HTTPError
			if err != nil && errors.As(err, &httpErr) {
				status = httpErr.Code
				if msg, ok := httpErr.Message.(string); ok {
					errMsg = errors.New(msg)
				}
			}

			baseAttributes := []slog.Attr{}

			requestAttributes := []slog.Attr{
				slog.Time("time", start.UTC()),
				slog.String("method", method),
				slog.String("host", host),
				slog.String("path", path),
				slog.String("route", route),
				slog.String("ip", ip),
			}

			if len(params) > 0 {
				requestAttributes = append(requestAttributes, slog.Any("params", params))
			}

			if len(query) > 0 {
				requestAttributes = append(requestAttributes, slog.String("query", query))
			}

			if len(referer) > 0 {
				requestAttributes = append(requestAttributes, slog.String("referer", referer))
			}

			responseAttributes := []slog.Attr{
				slog.Time("time", end.UTC()),
				slog.Duration("latency", latency),
				slog.Int("status", status),
			}

			if config.WithRequestID {
				requestID := req.Header.Get(echo.HeaderXRequestID)
				if requestID == "" {
					requestID = res.Header().Get(echo.HeaderXRequestID)
				}
				if requestID != "" {
					baseAttributes = append(baseAttributes, slog.String(RequestIDKey, requestID))
				}

				requestID = req.Header.Get(config.CustomHeaderXRequestID)
				if requestID == "" {
					requestID = res.Header().Get(config.CustomHeaderXRequestID)
				}
				if requestID != "" {
					baseAttributes = append(baseAttributes, slog.String(CustomRequestIDKey, requestID))
				}
			}

			// tracing
			requestAttributes = append(requestAttributes, slog.String(TraceIDKey, traceID), slog.String(SpanIDKey, spanID))

			// request body
			requestAttributes = append(requestAttributes, slog.Int("length", br.bytes))
			if config.WithRequestBody {
				requestAttributes = append(requestAttributes, slog.String("body", br.body.String()))
			}

			// request headers
			if config.WithRequestHeader {
				kv := []any{}

				for k, v := range c.Request().Header {
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

				for k, v := range c.Response().Header() {
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
				if !filter(c) {
					return
				}
			}

			if status >= http.StatusInternalServerError {
				level = config.ServerErrorLevel
				if err != nil {
					msg = errMsg.Error()
				} else {
					msg = http.StatusText(status)
				}
			} else if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
				level = config.ClientErrorLevel
				if err != nil {
					msg = errMsg.Error()
				} else {
					msg = http.StatusText(status)
				}
			}

			if httpErr != nil {
				attributes = append(
					attributes,
					slog.Any("error", map[string]any{
						"code":     httpErr.Code,
						"message":  httpErr.Message,
						"internal": httpErr.Internal,
					}),
				)

				if httpErr.Internal != nil {
					attributes = append(attributes, slog.String("internal", httpErr.Internal.Error()))
				}
			}

			logger.LogAttrs(c.Request().Context(), level, msg, attributes...)

			return
		}
	}
}
