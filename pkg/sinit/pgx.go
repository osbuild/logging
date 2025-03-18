package sinit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/osbuild/logging/pkg/strc"
)

// PgxTracer returns a wrapper for PGX logger which logs into the initialized slog.
// Function InitializeLogging must be called first. Usage:
//
// pgxConfig.Tracer = sinit.PgxTracer(slog.Default())
func PgxTracer(logger *slog.Logger) pgx.QueryTracer {
	if logger == nil {
		panic("called with nil logger")
	}

	return &dbTracer{logger: logger}
}

// Used for pgx logging with context information
type dbTracer struct {
	logger *slog.Logger
}

var _ pgx.QueryTracer = &dbTracer{}

func (dt *dbTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if !dt.logger.Enabled(ctx, slog.LevelDebug) {
		return ctx
	}

	sql, args := formatSqlLog(data)
	var pid uint32
	if conn != nil {
		pid = conn.PgConn().PID()
	}

	tracer := strc.NewTracer(dt.logger)
	span, ctx := tracer.Start(ctx, "query",
		"conn_id", pid,
		"sql", sql,
		"args", args,
	)
	return strc.WithSpan(ctx, span)
}

func (dt *dbTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if !dt.logger.Enabled(ctx, slog.LevelDebug) {
		return
	}
	span := strc.SpanFromContext(ctx)
	if span == nil {
		return
	}

	if data.Err != nil {
		span.End("err", data.Err, "ct", data.CommandTag.String())
	} else {
		span.End("ct", data.CommandTag.String())
	}
}

const maxSqlLogLength = 100

func formatSqlLog(data pgx.TraceQueryStartData) (string, string) {
	d := make([]any, len(data.Args))
	copy(d, data.Args)
	for i, v := range d {
		if b, ok := v.([]byte); ok {
			d[i] = ellipsis(string(b), maxSqlLogLength)
		} else if j, ok := v.(json.RawMessage); ok {
			d[i] = ellipsis(string(j), maxSqlLogLength)
		}
	}

	return data.SQL, fmt.Sprintf("%v", d)
}

func ellipsis(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return string(runes[0:maxLen-3]) + "..."
}
