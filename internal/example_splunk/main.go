package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/osbuild/logging/pkg/splunk"
)

func main() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		fmt.Println(buf.String())
	}))
	defer srv.Close()

	url, ok := os.LookupEnv("SPLUNK_URL")
	if !ok {
		url = srv.URL
	}
	token, ok := os.LookupEnv("SPLUNK_TOKEN")

	h := splunk.NewSplunkHandler(context.Background(), slog.LevelDebug, url, token, "source", "hostname")

	log := slog.New(h)
	log.Debug("message", "k1", "v1")

	defer func() {
		// block until all logs are sent but not more than 2 seconds
		h.Close()

		s := h.Statistics()
		fmt.Printf("sent %d events in %d batches\n", s.EventCount, s.BatchCount)
	}()
}
