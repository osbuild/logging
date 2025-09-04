package strc_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/osbuild/logging/pkg/collect"
	"github.com/osbuild/logging/pkg/strc"
)

func TestEchoMiddlewareLogging(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(logHandler)

	e := echo.New()
	e.Use(strc.NewEchoV4MiddlewareWithConfig(logger, strc.MiddlewareConfig{}))
	e.GET("/x", func(c echo.Context) error {
		return c.String(http.StatusAccepted, "OK")
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/x", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if !logHandler.Contains("202: Accepted", slog.MessageKey) {
		t.Errorf("Message not found: %s", logHandler.All())
	}

	if !logHandler.Contains("GET", "request", "method") {
		t.Errorf("Request method not found: %s", logHandler.All())
	}

	if !logHandler.Contains("/x", "request", "path") {
		t.Errorf("Request path not found: %s", logHandler.All())
	}

	if !logHandler.Contains("example.com", "request", "host") {
		t.Errorf("Request host not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(0), "request", "length") {
		t.Errorf("Request length not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(202), "response", "status") {
		t.Errorf("Response status not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(2), "response", "length") {
		t.Errorf("Response length not found: %s", logHandler.All())
	}
}

func TestEchoMiddlewareErrorLogging(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(logHandler)

	e := echo.New()
	e.Use(strc.NewEchoV4MiddlewareWithConfig(logger, strc.MiddlewareConfig{}))
	e.GET("/x", func(c echo.Context) error {
		return echo.NewHTTPError(http.StatusInternalServerError, "test error")
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/x", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if !logHandler.Contains("500: Internal Server Error", slog.MessageKey) {
		t.Errorf("Message not found: %s", logHandler.All())
	}

	if !logHandler.Contains("GET", "request", "method") {
		t.Errorf("Request method not found: %s", logHandler.All())
	}

	if !logHandler.Contains("/x", "request", "path") {
		t.Errorf("Request path not found: %s", logHandler.All())
	}

	if !logHandler.Contains("example.com", "request", "host") {
		t.Errorf("Request host not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(0), "request", "length") {
		t.Errorf("Request length not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(500), "response", "status") {
		t.Errorf("Response status not found: %s", logHandler.All())
	}

	if !slices.ContainsFunc(logHandler.All(), func(e map[string]any) bool {
		if e["msg"] != "500: Internal Server Error" {
			return false
		}
		if r, ok := e["request"]; ok {
			r := r.(map[string]any)
			return r["method"] == "GET" && r["path"] == "/x" && r["host"] == "example.com" && r["length"] == int64(0)
		}
		if r, ok := e["response"]; !ok {
			r := r.(map[string]any)
			return r["status"] == 500 && r["length"] == int64(0)
		}
		return true
	}) {
		t.Log(logHandler.All())
		t.Error("Log message not found")
	}
}

func TestEchoMiddlewareFilter(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(logHandler)

	e := echo.New()
	e.Use(strc.NewEchoV4MiddlewareWithConfig(logger, strc.MiddlewareConfig{
		Filters: []strc.Filter{strc.IgnorePathPrefix("/metrics")},
	}))
	e.GET("/metrics", func(c echo.Context) error {
		return c.String(http.StatusOK, "metrics")
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/metrics", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if logHandler.Count() != 0 {
		t.Errorf("Log entries found, expected none: %s", logHandler.All())
	}

	if rec.Body != nil && rec.Body.String() != "metrics" {
		t.Errorf("Response body not as expected: %s", rec.Body.String())
	}
}
