package strc

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	echoproxy "github.com/osbuild/logging/pkg/echo"
)

func NewEchoV4MiddlewareWithConfig(logger *slog.Logger, config MiddlewareConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			m := middleware{
				logger: logger,
				config: config,
				w:      c.Response().Writer,
				r:      c.Request(),
				callFunc: func(w http.ResponseWriter, r *http.Request) {
					err := next(c)
					httpErr := new(echo.HTTPError)
					if errors.As(err, &httpErr) {
						var intStr string
						if httpErr.Internal != nil {
							intStr = httpErr.Internal.Error()
						}
						logger.WarnContext(c.Request().Context(), "http error",
							slog.Int("code", httpErr.Code),
							slog.String("internal", intStr),
							slog.Any("message", httpErr.Message),
						)
					}
				},
			}

			// Set per-request logger to the default slog instance with context
			c.SetLogger(echoproxy.NewProxyWithContextFor(slog.Default(), c.Request().Context()))

			m.before()
			defer m.after()

			c.Response().Writer = m.bw

			err := next(c)
			httpErr := new(echo.HTTPError)
			if errors.As(err, &httpErr) {
				m.status = httpErr.Code
				m.internal = httpErr.Internal
				m.message = httpErr.Message
				var intStr string
				if m.internal != nil {
					intStr = m.internal.Error()
				}
				logger.WarnContext(c.Request().Context(), "http error",
					slog.Int("code", httpErr.Code),
					slog.String("internal", intStr),
					slog.Any("message", httpErr.Message),
				)
			}

			return err
		}
	}
}
