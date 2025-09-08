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

func startServers(logger *slog.Logger) (*echo.Echo, *echo.Echo) {
	tracerMW := strc.EchoTraceExtractor()
	loggerMW := strc.EchoRequestLogger(logger, strc.MiddlewareConfig{})
	setLoggerMW := strc.EchoContextSetLogger(logger)
	headfieldMW := strc.HeadfieldPairMiddleware(pairs)

	s1 := echo.New()
	s1.HideBanner = true
	s1.HidePort = true
	s1.Logger = echoproxy.NewProxyFor(logger)
	s1.Use(
		tracerMW,
		echo.WrapMiddleware(headfieldMW),
		setLoggerMW,
		loggerMW,
	)
	s1.GET("/", func(c echo.Context) error {
		span, ctx := strc.Start(c.Request().Context(), "s1")
		defer span.End()

		subProcess(ctx)

		slog.DebugContext(ctx, "slog msg", "service", "s1")
		logrus.WithContext(ctx).WithField("service", "s1").Debug("logrus msg")
		c.Logger().Debug("echo msg 1")
		c.Logger().Debugj(map[string]any{"service": "s1", "msg": "echo msg 2"})

		r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8132/", nil)
		doer := strc.NewTracingDoer(http.DefaultClient)
		doer.Do(r)
		return c.String(200, "ok")
	})
	go s1.Start(":8131")

	s2 := echo.New()
	s2.HideBanner = true
	s2.HidePort = true
	s2.Use(
		tracerMW,
		echo.WrapMiddleware(headfieldMW),
		setLoggerMW,
		loggerMW,
	)
	s2.GET("/", func(c echo.Context) error {
		span, ctx := strc.Start(c.Request().Context(), "s2")
		defer span.End()

		subProcess(ctx)
		return c.String(200, "ok")
	})
	go s2.Start(":8132")

	return s1, s2
}

func request(req *http.Request, handler slog.Handler) {
	logger := slog.New(strc.NewMultiHandlerCustom(nil, strc.HeadfieldPairCallback(pairs), handler))
	slog.SetDefault(logger)
	strc.SetLogger(logger)
	strc.SkipSource = true // for better readability
	logrus.SetDefault(logrus.NewProxyFor(logger, logrus.Options{NoExit: true}))

	s1, s2 := startServers(logger)
	defer s1.Close()
	defer s2.Close()

	client := http.Client{}
	req.Header.Add("X-Request-Id", "abcdef")
	client.Do(req)
}

func main() {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug, ReplaceAttr: cleaner})
	req, _ := http.NewRequest("GET", "http://localhost:8131/", nil)
	request(req, handler)
}
