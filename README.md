## Common logging code for osbuild projects

Common code for:

* logging
* tracing
* exporting
* correlation

Together with some best practices on how to do logging and tracing in osbuild projects.

### Overview

Without sinking deep into the details of what is in this repo, here is a summary of what are the best practices for osbuild projects. Note these project were already configured with these libraries and use provided functionality like middleware, HTTP client wrappers, context passing or exporting logs and traces so you do not need to worry about any of it.

For all logging use `log/slog` from the standard library. Note there are not formatting functions available as `slog` is purely structured logging library. Instead of this:

```go
logger.Infof("user with ID %d was authenticated", user.ID)
```

write this:

```go
slog.Info("user was authenticated", "user_id", user.ID)
```

or better this:

```go
slog.InfoContext(ctx, "user was authenticated", "user_id", user.ID)
```

You do not need to know much about `slog` other than it provides four logging functions (levels) and you can optionally provide zero or more key-value pairs. It is possible to create groups (sub-fields) with `WithGroup` or `slog.Group` attriute, instatiate logger using `With` functions and furhter optimize attribute creation but this is only useful when dealing with many logging statements (e.g. in loops). For a normal day to day logging in osbuild project, you do not need any of that.

Please do all logging with lowercase letters, since the standard practice is to use lowercase for Go errors as well, it works nicely hand in hand. This is useful for error wrapping and you do not need to think about case anymore.

Always pass `context.Context` down the stack, this allows for automatic log correlation. You do not need to do anything and all log statements will have `trace_id` that can be used for searching in Kibana or Splunk. Example:

```go
// an Echo handler
ctx := c.Request().Context()

// log like this
slog.InfoContext(ctx, "some message")

// pass the context to all functions
doSomething(ctx)
```

Of course, small functions that will unlikely do any logging or call an external service do not need a context.

There is additional package named `strc` which provides simple tracing, use this when you want to be able to tell how much time was spent in specific block of code (e.g. a function). These blocks are called "spans" and are nested, the tracing information carries over to external systems as well, this is all automatic as long as you call osbuild services you do not need to do anything:

```go
span, ctx := strc.StartContext(ctx, "calculating something big")
defer span.End()
```

Note the `ctx` variable is overwritten by the `StartContext` function, it is best to overwrite it not creating a copy that can be confusing:

```go
span, ctx2 := strc.StartContext(ctx, "calculating something big")
defer span.End()

doSomething(ctx) // wrong, should be ctx2
```

These are all the basics.

Further information about individual packages:

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

## Full example

For a full example, see [example_web](internal/example_web/main.go). To see it in action:

```
go run github.com/osbuild/logging/internal/example_web/
```

See [logrus](pkg/logrus) package for more info.

## AUTHORS and LICENSE

License: MIT

Some code in `splunk` was borrowed from https://github.com/osbuild/osbuild-composer and some code in `strc` is from https://github.com/samber/slog-http

## TODO

* Tracing must be off by default.
* Auto trace_id adding via callback function (to allow UUID adding).
* Rebuild examples in READMEs.
* CLI tool for analyzing data from Splunk.
