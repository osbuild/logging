package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"
)

// TraceBuilder is a builder for traces.
type TraceBuilder struct {
	// RootKey is the root key of the trace concatenated by dots. For example, to
	// process only spans under "event" map and "span" child use "event.span".
	RootKey string

	// FilterTraceID is the trace ID to filter. If empty, all traces are processed.
	FilterTraceID string

	// StripSourcePrefix is the prefix to strip from the source field.
	StripSourcePrefix string

	Window *TraceWindow
}

// NewProfileBuilder creates a new profile builder. If rootKey is empty, "span"
// is used as the root key. If filterTraceID is empty, all traces are processed. If
// stripSourcePrefix is empty, no prefix is stripped.
func NewProfileBuilder(rootKey, filterTraceID, stripSourcePrefix string) *TraceBuilder {
	if rootKey == "" {
		rootKey = "span"
	}

	return &TraceBuilder{
		RootKey:           rootKey,
		FilterTraceID:     filterTraceID,
		StripSourcePrefix: stripSourcePrefix,
		Window:            &TraceWindow{spans: make(map[string][]*Span)},
	}
}

// Process reads a trace from the reader and builds the trace. It returns an error
// if the trace is invalid. The trace is built by appending spans to the window
// and popping them when the root span is found.
func (b *TraceBuilder) Process(r io.Reader, w io.Writer, format string) error {
	e := make(map[string]any)
	scanner := bufio.NewScanner(r)

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "panic while parsing: %v\n", e)
			panic(r)
		}
	}()

	for scanner.Scan() {
		e = make(map[string]any)

		slog.Debug(scanner.Text())
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			return fmt.Errorf("error unmarshalling event: %w", err)
		}

		for _, k := range strings.Split(b.RootKey, ".") {
			if k == "" {
				break
			}
			if v, ok := e[k]; ok {
				e, ok = v.(map[string]any)
				if !ok {
					return fmt.Errorf("key %s value is not a map: %s", k, scanner.Text())
				}
			} else {
				return fmt.Errorf("key %s does not exist: %s", k, scanner.Text())
			}
		}

		span := Span{
			ID:       e["id"].(string),
			ParentID: e["parent"].(string),
			Name:     e["name"].(string),
			TraceID:  e["trace"].(string),
			Source:   e["source"].(string),
		}
		if b.StripSourcePrefix != "" {
			span.Source = strings.TrimPrefix(span.Source, b.StripSourcePrefix)
		}
		if _, ok := e["dur"]; !ok {
			continue
		}
		dur, ok := e["dur"].(float64)
		if !ok {
			return fmt.Errorf("duration is not an int: %s", scanner.Text())
		}
		span.Duration = time.Duration(dur) * time.Nanosecond
		if b.FilterTraceID != "" && span.TraceID != b.FilterTraceID {
			continue
		}

		slog.Info("appending span", "span", span)
		rdy := b.Window.Append(&span)
		if rdy {
			spans := b.Window.Pop(span.TraceID)
			if len(spans) < 1 {
				return fmt.Errorf("empty trace: %s", span.TraceID)
			}

			slices.Reverse(spans)
			root := spans[0]
			root.Children = make([]*Span, 0)

			for _, s := range spans[1:] {
				s.Children = make([]*Span, 0)
				root.Each(func(p *Span) {
					if p.ID == s.ParentID {
						p.Children = append(p.Children, s)
					}
				})
			}

			root.EachWithParent(func(s *Span, p *Span) {
				s.Parent = p
			}, nil)

			if format == "flamegraph" {
				root.Each(func(s *Span) {
					fmt.Fprintf(w, "%s.%s %d\n", s.TraceID, s.JoinNames(";"), s.Duration.Microseconds())
				})
			} else {
				root.EachWithLevel(func(s *Span, lvl int) {
					fmt.Fprintf(w, "%s%s.%s: %s (%s)\n", strings.Repeat(" ", lvl*2), s.TraceID, s.JoinNames("."), s.Duration.String(), s.Source)
				}, 0)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}
