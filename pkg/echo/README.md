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

For a full web example run:

```
go run github.com/osbuild/logging/internal/example_web/
```
