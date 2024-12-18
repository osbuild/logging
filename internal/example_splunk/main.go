package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	"github.com/osbuild/logging/pkg/splunk"
)

func main() {
	// mock Splunk server that prints the received payload
	ctx := context.Background()
	count := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		//fmt.Println(buf.String())
		count++
	}))
	defer srv.Close()

	url, ok := os.LookupEnv("SPLUNK_URL")
	if !ok {
		url = srv.URL
	}
	token, ok := os.LookupEnv("SPLUNK_TOKEN")

	h := splunk.NewSplunkHandler(ctx, slog.LevelDebug, url, token, "source", "hostname")
	defer h.Close()
	log := slog.New(h)

	log.Debug("message", "k1", "v1")
	for i := 0; i < 2000; i++ {
		log.WithGroup("g1").With("k1", "v1").Debug("a very very very very very very long message", "k2", "v2")
	}
	log.WithGroup("g1").With("k1", "v1").Debug("message", "k2", "v2")
	h.Flush()

	time.Sleep(2 * time.Second)
	fmt.Printf("received %d batches\n", count)
}
