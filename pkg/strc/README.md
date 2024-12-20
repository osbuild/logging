## strc

A simple tracing library. When OpenTelemetry is a bit too much. Features:

* Simple code instrumentation.
* Serialization into `slog` (structured logging library).
* Handler with multiple subhandlers for futher processing.
* Simple exporting slog handler for callback-based exporters.
* Adding "trace_id" root field to all logs for easy correlation between logs and traces.

### Instrumentation

Add the following code to function you want to trace:

```go
span, ctx := strc.StartContext(ctx, "span name")
defer span.End()
```

Optionally, additional "savepoints" can be inserted:

```go
span.Event("an event")
```

### Result

All results are stored in `log/slog` records. Each span creates one record with group named `span` with the following data:

* `span.name`: span name 
* `span.id`: span ID (child spands are concatenated with `.`) 
* `span.trace`: trace ID 
* `span.event`: event name (only on event) 
* `span.at`: duration wihtin a span (only on event) 
* `span.duration`: trace duration (only when span ends) 
* `span.time`: log time (can be enabled in exporter) 

Spans natually end up in log sink too, for better readability, the following fields are added to the root namespace:

* `msg`: message in the form `span XXX started` or `event XXX`
* `trace_id`: correlation ID added to all logging messages (not only traces)

Adding the correlation ID can be turned off by setting `strc.TraceIDFieldKey = ""`.

### Propagation

A simple HTTP header-based propagation API is available. Note this is not meant to be used directly, there is HTTP middleware and client wrapper available:

```go
// create new trace id
id := strc.NewTraceID()

// create a new context with trace id value
ctx := strc.WithContext(context.Background(), id)

// fetch the id from the context
id := strc.TraceIDFromContext(ctx)

// add the id from context to a HTTP request
strc.AddTraceIDHeader(ctx, request)

// fetch the id from a request
id := TraceIDFromRequest(request)
```

### Middleware

A standard library HTTP middleware is provided for trace propagation

```go
middleware := strc.NewMiddleware(logger)
mux := http.NewServeMux()
mux.Handle("/", middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// ...
})))
```

The library does not provide native Echo middleware, but a standard middleware can be easily used from Echo:

```go
middleware := strc.NewMiddleware(logger)
e.Use(echo.WrapMiddleware(middleware))
```

The `strc.Middeware` can be optionally configured to log detailed debug information like request or reply HTTP headers or even full body. This is turned off by default, see `strc.MiddlewareConfig` for more info.

### HTTP client

A `TracingDoer` type can be used to decorate HTTP clients adding necessary propagation automatically as long as tracing information is in the request context:

```go
r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://home.zapletalovi.com/", nil)
doer := strc.NewTracingDoer(http.DefaultClient)
doer.Do(r)
```

The `TracingDoer` can be optionally configured to log detailed debug information like request or reply HTTP headers or even full body. This is turned off by default, see `strc.TracingDoerConfig` for more info.

### Full example

```go
package main

import (
	"context"
	"log/slog"

	"github.com/osbuild/logging/pkg/strc"
)

func subProcess(ctx context.Context) {
	span, ctx := strc.StartContext(ctx, "subProcess")
	defer span.End()

	span.Event("an event")
}

func process(ctx context.Context) {
	span, ctx := strc.StartContext(ctx, "process")
	defer span.End()

	subProcess(ctx)
}

func main() {
	// tracing logs via DebugLevel by default
	strc.Level = slog.LevelInfo

	span, ctx := strc.StartContext(context.Background(), "main")
	defer span.End()

	process(ctx)
}
```

Run the example with the following command:

```
go run github.com/osbuild/logging/internal/example_print/
```

Which prints something like (removed time and log level for readability):

```
span main started span.name=main span.id=qXiOgxhiBYkm span.trace=AHaORhHAsMmqqNRF span.source=example_print/main.go:28
span process started span.name=process span.id=qXiOgxhiBYkm.QWBZAYZgymiY span.trace=AHaORhHAsMmqqNRF span.source=example_print/main.go:18
span subProcess started span.name=subProcess span.id=qXiOgxhiBYkm.QWBZAYZgymiY.dLuGWZTAeFsP span.trace=AHaORhHAsMmqqNRF span.source=example_print/main.go:11
span subProcess event an event span.name=subProcess span.id=qXiOgxhiBYkm.QWBZAYZgymiY.dLuGWZTAeFsP span.trace=AHaORhHAsMmqqNRF span.event="an event" span.at=27.342µs span.source=example_print/main.go:14
span subProcess finished in 74.529µs span.name=subProcess span.id=qXiOgxhiBYkm.QWBZAYZgymiY.dLuGWZTAeFsP span.trace=AHaORhHAsMmqqNRF span.dur=65.397µs span.source=example_print/main.go:15
span process finished in 131.065µs span.name=process span.id=qXiOgxhiBYkm.QWBZAYZgymiY span.trace=AHaORhHAsMmqqNRF span.dur=122.761µs span.source=example_print/main.go:22
span main finished in 377.097µs span.name=main span.id=qXiOgxhiBYkm span.trace=AHaORhHAsMmqqNRF span.dur=369.847µs span.source=example_print/main.go:32
```

### Exporting data

While the main goal of this library is just instrumenting and sending data into `slog`, a simple function callback exporter handler is provided by the package for quick collecting or exporting capabilities as well as multi-handler for chaining handlers which is useful to keep sending standard logging data to logging systems. It uses `slog.Attr` type as the data carrier:

```
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/osbuild/logging/pkg/strc"
)

func exportFunc(ctx context.Context, attrs []slog.Attr) {
	for _, attr := range attrs {
		println("exporting trace data", attr.Key, attr.Value.String())
	}
}

func main() {
	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	exportHandler := strc.NewExportHandler(exportFunc)
	multiHandler := strc.NewMultiHandler(textHandler, exportHandler)

	logger := slog.New(multiHandler)
	slog.SetDefault(logger)
	strc.SetLogger(logger)

	span, _ := strc.StartContext(context.Background(), "main")
	defer span.End()
}
```

Run the example with the following command:

```
go run github.com/osbuild/logging/internal/example_export/
```

For the best performance, we a dedicated exporting handler should be written customized to the output format. For more info, see [writing an slog handler](https://pkg.go.dev/log/slog#hdr-Writing_a_handler).
