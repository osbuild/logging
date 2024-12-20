## Common logging code for osbuild projects

Common code for:

* logging
* tracing
* exporting
* correlation

Documentation available in the package directories.

### splunk - high-performance slog handler for Splunk

See [example_splunk](internal/example_splunk/main.go) for a fully working example. To see it in action:

```
go run github.com/osbuild/logging/internal/example_splunk/
```

See [splunk](pkg/splunk) package for more info.

### strc - simple tracing via slog

See [example_print](internal/example_print/main.go) and [example_export](internal/example_export/main.go) for fully working examples. To see it in action:

```
go run github.com/osbuild/logging/internal/example_print/
```

See [strc](pkg/strc) package for more info.

### logrus - proxy to slog

See [example_logrus](internal/example_logrus/main.go) for a fully working example. To see it in action:

```
go run github.com/osbuild/logging/internal/example_logrus/
```

See [logrus](pkg/logrus) package for more info.

### slogecho - tracing and logging middleware for Echo

See [example_echo](internal/example_echo/main.go) for a fully working example. It utilizes all the packages above. To see it in action:

```
go run github.com/osbuild/logging/internal/example_echo/
```

See [slogecho](pkg/slogecho) package for more info.

## AUTHORS and LICENSE

License: MIT

Some code in `splunk` was borrowed from https://github.com/osbuild/osbuild-composer and some code in `slogecho` is from https://github.com/samber/slog-echo - thank you.

## TODO

* Span ID must be struct (array of bytes, fixed length) with parent.child structure. Currently it concatenates the whole call stack which is dangerous and ineffective.
* Echo logging proxy.
* Tracing must be off by default.
* Add stdlib middleware and rename slogecho to strcware
* Fix tests.
* Rebuild examples in READMEs.
