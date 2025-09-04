package strc

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echoproxy "github.com/osbuild/logging/pkg/echo"
)

func slogAttributesFromRequest(r *http.Request) []slog.Attr {
	attrs := []slog.Attr{
		{
			Key: "request",
			Value: slog.GroupValue(
				[]slog.Attr{
					slog.String("method", r.Method),
					slog.String("host", r.Host),
					slog.String("path", r.URL.Path),
					slog.String("user-agent", r.UserAgent()),
					slog.String("ip", r.RemoteAddr),
					slog.Int64("length", r.ContentLength),
				}...,
			),
		},
	}

	traceID := TraceIDFromContext(r.Context())
	if traceID != EmptyTraceID {
		attrs = append(attrs, slog.String(TraceIDKey, traceID.String()))
	}
	spanID := SpanIDFromContext(r.Context())
	if spanID != EmptySpanID {
		attrs = append(attrs, slog.String(SpanIDKey, spanID.String()))
	}

	return attrs
}

// All errors should be handled by the echo.HTTPErrorHandler configured on the echo server
// itself. But let statusForLogging infer the status at least. Anything else should be handled by
// the error handler.
func statusForLogging(status int, err error) int {
	if err == nil {
		return status
	}
	if echoErr, ok := err.(*echo.HTTPError); ok {
		return echoErr.Code
	}
	return http.StatusInternalServerError
}

// This generates exactly one log statement per request processed.
//
// Meant to be chained after middlewares that add fields to the request context.
func EchoRequestLogger(logger *slog.Logger, config MiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, filter := range config.Filters {
				if !filter(nil, c.Request()) {
					return next(c)
				}
			}

			start := time.Now()
			err := next(c)
			latency := time.Since(start)

			var attrs []slog.Attr
			status := statusForLogging(c.Response().Status, err)
			level := config.DefaultLevel
			if status >= http.StatusInternalServerError {
				level = config.ServerErrorLevel
				if err != nil {
					attrs = append(attrs, slog.String("error", err.Error()))
				}
			} else if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
				level = config.ClientErrorLevel
			}

			attrs = append(
				attrs,
				slog.Time("time", start.UTC()),
				slog.Duration("latency", latency),
			)
			attrs = append(attrs, slogAttributesFromRequest(c.Request())...)
			attrs = append(
				attrs,
				slog.Attr{
					Key: "response",
					Value: slog.GroupValue(
						[]slog.Attr{
							slog.Int64("length", c.Response().Size),
							slog.Int("status", status),
						}...,
					),
				},
			)

			logger.LogAttrs(c.Request().Context(), level, fmt.Sprintf("%d: %s", status, http.StatusText(status)), attrs...)
			return err
		}
	}
}

// This sets the logger for each request to the specified logger. Anything processing the
// cecho.Context can just call echo.Context.Logger() to get the appropriate logger.
//
// Meant to be chained after middlewares that add fields to the request context.
func EchoContextSetLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.SetLogger(echoproxy.NewProxyWithContextFor(logger, c.Request().Context()))
			return next(c)
		}
	}
}
