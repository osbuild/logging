## strc

A simple tracing library. When OpenTelemetry is a bit too much. Features:

* Simple code instrumentation.
* Serialization into `slog` (structured logging library).
* No exporter, collector or UI - build your own.
* Build your own exporter / collector based on `slog` Handler API.
* The API is fixed - no breaking changes accepted.

### Provided tracing data

* `msg`: message in the form `span XXX started`
* `span.name`: span name 
* `span.trace`: trace ID 
* `span.id`: span ID (child spands are concatenated with `.`) 
* `span.event`: event name (only on event) 
* `span.at`: duration wihtin a span (only on event) 
* `span.duration`: trace duration (only when span ends) 

### How to use

```go
func subProcess(ctx context.Context) {
	span, ctx := strc.StartContext(ctx, "subProcess")
	defer span.End()
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
go run github.com/lzap/strc/internal/example_print/
```

Which prints something like:

```
2024/12/05 20:11:14 INFO span main started span.name=main span.trace=zXoiZVRQUxzlZVOKnvzx span.id=tJKMgdiAuYUQ
2024/12/05 20:11:14 INFO span process started span.name=process span.trace=zXoiZVRQUxzlZVOKnvzx span.id=tJKMgdiAuYUQ.NvoqTWHwmUZR
2024/12/05 20:11:14 INFO span subProcess started span.name=subProcess span.trace=zXoiZVRQUxzlZVOKnvzx span.id=tJKMgdiAuYUQ.NvoqTWHwmUZR.NpJCfmuObJPw
2024/12/05 20:11:14 INFO span subProcess finished span.name=subProcess span.duration=4.725µs span.trace=zXoiZVRQUxzlZVOKnvzx span.id=tJKMgdiAuYUQ.NvoqTWHwmUZR.NpJCfmuObJPw
2024/12/05 20:11:14 INFO span process finished span.name=process span.duration=20.785µs span.trace=zXoiZVRQUxzlZVOKnvzx span.id=tJKMgdiAuYUQ.NvoqTWHwmUZR
2024/12/05 20:11:14 INFO span main finished span.name=main span.duration=120.108µs span.trace=zXoiZVRQUxzlZVOKnvzx span.id=tJKMgdiAuYUQ
```

### Exporting data

The goal of this library is just instrumenting and sending data into `slog`. You need to implement your own exporter or collector in order to do anything meaningful. Luckily, [writing an slog handler is easy](https://pkg.go.dev/log/slog#hdr-Writing_a_handler). For handler chaining (data fanout), make sure to check out the excellent [slog-multi](https://github.com/samber/slog-multi) library.

## TODO

* PC modification for file/line

### Projects using strc

* Let us know!
