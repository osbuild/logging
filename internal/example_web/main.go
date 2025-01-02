package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	echoproxy "github.com/osbuild/logging/pkg/echo"
	"github.com/osbuild/logging/pkg/logrus"
	"github.com/osbuild/logging/pkg/splunk"
	"github.com/osbuild/logging/pkg/strc"
)

const splunkURL = "http://localhost:8133/services/collector/event"

// for better readability
func cleaner(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "time" || a.Key == "level" {
		return slog.Attr{}
	}
	return a
}

func subProcess(ctx context.Context) {
	span, ctx := strc.StartContext(ctx, "subProcess")
	defer span.End()

	span.Event("an event")
}

func startServers(logger *slog.Logger) (*echo.Echo, *echo.Echo, *http.Server) {
	middleware := strc.NewMiddleware(logger)
	loggingMiddleware := strc.NewMiddlewareWithConfig(logger, strc.MiddlewareConfig{
		WithUserAgent:      true,
		WithRequestBody:    true,
		WithRequestHeader:  true,
		WithResponseBody:   true,
		WithResponseHeader: true,
		Filters:            []strc.Filter{strc.IgnorePathPrefix("/metrics")},
	})

	s1 := echo.New()
	s1.HideBanner = true
	s1.HidePort = true
	s1.Logger = echoproxy.NewProxyFor(logger)
	s1.Use(echo.WrapMiddleware(middleware))
	s1.GET("/", func(c echo.Context) error {
		span, ctx := strc.StartContext(c.Request().Context(), "s1")
		defer span.End()

		subProcess(ctx)

		slog.DebugContext(ctx, "slog msg", "service", "s1")
		logrus.WithField("service", "s1").Debug("logrus msg")
		c.Logger().Debug("echo msg 1")
		c.Logger().Debugj(map[string]interface{}{"service": "s1", "msg": "echo msg 2"})

		r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8132/", nil)
		doer := strc.NewTracingDoer(http.DefaultClient)
		doer.Do(r)
		return nil
	})
	go s1.Start(":8131")

	s2 := echo.New()
	s2.HideBanner = true
	s2.HidePort = true
	s2.Use(echo.WrapMiddleware(middleware))
	s2.GET("/", func(c echo.Context) error {
		span, ctx := strc.StartContext(c.Request().Context(), "s2")
		defer span.End()

		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8133/", nil)
		doer := strc.NewTracingDoer(http.DefaultClient)
		doer.Do(req)
		return nil
	})
	go s2.Start(":8132")

	mux3 := http.NewServeMux()
	srv3 := &http.Server{
		Addr:    ":8133",
		Handler: mux3,
	}
	h3 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span, ctx := strc.StartContext(r.Context(), "s3")
		defer span.End()

		subProcess(ctx)
	})
	mux3.Handle("/", loggingMiddleware(h3))
	mux3.Handle("/services/collector/event", http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	go srv3.ListenAndServe()

	return s1, s2, srv3
}

func main() {
	hSplunk := splunk.NewSplunkHandler(context.Background(), slog.LevelDebug, splunkURL, "t", "s", "h")
	hStdout := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: cleaner})
	logger := slog.New(strc.NewMultiHandler(hSplunk, hStdout))
	slog.SetDefault(logger)
	strc.SetLogger(logger)
	strc.SkipSource = true // for better readability
	logrus.SetDefault(logrus.NewProxyFor(logger))

	s1, s2, s3 := startServers(logger)
	defer s1.Close()
	defer s2.Close()
	defer s3.Close()
	defer hSplunk.Close() // last in, first out (blocks until all logs are sent)

	http.Get("http://localhost:8131/")
	stats := hSplunk.Statistics()
	fmt.Printf("sent %d events to splunk mock\n", stats.EventCount)
}
