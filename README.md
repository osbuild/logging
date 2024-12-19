## Common logging code for osbuild projects

Common code for:

* logging
* tracing
* exporting
* correlation

Documentation available in the package directories.

### splunk - high-performance slog handler for Splunk

See [internal/example_splunk/main.go](example_splunk) for a fully working example.

See [pkg/splunk](splunk) package for more info.

### strc - simple tracing via slog

See [internal/example_print/main.go](example_print) and [internal/example_export/main.go](example_export) for fully working examples.

See [pkg/strc](strc) package for more info.

### logrus - proxy to slog

See [internal/example_logrus/main.go](example_logrus) for a fully working example.

See [pkg/logrus](logrus) package for more info.

## Example application

See [internal/example_echo/main.go](example_echo) for a fully working example.

See [pkg/slogecho](slogecho) package for more info.

## TODO

* THIS IS A WORK IN PROGRESS DO NOT USE IT YET.
* Trace/Span ID HTTP propagation does not work yet.
* Echo logging proxy.
* Propagation tests.
* Fix tests.
