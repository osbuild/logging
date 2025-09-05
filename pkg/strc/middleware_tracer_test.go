package strc_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/osbuild/logging/pkg/collect"
	"github.com/osbuild/logging/pkg/strc"
)

func TestEchoTracer(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(logHandler)
	e := echo.New()

	e.Use(strc.EchoTracer())
	e.Use(strc.EchoRequestLogger(logger, strc.MiddlewareConfig{}))

	var traceID string
	e.GET("/ok", func(c echo.Context) error {
		traceID = strc.TraceIDFromContext(c.Request().Context()).String()
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.True(t, logHandler.Contains(traceID, strc.TraceIDKey))
}
