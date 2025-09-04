package strc

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func httpRequestWithTracing(r *http.Request) (TraceID, *http.Request) {
	traceID := TraceIDFromRequest(r)
	if traceID == EmptyTraceID {
		traceID = NewTraceID()
	}

	newCtx := WithTraceID(r.Context(), traceID)
	r = r.WithContext(newCtx)

	spanID := SpanIDFromRequest(r)
	if spanID != EmptySpanID {
		newCtx = WithSpanID(r.Context(), spanID)
		r = r.WithContext(newCtx)
	}
	return traceID, r
}

// This extracts trace IDs and span IDs from HTTP headers and sets them
// in the request context.
//
// Meant to be chained before any logging middleware.
func EchoTracer(config MiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, filter := range config.Filters {
				if !filter(nil, c.Request()) {
					return next(c)
				}
			}

			traceID, req := httpRequestWithTracing(c.Request())
			c.SetRequest(req)

			err := next(c)

			c.Response().Header().Add(TraceHTTPHeaderName, traceID.String())
			return err
		}
	}
}
