package strc

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
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
		// make sure the source for trace/span IDs is deterministic
		src = rand.NewSource(0)

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
	span, ctx := StartContext(ctx, "subProcess")
	defer span.End()

	span.Event("e")
}

func process(ctx context.Context) {
	span, ctx := StartContext(ctx, "process", "k1", "v1")
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
		f           func(*slog.Logger)
		skipSource  bool
		includeTime bool
		want        []slog.Attr
	}{
		// tracing tests without source
		{
			f: func(l *slog.Logger) {
				process(context.Background())
			},
			skipSource: true,
			want: []slog.Attr{
				slog.Group("span",
					"name", "process", "id", "IvQORsVBkYcT", "trace", "bqzcRlJahlbbBZHS", "source", ""),
				slog.String("k1", "v1"),
				slog.Group("span",
					"name", "subProcess", "id", "IvQORsVBkYcT.VIPEcESnuyuu", "trace", "bqzcRlJahlbbBZHS", "source", ""),
				slog.Group("span",
					"name", "subProcess", "id", "IvQORsVBkYcT.VIPEcESnuyuu", "trace", "bqzcRlJahlbbBZHS", "event", "e", "at", "0s", "source", ""),
				slog.Group("span",
					"name", "subProcess", "id", "IvQORsVBkYcT.VIPEcESnuyuu", "trace", "bqzcRlJahlbbBZHS", "dur", "0s", "source", ""),
				slog.Group("span",
					"name", "process", "id", "IvQORsVBkYcT", "trace", "bqzcRlJahlbbBZHS", "dur", "0s", "source", ""),
			},
		},
		// tracing tests with source
		{
			f: func(l *slog.Logger) {
				process(context.Background())
			},
			want: []slog.Attr{
				slog.Group("span",
					"name", "process", "id", "IvQORsVBkYcT", "trace", "bqzcRlJahlbbBZHS", "source", "exporter_test.go:0"),
				slog.String("k1", "v1"),
				slog.Group("span",
					"name", "subProcess", "id", "IvQORsVBkYcT.VIPEcESnuyuu", "trace", "bqzcRlJahlbbBZHS", "source", "exporter_test.go:0"),
				slog.Group("span",
					"name", "subProcess", "id", "IvQORsVBkYcT.VIPEcESnuyuu", "trace", "bqzcRlJahlbbBZHS", "event", "e", "at", "0s", "source", "exporter_test.go:0"),
				slog.Group("span",
					"name", "subProcess", "id", "IvQORsVBkYcT.VIPEcESnuyuu", "trace", "bqzcRlJahlbbBZHS", "dur", "0s", "source", "exporter_test.go:0"),
				slog.Group("span",
					"name", "process", "id", "IvQORsVBkYcT", "trace", "bqzcRlJahlbbBZHS", "dur", "0s", "source", "exporter_test.go:0"),
			},
		},
		{
			f: func(l *slog.Logger) {
				process(context.Background())
			},
			includeTime: true,
			want: []slog.Attr{
				slog.Group("span",
					"name", "process", "id", "IvQORsVBkYcT", "trace", "bqzcRlJahlbbBZHS", "source", "exporter_test.go:0"),
				slog.String("k1", "v1"),
				slog.Group("span", slog.Time("time", tm)),
				slog.Group("span",
					"name", "subProcess", "id", "IvQORsVBkYcT.VIPEcESnuyuu", "trace", "bqzcRlJahlbbBZHS", "source", "exporter_test.go:0"),
				slog.Group("span", slog.Time("time", tm)),
				slog.Group("span",
					"name", "subProcess", "id", "IvQORsVBkYcT.VIPEcESnuyuu", "trace", "bqzcRlJahlbbBZHS", "event", "e", "at", "0s", "source", "exporter_test.go:0"),
				slog.Group("span", slog.Time("time", tm)),
				slog.Group("span",
					"name", "subProcess", "id", "IvQORsVBkYcT.VIPEcESnuyuu", "trace", "bqzcRlJahlbbBZHS", "dur", "0s", "source", "exporter_test.go:0"),
				slog.Group("span", slog.Time("time", tm)),
				slog.Group("span",
					"name", "process", "id", "IvQORsVBkYcT", "trace", "bqzcRlJahlbbBZHS", "dur", "0s", "source", "exporter_test.go:0"),
				slog.Group("span", slog.Time("time", tm)),
			},
		},
	}

	sourceRegexp := regexp.MustCompile(`exporter_test.go:\d+`)
	for _, tt := range tests {
		// make sure the source for trace/span IDs is deterministic
		src = rand.NewSource(0)

		want := tt.want
		t.Run(fmt.Sprintf("%v", want), func(t *testing.T) {
			got = make([]slog.Attr, 0, 2)
			SkipSource = tt.skipSource
			if tt.includeTime {
				handler = NewExportHandler(callback, IncludeTime())
			} else {
				handler = NewExportHandler(callback)
			}
			logger := slog.New(handler)
			SetLogger(logger)
			tt.f(logger)

			if len(got) != len(want) {
				logAttrs(t, want, got)
				t.Fatalf("len(got) = %d, len(want) = %d", len(got), len(want))
			}

			for i, r := range got {
				w := want[i]

				if r.Value.Kind() == slog.KindGroup {
					group := r.Value.Group()
					for i, g := range group {
						// reset duration to 0s in all groups up until level 1
						if g.Value.Kind() == slog.KindDuration {
							group[i].Value = slog.DurationValue(0)
						}

						// round time to 0 in all groups up until level 1
						if g.Value.Kind() == slog.KindTime {
							group[i].Value = slog.TimeValue(tm)
						}

						// validate source against regexp and reset to exporter_test.go:0
						if g.Key == slog.SourceKey && !tt.skipSource {
							if !sourceRegexp.MatchString(g.Value.String()) {
								t.Errorf("got[%d].Value = %q does not match exporter_text source filename", i, g.Value.String())
							}
							group[i].Value = slog.StringValue("exporter_test.go:0")
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
