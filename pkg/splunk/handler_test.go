package splunk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/slogtest"
)

func TestSplunkHandler(t *testing.T) {
	emptyLines := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		lines := strings.Split(buf.String(), "\n")
		for _, line := range lines {
			if line == "" {
				emptyLines++
				continue
			}
			if !json.Valid([]byte(line)) {
				t.Fatalf("invalid json: %s", line)
			}
		}
	}))
	defer srv.Close()

	tests := []struct {
		name       string
		f          func(*slog.Logger)
		maxBufSize int
		events     int
		batches    int
	}{
		{
			name: "1 batch 1 event",
			f: func(l *slog.Logger) {
				l.Debug("message", "k1", "v1")
			},
			maxBufSize: DefaultMaximumSize,
			events:     1,
			batches:    1,
		},
		{
			name: "1 quoted batch 1 event",
			f: func(l *slog.Logger) {
				l.Debug(`msg: "'@#$~^&*}{ęLukáš`, "k1", "v1")
			},
			maxBufSize: DefaultMaximumSize,
			events:     1,
			batches:    1,
		},
		{
			name: "1 batch 3 events",
			f: func(l *slog.Logger) {
				l.Debug("m1")
				l.Debug("m2")
				l.Debug("m3")
			},
			maxBufSize: DefaultMaximumSize,
			events:     3,
			batches:    1,
		},
		{
			name: "10 batches",
			f: func(l *slog.Logger) {
				for i := 0; i < 10; i++ {
					l.Debug("m", "i", i)
				}
			},
			maxBufSize: 1,
			events:     10,
			batches:    10,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.name), func(t *testing.T) {
			emptyLines = 0

			c := SplunkConfig{
				Level:              slog.LevelDebug,
				URL:                srv.URL,
				Source:             "s",
				Hostname:           "h",
				DefaultMaximumSize: tt.maxBufSize,
			}
			h := NewSplunkHandler(context.Background(), c)
			logger := slog.New(h)
			tt.f(logger)
			h.Close()
			stats := h.Statistics()

			if int(stats.EventCount) != tt.events {
				t.Fatalf("expected %d events, got %d", tt.events, stats.EventCount)
			}

			if int(stats.BatchCount) != tt.batches {
				t.Fatalf("expected %d batches, got %d", tt.batches, stats.BatchCount)
			}

			// one emtpy line per batch
			if emptyLines != tt.batches {
				t.Fatalf("expected %d empty lines, got %d", tt.batches, emptyLines)
			}
		})
	}
}

func TestSplunkHandlerBatching(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)

		lines := strings.Split(buf.String(), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			if !json.Valid([]byte(line)) {
				t.Fatalf("invalid json: %s", line)
			}
		}
	}))
	defer srv.Close()

	c := SplunkConfig{
		Level:              slog.LevelDebug,
		URL:                srv.URL,
		Source:             "s",
		Hostname:           "h",
		DefaultMaximumSize: 1000,
	}
	h := NewSplunkHandler(context.Background(), c)
	logger := slog.New(h).WithGroup("g").With("kg1", "kv1")

	for i := 0; i < 4000; i++ {
		logger.Debug("msg", "i", i)
	}
	h.Close()
	stats := h.Statistics()

	t.Logf("events: %d, batches: %d", stats.EventCount, stats.BatchCount)
	if stats.EventCount != 4000 {
		t.Fatalf("expected 4000 events, got %d", stats.BatchCount)
	}

	// This can depend on call stack (event length)
	if stats.BatchCount == 0 || stats.BatchCount == stats.EventCount {
		t.Fatalf("expected 1000 batches, got %d", stats.BatchCount)
	}
}

func TestSlogtest(t *testing.T) {
	messages := make(chan []byte)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read error: %v", err)
		}
		messages <- buf
	}))
	defer srv.Close()
	defer close(messages)

	c := SplunkConfig{
		Level:              slog.LevelDebug,
		URL:                srv.URL,
		Source:             "s",
		Hostname:           "h",
		DefaultMaximumSize: 1, // force batching
	}

	slogtest.Run(t, func(t *testing.T) slog.Handler {
		return NewSplunkHandler(context.Background(), c)
	}, func(t *testing.T) map[string]any {
		m := make(map[string]any)

		buf := <-messages
		err := json.Unmarshal(buf, &m)
		if err != nil {
			t.Fatalf("invalid json: %v\n%s", err, string(buf))
		}

		// unwrap the event
		return m["event"].(map[string]any)["message"].(map[string]any)
	})
}
