package splunk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestSplunkHandler(t *testing.T) {
	batches := 0
	events := 0
	emptyLines := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batches++
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		lines := strings.Split(buf.String(), "\n")
		for _, line := range lines {
			if line == "" {
				emptyLines++
				continue
			}
			events++
			//t.Log(line)
			if !json.Valid([]byte(line)) {
				t.Fatalf("invalid json: %s", line)
			}
		}
	}))
	defer srv.Close()

	tests := []struct {
		name        string
		f           func(*slog.Logger)
		maxChanSize int
		maxBufSize  int
		events      int
		batches     int
	}{
		{
			name: "1 batch 1 event",
			f: func(l *slog.Logger) {
				l.Debug("message", "k1", "v1")
			},
			maxChanSize: DefaultPayloadsChannelSize,
			maxBufSize:  DefaultMaximumSize,
			events:      1,
			batches:     1,
		},
		{
			name: "1 quoted batch 1 event",
			f: func(l *slog.Logger) {
				l.Debug(`msg: "'@#$~^&*}{ęLukáš`, "k1", "v1")
			},
			maxChanSize: DefaultPayloadsChannelSize,
			maxBufSize:  DefaultMaximumSize,
			events:      1,
			batches:     1,
		},
		{
			name: "1 batch 3 events",
			f: func(l *slog.Logger) {
				l.Debug("m1")
				l.Debug("m2")
				l.Debug("m3")
			},
			maxChanSize: DefaultPayloadsChannelSize,
			maxBufSize:  DefaultMaximumSize,
			events:      3,
			batches:     1,
		},
		{
			name: "10 batches",
			f: func(l *slog.Logger) {
				for i := 0; i < 10; i++ {
					l.Debug("m", "i", i)
				}
			},
			maxChanSize: DefaultPayloadsChannelSize,
			maxBufSize:  1,
			events:      10,
			batches:     10,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.name), func(t *testing.T) {
			events = 0
			batches = 0
			emptyLines = 0

			h := NewSplunkHandler(context.Background(), slog.LevelDebug, srv.URL, "", "s", "h")
			h.splunk.payloadsChannelSize = tt.maxChanSize
			h.splunk.maximumSize = tt.maxBufSize
			logger := slog.New(h)
			tt.f(logger)
			h.Close()

			if events != tt.events {
				t.Fatalf("expected %d events, got %d", tt.events, events)
			}

			if batches != tt.batches {
				t.Fatalf("expected %d batches, got %d", tt.batches, batches)
			}

			// one emtpy line per batch
			if emptyLines != tt.batches {
				t.Fatalf("expected %d empty lines, got %d", tt.batches, emptyLines)
			}
		})
	}
}

func TestSplunkHandlerBatching(t *testing.T) {
	var batches atomic.Int64
	var events atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		batches.Add(1)
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		lines := strings.Split(buf.String(), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			events.Add(1)
			if !json.Valid([]byte(line)) {
				t.Fatalf("invalid json: %s", line)
			}
		}
	}))
	defer srv.Close()

	h := NewSplunkHandler(context.Background(), slog.LevelDebug, srv.URL, "", "s", "h")
	h.splunk.maximumSize = 1000
	logger := slog.New(h).WithGroup("g").With("kg1", "kv1")

	for i := 0; i < 4000; i++ {
		logger.Debug("msg", "i", i)
	}
	h.Close()

	t.Logf("events: %d, batches: %d", events.Load(), batches.Load())
	if events.Load() != 4000 {
		t.Fatalf("expected 4000 events, got %d", events.Load())
	}

	// TODO this must be a range it will vary
	if batches.Load() != 1000 {
		t.Fatalf("expected 1000 batches, got %d", batches.Load())
	}
}
