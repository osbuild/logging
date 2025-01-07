## Journal log/slog handler

A log/slog handler for systemd-journal.

### How to use

```go
h := journal.NewHandler(context.Background(), slog.LevelDebug)
log := slog.New(h)
log.Debug("message")
```

To see it in action:

```
go run github.com/osbuild/logging/internal/example_journal/
```

Then search for the logs:

```
journalctl -o json-pretty -e
```
