## splunk

A Splunk event handler for `log/slog`. Features:

* Configurable URL, token, source and hostname.
* Batching support.
* Non-blocking flush call support.
* Blocking close call support with a timeout.
* Memory pool for event byte buffers.
* Utilizes JSON handler from the standard library.
* Built for performance (no JSON encoding or decoding).

### How to use

```go
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
```

Run the example against mock Splunk with the following command:

```
go run github.com/osbuild/logging/internal/example_splunk/
```

Run the example against real Splunk with the following command:

```
export SPLUNK_URL=https://xxx.splunkcloud.com/services/collector/event
export SPLUNK_TOKEN=x7d04bb1-7eae-4bde-9d92-89837206239x
go run github.com/osbuild/logging/internal/example_splunk/
```
