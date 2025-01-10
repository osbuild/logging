// Simple tracing graph processor
//
// Package stgraph provides a simple tracing graph processor
// that can be used to generate flamegraphs from structured logs.
// It reads structured logs from stdin and writes the flamegraph
// to stdout. Optionally, it can filter traces by trace ID and
// strip a prefix from the source field.
//
// The processor reads structured logs in JSON format generated by
// strc package. The logs are expected to have the following fields:
// - trace: the trace ID
// - id: the span ID
// - parent: the parent span ID
// - name: the span name
// - source: the source file and line number
// - dur: the duration of the span in nanoseconds
package main
