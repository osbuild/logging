package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lzap/strc"
	slogmulti "github.com/samber/slog-multi"
)

func traceIDMiddleware(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
	if tid := strc.TraceID(ctx); tid != "" {
		record.AddAttrs(slog.String(strc.TraceIDName, tid))
	}

	return next(ctx, record)
}

func exportMiddleware(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
	record.Attrs(func(attr slog.Attr) bool {
		println("attr", attr.Key, attr.Value.Kind(), attr.Value.String())
		if attr.Key == "span" && attr.Value.Kind() == slog.KindGroup {
			fmt.Print("sending span event to somewhere: ")
			for _, a := range attr.Value.Group() {
				fmt.Printf("%s=%v ", a.Key, a.Value)
			}
			fmt.Println()

			return false
		}

		return true
	})

	return next(ctx, record)
}

func printMiddleware(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
	fmt.Printf("%+v\n", record)
	record.Attrs(func(attr slog.Attr) bool {
		println("attr", attr.Key, attr.Value.Kind(), attr.Value.String())
		return true
	})

	return next(ctx, record)
}

func function(ctx context.Context) {
	span, ctx := strc.StartContext(ctx, "function")
	defer span.End()

	slog.With(
		slog.Group("user",
			slog.String("id", "user-123"),
			slog.String("email", "user-123"),
			slog.Time("created_at", time.Now()),
		),
	).With("environment", "dev").
		ErrorContext(ctx, "A message",
			slog.String("foo", "bar"),
			slog.Any("error", fmt.Errorf("an error")),
		)
}

func main() {
	//traceIDMiddleware := slogmulti.NewHandleInlineMiddleware(traceIDMiddleware)
	//exportMiddleware := slogmulti.NewHandleInlineMiddleware(exportMiddleware)
	printMiddleware := slogmulti.NewHandleInlineMiddleware(printMiddleware)
	sink := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{})

	//logger := slog.New(slogmulti.Pipe(traceIDMiddleware).Pipe(exportMiddleware).Handler(sink))
	logger := slog.New(slogmulti.Pipe(printMiddleware).Handler(sink))
	slog.SetDefault(logger)
	strc.SetLogger(logger)

	slog.Default().With(slog.Group("group", slog.String("test", "test"))).Log(context.Background(), slog.LevelInfo, "test")
	os.Exit(0)
	span, ctx := strc.StartContext(context.Background(), "main")
	defer span.End()

	function(ctx)
}
