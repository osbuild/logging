## slogecho

Echo middleware adding `slog` info request message and `strc` tracing HTTP header propagation.

### How to use

Just use as a regular echo middleware:

```go
e.Use(slogecho.New(logger))
```

It is possible to filter out specific reqests from being logged:

```go
e.Use(slogecho.NewWithFilters(logger, slogecho.IgnorePathPrefix("/metrics")))
```

It uses `log/slog` for all logging and `strc` for tracing

```go
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/osbuild/logging/pkg/slogecho"
	"github.com/osbuild/logging/pkg/splunk"
	"github.com/osbuild/logging/pkg/strc"
)

const splunkURL = "http://localhost:8132/services/collector/event"

func main() {
	hSplunk := splunk.NewSplunkHandler(context.Background(), slog.LevelDebug, splunkURL, "t", "s", "h")
	hStdout := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(strc.NewMultiHandler(hSplunk, hStdout))
	slog.SetDefault(logger)
	strc.SetLogger(logger)
	defer hSplunk.Close()

    e := echo.New()
    e.Use(slogecho.New(logger))
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})
	e.Logger.Fatal(e.Start(":1323"))
}
```
