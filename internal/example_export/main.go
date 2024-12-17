package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/osbuild/logging/pkg/strc"
)

func exportFunc(ctx context.Context, attrs []slog.Attr) {
	for _, attr := range attrs {
		println("exporting trace data", attr.Key, attr.Value.String())
	}
}

func main() {
	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	exportHandler := strc.NewExportHandler(exportFunc)
	multiHandler := strc.NewMultiHandler(textHandler, exportHandler)

	logger := slog.New(multiHandler)
	slog.SetDefault(logger)
	strc.SetLogger(logger)

	span, _ := strc.StartContext(context.Background(), "main")
	defer span.End()
}
