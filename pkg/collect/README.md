## collect

A collector handler - a special slog handler used for testing. Usage:

```
ch := collect.NewTestHandler(slog.LevelDebug, false, false, false)
logger := slog.New(ch)
logger.Debug("message")
logger.Info("message")

fmt.Printf("%v\n", ch.All())
fmt.Printf("%v\n", ch.Last())
```

Logs are collected into a slice of `map[string]any` where they can be picked up for introspection. This package has no other use than in testing.
