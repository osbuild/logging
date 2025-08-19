## echo

A temporary `echo` proxy to `log/slog`. Only used for the transition period until our projects are fully migrated. Please keep in mind this is an incomplete implementation, only functions used in osbuild project are present.

### How to use

```go
package main

import (
	"github.com/labstack/echo/v4"

	echoproxy "github.com/osbuild/logging/pkg/echo"
)

func main() {
	e := echo.New()
	e.Logger = echoproxy.NewProxy()

	// ...
}
```

### Per-request logger

It is possible to set individual echo proxy for each request which performs context passing into every single logging record automatically. Just make sure to put middleware that sets context fields before echo proxy middleware:

```go
s1 := echo.New()
s1.Logger = echoproxy.NewProxyFor(slog.Default())
// strc middleware sets "trace_id" and "request_id"
s1.Use(echo.WrapMiddleware(strc.NewMiddleware(slog.Default())))
// pass the context with the fields into echo stack
s1.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.SetLogger(echoproxy.NewProxyContextFor(slog.Default(), c.Request().Context()))
		return next(c)
	}
})
```

By default, only `trace_id` and `build_id` are logged by the multi handler, this can be customized by providing a callback function via `NewMultiHandlerCustom`.

### Full example

For a full web example run:

```
go run github.com/osbuild/logging/internal/example_web/
```
