package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/osbuild/logging/pkg/logrus"
	"github.com/osbuild/logging/pkg/slogecho"
	"github.com/osbuild/logging/pkg/splunk"
	"github.com/osbuild/logging/pkg/strc"
)

const splunkURL = "http://localhost:8132/services/collector/event"

var wg sync.WaitGroup

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

func startServers(logger *slog.Logger) (*echo.Echo, *echo.Echo) {
	e1 := echo.New()
	e1.HideBanner = true
	e1.HidePort = true
	e1.Use(slogecho.New(logger))
	e1.GET("/", func(c echo.Context) error {
		ctx := c.Request().Context()
		defer wg.Done()

		span, ctx := strc.StartContext(ctx, "e1")
		defer span.End()

		subProcess(ctx)

		slog.DebugContext(ctx, "slog msg", "service", "e1")
		logrus.WithField("service", "e1").Debug("logrus msg")
		c.Logger().Debugj(map[string]interface{}{"service": "e1", "msg": "echo msg"})

		r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8132/", nil)
		//doer := strc.NewTracingDoerWithConfig(http.DefaultClient, strc.TracingDoerConfig{true, true, true, true})
		doer := strc.NewTracingDoer(http.DefaultClient)
		doer.Do(r)

		return c.String(http.StatusOK, "OK (e1)")
	})
	go e1.Start(":8131")

	e2 := echo.New()
	e2.HideBanner = true
	e2.HidePort = true
	mw := slogecho.NewWithFilters(logger, slogecho.IgnorePathPrefix("/services/collector/event"))
	e2.Use(mw)
	e2.GET("/", func(c echo.Context) error {
		ctx := c.Request().Context()
		defer wg.Done()

		span, ctx := strc.StartContext(ctx, "e2")
		defer span.End()

		subProcess(ctx)

		slog.With("w1", "v1").DebugContext(ctx, "slog msg", "service", "e1")
		logrus.WithField("service", "e1").Debug("logrus msg")
		c.Logger().Debugj(map[string]interface{}{"service": "e1", "msg": "echo msg"})

		return c.String(http.StatusOK, "OK (e2)")
	})
	e2.POST("/services/collector/event", func(c echo.Context) error {
		return c.String(http.StatusOK, "splunk mock")
	})
	go e2.Start(":8132")

	// TODO http stdlib server

	return e1, e2
}

func main() {
	hSplunk := splunk.NewSplunkHandler(context.Background(), slog.LevelDebug, splunkURL, "t", "s", "h")
	hStdout := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: cleaner})
	logger := slog.New(strc.NewMultiHandler(hSplunk, hStdout))
	slog.SetDefault(logger)
	strc.SetLogger(logger)
	strc.SkipSource = true // for better readability
	logrus.SetDefault(logrus.NewProxyFor(logger))

	e1, e2 := startServers(logger)
	defer e1.Close()
	defer e2.Close()
	defer hSplunk.Close() // last in, first out (blocks until all logs are sent)

	wg.Add(2)
	http.Get("http://localhost:8131/")
	wg.Wait()
	stats := hSplunk.Statistics()
	fmt.Printf("sent %d events to splunk mock\n", stats.EventCount)
}
