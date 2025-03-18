package sinit

import (
	pgxslog "github.com/mcosta74/pgx-slog"

	"github.com/jackc/pgx/v5/tracelog"
)

// PgxLogger returns a wrapper for PGX logger which logs into the initialized slog.
// Function InitializeLogging must be called first. Usage:
//
// pgxConfig.Logger = sinit.PgxLogger()
func PgxLogger() tracelog.Logger {
	if logger == nil {
		panic("InitializeLogging must be called first")
	}

	return pgxslog.NewLogger(logger)
}
