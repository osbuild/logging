## Common logging code for osbuild projects

Common code for:

* logging
* tracing
* exporting
* correlation

Documentation available in the package directories.

### splunk - high-performance slog handler for Splunk

See [example_splunk](internal/example_splunk/main.go) for a fully working example.

See [splunk](pkg/splunk) package for more info.

### strc - simple tracing via slog

See [example_print](internal/example_print/main.go) and [example_export](internal/example_export/main.go) for fully working examples.

See [strc](pkg/strc) package for more info.

### logrus - proxy to slog

See [example_logrus](internal/example_logrus/main.go) for a fully working example.

See [logrus](pkg/logrus) package for more info.

## Example application

See [example_echo](internal/example_echo/main.go) for a fully working example.

See [slogecho](pkg/slogecho) package for more info.

## TODO

* THIS IS A WORK IN PROGRESS DO NOT USE IT YET.
* Trace/Span ID HTTP propagation does not work yet.
* Document echo example and output.
* Echo logging proxy.
* Propagation tests.
* Fix tests.
