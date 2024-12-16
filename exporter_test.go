package strc

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
)

func TestMakeGroup(t *testing.T) {
	tests := []struct {
		group []string
		attrs []slog.Attr
		want  []slog.Attr
	}{
		{},
		{
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
			},
			want: []slog.Attr{
				slog.String("k1", "v1"),
			},
		},
		{
			attrs: []slog.Attr{
				slog.Group("g1", slog.String("k1", "v1")),
			},
			want: []slog.Attr{
				slog.Group("g1", slog.String("k1", "v1")),
			},
		},
		{
			group: []string{"g0"},
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
			},
			want: []slog.Attr{
				slog.Group("g0", slog.String("k1", "v1")),
			},
		},
		{
			group: []string{"g0"},
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
			want: []slog.Attr{
				slog.Group("g0", slog.String("k1", "v1"), slog.String("k2", "v2")),
			},
		},
		{
			group: []string{"g0", "g1"},
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
			},
			want: []slog.Attr{
				slog.Group("g0", slog.Group("g1", slog.String("k1", "v1"))),
			},
		},
		{
			group: []string{"g0", "g1"},
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
			want: []slog.Attr{
				slog.Group("g0", slog.Group("g1", slog.String("k1", "v1"), slog.String("k2", "v2"))),
			},
		},
		{
			group: []string{"g0", "g1"},
			attrs: []slog.Attr{
				slog.Group("g3", slog.String("k1", "v1")),
			},
			want: []slog.Attr{
				slog.Group("g0", slog.Group("g1", slog.Group("g3", slog.String("k1", "v1")))),
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.want), func(t *testing.T) {
			got := makeGroup(tt.group, tt.attrs)
			if len(got) != len(tt.want) {
				t.Fatalf("len(got) = %d, len(tt.want) = %d", len(got), len(tt.want))
			}
			for i := range got {
				if got[i].Key != tt.want[i].Key {
					t.Errorf("got[%d].Key = %q, tt.want[%d].Key = %q", i, got[i].Key, i, tt.want[i].Key)
				}
				if got[i].Value.String() != tt.want[i].Value.String() {
					t.Errorf("got[%d].Value = %q, tt.want[%d].Value = %q", i, got[i].Value.String(), i, tt.want[i].Value.String())
				}
			}
		})
	}
}

func convertToAny(attrs []slog.Attr) []any {
	anyAttrs := make([]any, len(attrs))
	for i, attr := range attrs {
		anyAttrs[i] = attr
	}
	return anyAttrs
}

func TestExporter(t *testing.T) {
	var result []slog.Attr

	e := NewExportHandler(func(attrs []slog.Attr) {
		result = attrs
	})
	g := slog.New(e)

	tests := []struct {
		attrs []slog.Attr
		with  [][]slog.Attr
		group []string
		want  []slog.Attr
	}{
		{
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
			},
			want: []slog.Attr{
				slog.String("k1", "v1"),
			},
		},
		{
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
			want: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
		},
		{
			group: []string{"g0"},
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
			want: []slog.Attr{
				slog.Group("g0", slog.String("k1", "v1"), slog.String("k2", "v2")),
			},
		},
		{
			group: []string{"g0", "g1"},
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
			want: []slog.Attr{
				slog.Group("g0", slog.Group("g1", slog.String("k1", "v1"), slog.String("k2", "v2"))),
			},
		},
		{
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
			},
			with: [][]slog.Attr{
				{
					slog.String("w1", "v1"),
				},
			},
			want: []slog.Attr{
				slog.String("w1", "v1"),
				slog.String("k1", "v1"),
			},
		},
		{
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
			},
			with: [][]slog.Attr{
				{
					slog.String("w1", "v1"),
					slog.String("w2", "v1"),
				},
			},
			want: []slog.Attr{
				slog.String("w1", "v1"),
				slog.String("w2", "v1"),
				slog.String("k1", "v1"),
			},
		},
		{
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
			},
			with: [][]slog.Attr{
				{
					slog.String("w1", "v1"),
				},
				{
					slog.String("w2", "v1"),
				},
			},
			want: []slog.Attr{
				slog.String("w1", "v1"),
				slog.String("w2", "v1"),
				slog.String("k1", "v1"),
			},
		},
		{
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
			},
			with: [][]slog.Attr{
				{
					slog.String("w1", "v1"),
					slog.String("w2", "v1"),
				},
				{
					slog.String("w3", "v1"),
				},
			},
			want: []slog.Attr{
				slog.String("w1", "v1"),
				slog.String("w2", "v1"),
				slog.String("w3", "v1"),
				slog.String("k1", "v1"),
			},
		},
		{
			group: []string{"g0"},
			with: [][]slog.Attr{
				{
					slog.String("w1", "v1"),
				},
			},
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
			want: []slog.Attr{
				slog.Group("g0", slog.String("w1", "v1"), slog.String("k1", "v1"), slog.String("k2", "v2")),
			},
		},
		{
			group: []string{"g0", "g1"},
			attrs: []slog.Attr{
				slog.String("k1", "v1"),
				slog.String("k2", "v2"),
			},
			want: []slog.Attr{
				slog.Group("g0", slog.Group("g1", slog.String("k1", "v1"), slog.String("k2", "v2"))),
			},
		},
	}
			
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.want), func(t *testing.T) {
			result = nil
			lg := g
			for _, group := range tt.group {
				lg = lg.WithGroup(group)
			}
			for _, attrs := range tt.with {
				lg = lg.With(convertToAny(attrs)...)
			}
			lg.Log(context.Background(), slog.LevelInfo, "unused", convertToAny(tt.attrs)...)

			if len(result) != len(tt.want) {
				t.Logf("result: %v", result)
				t.Fatalf("len(result) = %d, len(tt.want) = %d", len(result), len(tt.want))
			}

			for i := range result {
				if result[i].Key != tt.want[i].Key {
					t.Errorf("result[%d].Key = %q, tt.want[%d].Key = %q", i, result[i].Key, i, tt.want[i].Key)
				}
				if result[i].Value.String() != tt.want[i].Value.String() {
					t.Errorf("result[%d].Value = %q, tt.want[%d].Value = %q", i, result[i].Value.String(), i, tt.want[i].Value.String())
				}
			}

		})
	}
}
