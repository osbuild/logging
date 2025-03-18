package sinit

import (
	"context"
	"encoding/json"
	"log/slog"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/osbuild/logging/pkg/collect"
)

func TestFormatSqlSimple(t *testing.T) {
	data := pgx.TraceQueryStartData{
		SQL: "select $1 $2",
		Args: []any{
			"hello",
			13,
		},
	}

	sql, args := formatSqlLog(data)

	if !cmp.Equal(sql, "select $1 $2") {
		t.Errorf("Expected data.SQL to be unchanged, got: %v", sql)
	}

	if !cmp.Equal(args, "[hello 13]") {
		t.Errorf("Expected data.Args to be formatted, got: %v", args)
	}
}

func TestFormatSqlByteSlice(t *testing.T) {
	data := pgx.TraceQueryStartData{
		SQL: "select $1 $2",
		Args: []any{
			[]byte("hello"),
			13,
		},
	}

	sql, args := formatSqlLog(data)
	if !cmp.Equal(sql, "select $1 $2") {
		t.Errorf("Expected data.SQL to be unchanged, got: %v", sql)
	}

	if !cmp.Equal(args, "[hello 13]") {
		t.Errorf("Expected data.Args to be formatted, got: %v", args)
	}
}

func TestFormatSqlRawMessage(t *testing.T) {
	data := pgx.TraceQueryStartData{
		SQL: "select $1 $2",
		Args: []any{
			json.RawMessage("hello"),
			13,
		},
	}

	sql, args := formatSqlLog(data)
	if !cmp.Equal(sql, "select $1 $2") {
		t.Errorf("Expected data.SQL to be unchanged, got: %v", sql)
	}

	if !cmp.Equal(args, "[hello 13]") {
		t.Errorf("Expected data.Args to be formatted, got: %v", args)
	}
}

func TestFormatSqlLongSlice(t *testing.T) {
	data := pgx.TraceQueryStartData{
		SQL: "select $1 $2",
		Args: []any{
			make([]byte, 200),
		},
	}
	for i := range data.Args[0].([]byte) {
		data.Args[0].([]byte)[i] = 'a'
	}

	sql, args := formatSqlLog(data)
	if !cmp.Equal(sql, "select $1 $2") {
		t.Errorf("Expected data.SQL to be unchanged, got: %v", sql)
	}

	if !regexp.MustCompile(`\[a{97}\.\.\.\]`).MatchString(args) {
		t.Errorf("Expected data.Args to be formatted, got: %v", args)
	}
}

func TestTraceQuery(t *testing.T) {
	ctx := context.Background()
	ch := collect.NewTestHandler(slog.LevelDebug, false, false, false)

	dt := dbTracer{
		logger: slog.New(ch),
	}

	startData := pgx.TraceQueryStartData{
		SQL: "select $1 $2",
		Args: []any{
			"hello",
			13,
		},
	}

	endData := pgx.TraceQueryEndData{
		CommandTag: pgconn.NewCommandTag("SELECT"),
	}

	tests := []struct {
		f    func()
		want map[string]any
	}{
		{
			f: func() {
				ctx = dt.TraceQueryStart(ctx, nil, startData)
			},
			want: map[string]any{
				"msg": "span query started",
				"span": map[string]any{
					"name":    "query",
					"sql":     "select $1 $2",
					"args":    "[hello 13]",
					"conn_id": uint64(0),
				},
			},
		},
		{
			f: func() {
				dt.TraceQueryEnd(ctx, nil, endData)
			},
			want: map[string]any{
				"msg": "span query finished",
				"span": map[string]any{
					"name":    "query",
					"sql":     "select $1 $2",
					"args":    "[hello 13]",
					"conn_id": uint64(0),
					"ct":      "SELECT",
				},
			},
		},
	}

	formatMsg := regexp.MustCompile("span .+ finished")
	for _, tt := range tests {
		tt.f()
		got := ch.Last()
		span := got["span"].(map[string]any)
		delete(span, "id")
		delete(span, "trace")
		delete(span, "parent")
		delete(span, "source")
		delete(span, "dur")
		if strings.Contains(got["msg"].(string), "finished in") {
			got["msg"] = formatMsg.FindString(got["msg"].(string))
		}

		if !cmp.Equal(got, tt.want) {
			t.Errorf("Got:\n%v, want:\n%v", got, tt.want)
		}
	}
}
