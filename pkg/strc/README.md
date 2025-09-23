## strc

A simple tracing library. When OpenTelemetry is a bit too much. Features:

* Simple code instrumentation.
* Serialization into `log/slog` (structured logging library).
* Handler with multiple sub-handlers for further processing.
* Simple exporting slog handler for callback-based exporters.
* Adding "trace_id" root field to all logs for easy correlation between logs and traces.

### Instrumentation

Add the following code to function you want to trace:

```go
span, ctx := strc.Start(ctx, "span name")
defer span.End()
```

Optionally, additional "save points" can be inserted:

```go
span.Event("an event")
```

### Result

All results are stored in `log/slog` records. Each span creates one record with group named `span` with the following data:

* `span.name`: span name 
* `span.id`: span ID
* `span.parent`: parent ID or `0000000` when there is no parent
* `span.trace`: trace ID 
* `span.event`: event name (only on event) 
* `span.at`: duration within a span (only on event) 
* `span.duration`: trace duration (only when span ends) 
* `span.time`: log time (can be enabled in exporter) 

Spans end up in log sink too, for better readability, the following fields are added to the root namespace:

* `msg`: message in the form `span XXX started` or `event XXX`
* `trace_id` - trace ID (disable by setting `strc.TraceIDFieldKey` to empty string)
* `build_id` - build Git sha (disable by setting `strc.BuildIDFieldKey` to empty string)

### Overriding time

Span start, event and end time is automatically taken via `time.Now()` call but there are some use cases when this needs to be overridden to a specific time. Use special attributes to do that:

```go
span := strc.Start(ctx, "span name", "started", time.Now())
span.Event("an event", "at", time.Now())
span.End("finished", time.Now())
```

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

The library provides native Echo middleware functions:

```go
e.Use(strc.EchoRequestLogger(logger, strc.MiddlewareConfig{}))
```

Available Echo middleware in the preferred order of call:

* `EchoTraceExtractor` extracts `X-Strc-Trace-Id` (and span) headers and stores them in the context. When no trace id is available, a random one is created.
* `EchoContextSetLogger`: overrides the default Echo logger with per-request instance which captures context from the request. This means all logs created via Echo library will be forwarded into `slog` with values from context.
* `EchoHeadersExtractor` extracts custom HTTP headers and stores them in the context. Can be appended to all logs via handler callback, useful for external correlation fields like `request_id` or `edge_id`.
* `EchoRequestLogger`: creates a log record for every single HTTP request with configurable log level.

See `strc.MiddlewareConfig` for more info about configuration.

### HTTP client

A `TracingDoer` type can be used to decorate HTTP clients adding necessary propagation automatically as long as tracing information is in the request context:

```go
r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://home.zapletalovi.com/", nil)
doer := strc.NewTracingDoer(http.DefaultClient)
doer.Do(r)
```

The `TracingDoer` can be optionally configured to log detailed debug information like request or reply HTTP headers or even full body. This is turned off by default, see `strc.TracingDoerConfig` for more info.

Example headers generated or parsed by HTTP client & middleware code:

```
X-Strc-Trace-ID: LOlIxiHprrrvHqD
X-Strc-Span-ID: VIPEcES.yuufaHI
```

### Full example

```go
package main

import (
	"context"
	"log/slog"

	"github.com/osbuild/logging/pkg/strc"
)

func subProcess(ctx context.Context) {
	span, ctx := strc.Start(ctx, "subProcess")
	defer span.End()

	span.Event("an event")
}

func process(ctx context.Context) {
	span, ctx := strc.Start(ctx, "process")
	defer span.End()

	subProcess(ctx)
}

func main() {
	span, ctx := strc.Start(context.Background(), "main")
	defer span.End()

	process(ctx)
}
```

Note the above example will print nothing as tracing is disabled by default, two things must be done. First, a `slog` destination logger must be set:

```
strc.SetLogger(slog.Default())
```

Second, the destination logger must have debug level handing enabled. The default logging level can be increased if needed:

```
strc.Level = slog.LevelInfo
```

Run the example with the following command:

```
go run github.com/osbuild/logging/internal/example_print/
```

Which prints something like (removed time and log level for readability):

```
span main started span.name=main span.id=pEnFDti span.parent=0000000 span.trace=SVyjVloJYogpNPq span.source=main.go:28
span process started span.name=process span.id=fRfWksO span.parent=pEnFDti span.trace=SVyjVloJYogpNPq span.source=main.go:18
span subProcess started span.name=subProcess span.id=gSouhiv span.parent=fRfWksO span.trace=SVyjVloJYogpNPq span.source=main.go:11
span subProcess event an event span.name=subProcess span.id=gSouhiv span.parent=fRfWksO span.trace=SVyjVloJYogpNPq span.event="an event" span.at=21.644µs span.source=main.go:14
span subProcess finished in 47.355µs span.name=subProcess span.id=gSouhiv span.parent=fRfWksO span.trace=SVyjVloJYogpNPq span.dur=47.355µs span.source=main.go:15
span process finished in 94.405µs span.name=process span.id=fRfWksO span.parent=pEnFDti span.trace=SVyjVloJYogpNPq span.dur=94.405µs span.source=main.go:22
span main finished in 285.246µs span.name=main span.id=pEnFDti span.parent=0000000 span.trace=SVyjVloJYogpNPq span.dur=285.246µs span.source=main.go:32
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

	span, _ := strc.Start(context.Background(), "main")
	defer span.End()
}
```

There is additional `NewMultiHandlerCustom` which allows adding custom attributes from context via a callback function. This is useful when additional correlation id (e.g. background job UUID) needs to be added to every single regular log record. The multi-handler creates the following new keys in the root element:

* `trace_id` - trace ID (disable by setting `strc.TraceIDFieldKey` to empty string)
* `build_id` - build Git sha (disable by setting `strc.BuildIDFieldKey` to empty string)

Run the example with the following command:

```
go run github.com/osbuild/logging/internal/example_export/
```

For the best performance, we a dedicated exporting handler should be written customized to the output format. For more info, see [writing an slog handler](https://pkg.go.dev/log/slog#hdr-Writing_a_handler).
