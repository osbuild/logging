package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
)

var (
	verbose           bool
	debug             bool
	rootKey           string
	filterTraceID     string
	stripSourcePrefix string
	outputFormat      string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose logging")
	flag.StringVar(&rootKey, "root-key", "span", "root key of the trace")
	flag.StringVar(&filterTraceID, "filter-trace-id", "", "trace ID to filter")
	flag.StringVar(&stripSourcePrefix, "strip-source-prefix", "", "prefix to strip from the source field")
	flag.StringVar(&outputFormat, "output-format", "text", "output format (text/flamegraph)")
}

func main() {
	flag.Parse()

	if verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	} else if debug {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	} else {
		slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError})))
	}

	b := NewProfileBuilder(rootKey, filterTraceID, stripSourcePrefix)
	if err := b.Process(os.Stdin, os.Stdout, outputFormat); err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}
}
