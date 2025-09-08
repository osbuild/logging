package strc_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/osbuild/logging/pkg/collect"
	"github.com/osbuild/logging/pkg/strc"
)

func TestEchoRequestLoggerOk(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, true)
	logger := slog.New(logHandler)

	e := echo.New()
	e.Use(strc.EchoRequestLogger(logger, strc.MiddlewareConfig{}))
	e.GET("/ok", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, logHandler.Contains("200: OK", slog.MessageKey))
	assert.True(t, logHandler.Contains("GET", "request", "method"))
	assert.True(t, logHandler.Contains("/ok", "request", "path"))
	assert.True(t, logHandler.Contains("example.com", "request", "host"))
	assert.True(t, logHandler.Contains(int64(0), "request", "length"))
	assert.True(t, logHandler.Contains(int64(200), "response", "status"))
	assert.True(t, logHandler.Contains(int64(2), "response", "length"))
	assert.True(t, logHandler.Contains(slog.LevelInfo.String(), "level"))
}

func TestEchoRequestLoggerHTTPError400(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, true)
	logger := slog.New(logHandler)

	e := echo.New()
	e.Use(strc.EchoRequestLogger(logger, strc.MiddlewareConfig{
		ClientErrorLevel: slog.LevelWarn,
	}))
	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusBadRequest, "bad-request")
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, logHandler.Contains("400: Bad Request", slog.MessageKey))
	assert.True(t, logHandler.Contains("GET", "request", "method"))
	assert.True(t, logHandler.Contains("/error", "request", "path"))
	assert.True(t, logHandler.Contains("example.com", "request", "host"))
	assert.True(t, logHandler.Contains(int64(0), "request", "length"))
	assert.True(t, logHandler.Contains(int64(400), "response", "status"))
	assert.True(t, logHandler.Contains(int64(0), "response", "length"))
	assert.True(t, logHandler.Contains(slog.LevelWarn.String(), "level"))
}

func TestEchoRequestLoggerHTTPError503(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, true)
	logger := slog.New(logHandler)

	e := echo.New()
	e.Use(strc.EchoRequestLogger(logger, strc.MiddlewareConfig{
		ServerErrorLevel: slog.LevelError,
	}))
	e.GET("/error", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "service-unavailable")
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, logHandler.Contains("503: Service Unavailable", slog.MessageKey))
	assert.True(t, logHandler.Contains("GET", "request", "method"))
	assert.True(t, logHandler.Contains("/error", "request", "path"))
	assert.True(t, logHandler.Contains("example.com", "request", "host"))
	assert.True(t, logHandler.Contains(int64(0), "request", "length"))
	assert.True(t, logHandler.Contains(int64(503), "response", "status"))
	assert.True(t, logHandler.Contains(int64(0), "response", "length"))
	assert.True(t, logHandler.Contains(slog.LevelError.String(), "level"))
}

func TestEchoRequestLoggerError(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, true)
	logger := slog.New(logHandler)

	e := echo.New()
	e.Use(strc.EchoRequestLogger(logger, strc.MiddlewareConfig{
		ServerErrorLevel: slog.LevelError,
	}))
	e.GET("/error", func(c echo.Context) error {
		return fmt.Errorf("random error")
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/error", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.True(t, logHandler.Contains("500: Internal Server Error", slog.MessageKey))
	assert.True(t, logHandler.Contains("GET", "request", "method"))
	assert.True(t, logHandler.Contains("/error", "request", "path"))
	assert.True(t, logHandler.Contains("example.com", "request", "host"))
	assert.True(t, logHandler.Contains(int64(0), "request", "length"))
	assert.True(t, logHandler.Contains(int64(500), "response", "status"))
	assert.True(t, logHandler.Contains(int64(0), "response", "length"))
	assert.True(t, logHandler.Contains(slog.LevelError.String(), "level"))
}

func TestEchoContextSetLogger(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(strc.NewMultiHandler(logHandler))

	e := echo.New()
	e.Use(strc.EchoTraceExtractor(), strc.EchoContextSetLogger(logger))
	e.GET("/x", func(c echo.Context) error {
		c.Logger().Debug("echo logger test")
		c.Logger().Debugj(map[string]any{"test": "json"})
		return c.String(http.StatusAccepted, "OK")
	})

	trace_id := "1zapXiHprrrvHqD"
	req := httptest.NewRequest(http.MethodGet, "http://example.com/x", nil)
	req.Header.Add("X-Strc-Trace-Id", trace_id)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if !logHandler.Contains("echo logger test", slog.MessageKey) {
		t.Errorf("Message not found: %s", logHandler.All())
	}

	if !logHandler.Contains("json", "test") {
		t.Errorf("JSON field not found: %s", logHandler.All())
	}

	if !logHandler.Contains(trace_id, "trace_id") {
		t.Errorf("TraceID not found: %s", logHandler.All())
	}
}
