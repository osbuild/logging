package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	echoproxy "github.com/osbuild/logging/pkg/echo"
	"github.com/osbuild/logging/pkg/logrus"
	"github.com/osbuild/logging/pkg/strc"
)

// for better readability
func cleaner(groups []string, a slog.Attr) slog.Attr {
	if a.Key == "time" || a.Key == "level" {
		return slog.Attr{}
	}
	return a
}

func subProcess(ctx context.Context) {
	span, _ := strc.Start(ctx, "subProcess")
	defer span.End()

	span.Event("an event")
}

var pairs = []strc.HeadfieldPair{
	{HeaderName: "X-Request-Id", FieldName: "request_id"},
}

func startServers(logger *slog.Logger) (*echo.Echo, *echo.Echo, *http.Server) {
	m1 := strc.NewMiddleware(logger)
	m2 := strc.NewMiddlewareWithConfig(logger, strc.MiddlewareConfig{
		WithUserAgent:      true,
		WithRequestBody:    true,
		WithRequestHeader:  true,
		WithResponseBody:   true,
		WithResponseHeader: true,
		Filters:            []strc.Filter{strc.IgnorePathPrefix("/metrics")},
	})
	m3 := strc.HeadfieldPairMiddleware(pairs)

	s1 := echo.New()
	s1.HideBanner = true
	s1.HidePort = true
	s1.Logger = echoproxy.NewProxyFor(logger)
	s1.Use(echo.WrapMiddleware(m1))
	s1.Use(echo.WrapMiddleware(m3))
	s1.GET("/", func(c echo.Context) error {
		span, ctx := strc.Start(c.Request().Context(), "s1")
		defer span.End()

		subProcess(ctx)

		slog.DebugContext(ctx, "slog msg", "service", "s1")
		logrus.WithContext(ctx).WithField("service", "s1").Debug("logrus msg")
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
	s2.Use(echo.WrapMiddleware(m1))
	s2.GET("/", func(c echo.Context) error {
		span, ctx := strc.Start(c.Request().Context(), "s2")
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
		span, ctx := strc.Start(r.Context(), "s3")
		defer span.End()

		subProcess(ctx)
	})
	mux3.Handle("/", m2(h3))
	go srv3.ListenAndServe()

	return s1, s2, srv3
}

func main() {
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: cleaner})
	logger := slog.New(strc.NewMultiHandlerCustom(nil, strc.HeadfieldPairCallback(pairs), h))
	slog.SetDefault(logger)
	strc.SetLogger(logger)
	strc.SkipSource = true // for better readability
	logrus.SetDefault(logrus.NewProxyFor(logger, logrus.Options{NoExit: true}))

	s1, s2, s3 := startServers(logger)
	defer s1.Close()
	defer s2.Close()
	defer s3.Close()

	client := http.Client{}
	req, _ := http.NewRequest("GET", "http://localhost:8131/", nil)
	req.Header.Add("X-Request-Id", "abcdef")
	client.Do(req)
}
