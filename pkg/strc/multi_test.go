package strc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"testing/slogtest"

	"github.com/google/go-cmp/cmp"
	"github.com/osbuild/logging/pkg/strc"
)

func TestMultiSlogtest(t *testing.T) {
	var buf bytes.Buffer
	hj := slog.NewJSONHandler(&buf, nil)
	h := strc.NewMultiHandler(hj)

	results := func() []map[string]any {
		var ms []map[string]any
		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(line, &m); err != nil {
				t.Fatal(err)
			}
			ms = append(ms, m)
		}
		return ms
	}
	err := slogtest.TestHandler(h, results)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMultiCustomSlogtest(t *testing.T) {
	testMultiCallback := func(ctx context.Context, a []slog.Attr) ([]slog.Attr, error) {
		return a, nil
	}

	var buf bytes.Buffer
	hj := slog.NewJSONHandler(&buf, nil)
	h := strc.NewMultiHandlerCustom(nil, testMultiCallback, hj)

	results := func() []map[string]any {
		var ms []map[string]any
		for _, line := range bytes.Split(buf.Bytes(), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			var m map[string]any
			if err := json.Unmarshal(line, &m); err != nil {
				t.Fatal(err)
			}
			ms = append(ms, m)
		}
		return ms
	}
	err := slogtest.TestHandler(h, results)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMultiCustomValuesSlogtest(t *testing.T) {
	type ctxKey string

	testMultiCallback := func(ctx context.Context, a []slog.Attr) ([]slog.Attr, error) {
		if v := ctx.Value(ctxKey("key")); v != nil {
			a = append(a, slog.Any("ctx", v))
		}

		return a, nil
	}

	tests := []struct {
		block func(logger *slog.Logger)
		want  string
	}{
		{
			block: func(logger *slog.Logger) {
				logger.Debug("test")
			},
			want: `{"msg":"test","a":"b"}`,
		},
		{
			block: func(logger *slog.Logger) {
				logger.InfoContext(context.Background(), "test")
			},
			want: `{"msg":"test","a":"b"}`,
		},
		{
			block: func(logger *slog.Logger) {
				ctx := context.WithValue(context.Background(), ctxKey("key"), "value")
				logger.ErrorContext(ctx, "test")
			},
			want: `{"msg":"test","a":"b","ctx":"value"}`,
		},
	}

	for _, tc := range tests {
		var buf bytes.Buffer
		hj := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
			ReplaceAttr: func(group []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey || a.Key == slog.LevelKey || a.Key == strc.BuildIDFieldKey {
					return slog.Attr{}
				}

				return a
			},
		})
		h := strc.NewMultiHandlerCustom([]slog.Attr{slog.String("a", "b")}, testMultiCallback, hj)
		tc.block(slog.New(h))
		if diff := cmp.Diff(strings.TrimSpace(tc.want), strings.TrimSpace(buf.String())); diff != "" {
			t.Errorf("unexpected diff: %s", diff)
		}
	}
}
