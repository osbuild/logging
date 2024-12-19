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

* Some code in `splunk` was borrowed from https://github.com/osbuild/osbuild-composer
* Some code in `slogecho` is from https://github.com/samber/slog-echo

## TODO

* THIS IS A WORK IN PROGRESS DO NOT USE IT YET.
* Trace/Span ID HTTP propagation does not work yet.
* Document echo example and output.
* Echo logging proxy.
* Propagation tests.
* Fix tests.
