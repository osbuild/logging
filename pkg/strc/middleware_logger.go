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

			// errors should be handled by echo.HTTPErrorHAndler, but let's grab the status at least.
			// Anthing else should be handled by the error handler.
			status := c.Response().Status
			if err != nil {
				echoErr, ok := err.(*echo.HTTPError)
				if ok {
					status = echoErr.Code
				} else {
					status = http.StatusInternalServerError
				}
			}

			attrs := []slog.Attr{
				slog.Time("time", start.UTC()),
				slog.Duration("latency", latency),
			}
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

			level := config.DefaultLevel
			if status >= http.StatusInternalServerError {
				level = config.ServerErrorLevel
				if err != nil {
					attrs = append(attrs, slog.String("error", err.Error()))
				}
			} else if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
				level = config.ClientErrorLevel
			}
			logger.LogAttrs(c.Request().Context(), level, fmt.Sprintf("%d: %s", status, http.StatusText(status)), attrs...)

			return err
		}
	}
}

func EchoContextSetLogger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.SetLogger(echoproxy.NewProxyWithContextFor(logger, c.Request().Context()))
			return next(c)
		}
	}
}
