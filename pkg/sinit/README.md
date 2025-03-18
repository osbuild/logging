## sinit

Common log config and initialization

For a full example run:

```
export SPLUNK_URL=xxx
export SPLUNK_TOKEN=xxx
export AWS_REGION=xxx
export AWS_SECRET=xxx
export AWS_KEY=xxx
export SENTRY_DSN=xxx
go run github.com/osbuild/logging/internal/example_sinit/
```

### pgx logging

This package provides a function which returns a wrapper that can be used for pgx SQL driver logging:

```go
pgxConfig.Logger = sinit.PgxLogger()
```
