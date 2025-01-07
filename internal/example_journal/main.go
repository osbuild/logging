package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/osbuild/logging/pkg/journal"
)

func main() {
	h := journal.NewHandler(context.Background(), slog.LevelDebug)
	log := slog.New(h)
	log.Debug("message", "k1", "v1",
		slog.Bool("b1", true),
		slog.Int("i1", 1),
		slog.Float64("f1", 1.1),
		slog.String("příliš žlutoučký", "kůň úpěl ďábelské ódy"),
		slog.Time("t1", time.Now()),
		slog.Duration("d1", 1*time.Second),
	)
	log.WithGroup("g1").With("w2", "v2").Debug("message", "k1", "v1")
}
