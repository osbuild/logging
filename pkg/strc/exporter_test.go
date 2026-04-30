package strc

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestLogExporter(t *testing.T) {
	var got []slog.Attr
	e := NewExportHandler(func(_ context.Context, attrs []slog.Attr) {
		got = append(got, attrs...)
	})

	tests := []struct {
		f    func(*slog.Logger)
		want []slog.Attr
	}{
		// plain slog tests
		{
			f: func(l *slog.Logger) {
				l.Info("msg", "k1", "v1")
			},
			want: []slog.Attr{
				slog.String("k1", "v1"),
			},
		},
		{
			f: func(l *slog.Logger) {
				l.Info("msg", "k1", "v1", "k2", "v2")
			},
			want: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
		},
		{
			f: func(l *slog.Logger) {
				l.With("k1", "v1").Info("msg")
			},
			want: []slog.Attr{
				slog.String("k1", "v1"),
			},
		},
		{
			f: func(l *slog.Logger) {
				l.With("k1", "v1").Info("msg", "k2", "v2")
			},
			want: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
		},
		{
			f: func(l *slog.Logger) {
				l.WithGroup("g1").Info("msg", "k1", "v1")
			},
			want: []slog.Attr{
				slog.Group("g1", "k1", "v1"),
			},
		},
		{
			f: func(l *slog.Logger) {
				l.WithGroup("g1").Info("msg", "k1", "v1")
				l.WithGroup("g1").Info("msg", "k1", "v1")
			},
			want: []slog.Attr{
				slog.Group("g1", "k1", "v1"),
				slog.Group("g1", "k1", "v1"),
			},
		},
		{
			f: func(l *slog.Logger) {
				l.WithGroup("g1").Info("msg", "k1", "v1", "k2", "v2")
			},
			want: []slog.Attr{
				slog.Group("g1", "k1", "v1"),
				slog.Group("g1", "k2", "v2"),
			},
		},
		{
			f: func(l *slog.Logger) {
				l.WithGroup("g1").WithGroup("g2").Info("msg", "k1", "v1", "k2", "v2")
			},
			want: []slog.Attr{
				slog.Group("g1", slog.Group("g2", "k1", "v1")),
				slog.Group("g1", slog.Group("g2", "k2", "v2")),
			},
		},
		{
			f: func(l *slog.Logger) {
				l.WithGroup("g1").With("w1", "v1").Info("msg", "k1", "v1", "k2", "v2")
				l.Error("msg") // should be ignored
				l.InfoContext(context.Background(), "msg", "k1", "v1")
			},
			want: []slog.Attr{
				slog.Group("g1", "w1", "v1"),
				slog.Group("g1", "k1", "v1"),
				slog.Group("g1", "k2", "v2"),
				slog.String("k1", "v1"),
			},
		},
	}

	for _, tt := range tests {
		want := tt.want
		t.Run(fmt.Sprintf("%v", want), func(t *testing.T) {
			got = make([]slog.Attr, 0, 2)
			logger := slog.New(e)
			SetLogger(logger)
			tt.f(logger)
			if len(got) != len(want) {
				logAttrs(t, want, got)
				t.Fatalf("len(got) = %d, len(want) = %d", len(got), len(want))
			}

			for i, r := range got {
				w := want[i]

				if r.Key != w.Key {
					t.Errorf("got[%d].Key = %q, want[%d].Key = %q", i, r.Key, i, w.Key)
				}
				if r.Value.String() != w.Value.String() {
					t.Errorf("got[%d].Value = %q, want[%d].Value = %q", i, r.Value.String(), i, w.Value.String())
				}
			}
		})
	}
}

func subProcess(ctx context.Context) {
	span, _ := Start(ctx, "subProcess")
	defer span.End()

	span.Event("e")
}

func process(ctx context.Context) {
	span, ctx := Start(ctx, "process", "k1", "v1")
	defer span.End()

	subProcess(ctx)
}

func TestTraceExporterPC(t *testing.T) {
	var got []slog.Attr
	var handler slog.Handler
	callback := func(_ context.Context, attrs []slog.Attr) {
		got = append(got, attrs...)
	}
	tm := time.Now()

	tests := []struct {
		name        string
		f           func(*slog.Logger)
		skipSource  bool
		includeTime bool
		want        []slog.Attr
	}{
		{
			name: "tracing tests without source",
			f: func(l *slog.Logger) {
				process(context.Background())
			},
			skipSource: true,
			want: []slog.Attr{
				slog.Group("span", "k1", "v1"),
				slog.Group("span", "name", "process"),
				slog.Group("span", "id", "IvQORsV"),
				slog.Group("span", "parent", "0000000"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "name", "subProcess"),
				slog.Group("span", "id", "kYcTpgn"),
				slog.Group("span", "parent", "IvQORsV"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "name", "subProcess"),
				slog.Group("span", "id", "kYcTpgn"),
				slog.Group("span", "parent", "IvQORsV"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "event", "e"),
				slog.Group("span", "at", "0s"),
				slog.Group("span", "name", "subProcess"),
				slog.Group("span", "id", "kYcTpgn"),
				slog.Group("span", "parent", "IvQORsV"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "dur", "0s"),
				slog.Group("span", "k1", "v1"),
				slog.Group("span", "name", "process"),
				slog.Group("span", "id", "IvQORsV"),
				slog.Group("span", "parent", "0000000"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "dur", "0s"),
			},
		},
		{
			name: "tracing tests with source",
			f: func(l *slog.Logger) {
				process(context.Background())
			},
			want: []slog.Attr{
				slog.Group("span", "k1", "v1"),
				slog.Group("span", "name", "process"),
				slog.Group("span", "id", "IvQORsV"),
				slog.Group("span", "parent", "0000000"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "source", "exporter_test.go:0"),
				slog.Group("span", "name", "subProcess"),
				slog.Group("span", "id", "kYcTpgn"),
				slog.Group("span", "parent", "IvQORsV"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "source", "exporter_test.go:0"),
				slog.Group("span", "name", "subProcess"),
				slog.Group("span", "id", "kYcTpgn"),
				slog.Group("span", "parent", "IvQORsV"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "event", "e"),
				slog.Group("span", "at", "0s"),
				slog.Group("span", "source", "exporter_test.go:0"),
				slog.Group("span", "name", "subProcess"),
				slog.Group("span", "id", "kYcTpgn"),
				slog.Group("span", "parent", "IvQORsV"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "dur", "0s"),
				slog.Group("span", "source", "exporter_test.go:0"),
				slog.Group("span", "k1", "v1"),
				slog.Group("span", "name", "process"),
				slog.Group("span", "id", "IvQORsV"),
				slog.Group("span", "parent", "0000000"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "dur", "0s"),
				slog.Group("span", "source", "exporter_test.go:0"),
			},
		},
		{
			name: "tracing tests with time",
			f: func(l *slog.Logger) {
				process(context.Background())
			},
			includeTime: true,
			want: []slog.Attr{
				slog.Group("span", "k1", "v1"),
				slog.Group("span", "name", "process"),
				slog.Group("span", "id", "IvQORsV"),
				slog.Group("span", "parent", "0000000"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "source", "exporter_test.go:0"),
				slog.Group("span", "time", tm),
				slog.Group("span", "name", "subProcess"),
				slog.Group("span", "id", "kYcTpgn"),
				slog.Group("span", "parent", "IvQORsV"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "source", "exporter_test.go:0"),
				slog.Group("span", "time", tm),
				slog.Group("span", "name", "subProcess"),
				slog.Group("span", "id", "kYcTpgn"),
				slog.Group("span", "parent", "IvQORsV"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "event", "e"),
				slog.Group("span", "at", "0s"),
				slog.Group("span", "source", "exporter_test.go:0"),
				slog.Group("span", "time", tm),
				slog.Group("span", "name", "subProcess"),
				slog.Group("span", "id", "kYcTpgn"),
				slog.Group("span", "parent", "IvQORsV"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "dur", "0s"),
				slog.Group("span", "source", "exporter_test.go:0"),
				slog.Group("span", "time", tm),
				slog.Group("span", "k1", "v1"),
				slog.Group("span", "name", "process"),
				slog.Group("span", "id", "IvQORsV"),
				slog.Group("span", "parent", "0000000"),
				slog.Group("span", "trace", "bqzcRlJahlbbBZH"),
				slog.Group("span", "dur", "0s"),
				slog.Group("span", "source", "exporter_test.go:0"),
				slog.Group("span", "time", tm),
			},
		},
	}

	sourceRegexp := regexp.MustCompile(`exporter_test.go:\d+`)

	// Map to store ID mappings: actual ID -> expected ID for comparison
	// This allows us to verify ID structure without hardcoding specific random values
	type idMap struct {
		trace  map[string]string
		span   map[string]string
		parent map[string]string
	}

	for _, tt := range tests {
		want := tt.want
		t.Run(fmt.Sprintf("%v", tt.name), func(t *testing.T) {
			got = make([]slog.Attr, 0, 2)
			SkipSource = tt.skipSource
			if tt.includeTime {
				handler = NewExportHandler(callback, IncludeTime())
			} else {
				handler = NewExportHandler(callback)
			}
			logger := slog.New(handler)
			SetLogger(logger)

			// Create ID mappings for this test run
			ids := idMap{
				trace:  make(map[string]string),
				span:   make(map[string]string),
				parent: make(map[string]string),
			}

			tt.f(logger)

			if len(got) != len(want) {
				logAttrs(t, want, got)
				t.Fatalf("len(got) = %d, len(want) = %d", len(got), len(want))
			}

			for i, r := range got {
				w := want[i]

				if r.Value.Kind() == slog.KindGroup {
					group := r.Value.Group()
					wantGroup := w.Value.Group()
					for j, g := range group {
						// reset duration to 0s in all groups up until level 1
						if g.Value.Kind() == slog.KindDuration {
							group[j].Value = slog.DurationValue(0)
						}

						// round time to 0 in all groups up until level 1
						if g.Value.Kind() == slog.KindTime {
							group[j].Value = slog.TimeValue(tm)
						}

						// validate source against regexp and reset to exporter_test.go:0
						if g.Key == slog.SourceKey && !tt.skipSource {
							if !sourceRegexp.MatchString(g.Value.String()) {
								t.Errorf("got[%d].Value = %q does not match exporter_text source filename", j, g.Value.String())
							}
							group[j].Value = slog.StringValue("exporter_test.go:0")
						}

						// Normalize trace/span/parent IDs by mapping actual values to expected values
						if g.Key == "trace" || g.Key == "id" || g.Key == "parent" {
							actualID := g.Value.String()
							var expectedID string
							for _, wg := range wantGroup {
								if wg.Key == g.Key {
									expectedID = wg.Value.String()
									break
								}
							}

							// Create or verify mapping
							var idMapping map[string]string
							switch g.Key {
							case "trace":
								idMapping = ids.trace
							case "id":
								idMapping = ids.span
							case "parent":
								idMapping = ids.parent
							}

							if mapped, exists := idMapping[actualID]; exists {
								// Verify consistency
								if mapped != expectedID {
									t.Errorf("inconsistent ID mapping for %s: %s -> %s, expected %s", g.Key, actualID, mapped, expectedID)
								}
							} else {
								// Store new mapping
								idMapping[actualID] = expectedID
							}

							// Replace with expected ID for comparison
							group[j].Value = slog.StringValue(expectedID)
						}
					}
				}

				if got[i].Key == slog.SourceKey {
					if !sourceRegexp.MatchString(got[i].Value.String()) {
						t.Errorf("got[%d].Value = %q, want[%d].Value = %q", i, got[i].Value.String(), i, w.Value.String())
					}
					continue
				}
				if got[i].Key != w.Key {
					t.Errorf("got[%d].Key = %q, want[%d].Key = %q", i, got[i].Key, i, w.Key)
				}
				if got[i].Value.String() != w.Value.String() {
					t.Errorf("got[%d].Value = %q, want[%d].Value = %q", i, got[i].Value.String(), i, w.Value.String())
				}
			}
		})
	}
}

func logAttrs(t *testing.T, want, got []slog.Attr) {
	sb := strings.Builder{}
	sb.WriteString("want:\n")
	for _, attr := range want {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", attr.Key, attr.Value.String()))
	}
	sb.WriteString("got:\n")
	for _, attr := range got {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", attr.Key, attr.Value.String()))
	}
	t.Log(sb.String())
}
