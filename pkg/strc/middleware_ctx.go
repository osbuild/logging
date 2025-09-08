package strc

import (
	"context"
	"log/slog"

	"github.com/labstack/echo/v4"
)

// HeaderField is a pair of header name and field name.
type HeaderField struct {
	HeaderName string
	FieldName  string
}

// EchoHeadersExtractor is a middleware that extracts values from headers and
// stores them in the context.
func EchoHeadersExtractor(pairs []HeaderField) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			for _, p := range pairs {
				if value := c.Request().Header.Get(p.HeaderName); value != "" {
					c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), p, value)))
				}
			}

			return next(c)
		}
	}
}

// HeadersCallback is a slog callback that extracts values from the
// context and adds them to the attributes.
func HeadersCallback(pairs []HeaderField) MultiCallback {
	return func(ctx context.Context, a []slog.Attr) ([]slog.Attr, error) {
		for _, p := range pairs {
			if value, ok := ctx.Value(p).(string); ok {
				a = append(a, slog.String(p.FieldName, value))
			}
		}

		return a, nil
	}
}
