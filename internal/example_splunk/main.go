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
	// mock Splunk server that prints the received payload
	ctx := context.Background()
	count := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		fmt.Println(buf.String())
		count++
	}))
	defer srv.Close()

	url, ok := os.LookupEnv("SPLUNK_URL")
	if !ok {
		url = srv.URL
	}
	token, ok := os.LookupEnv("SPLUNK_TOKEN")

	h := splunk.NewSplunkHandler(ctx, slog.LevelDebug, url, token, "source", "hostname")

	log := slog.New(h)
	log.Debug("message", "k1", "v1")

	// Close will block until all logs are sent but not more than 2 seconds
	defer func() {
		h.Close()

		fmt.Printf("received %d batches\n", count)
	}()
}