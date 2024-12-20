package splunk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	// DefaultPayloadsChannelSize is the size of the channel that holds payloads, default 1k.
	DefaultPayloadsChannelSize = 4096

	// DefaultEventSize is the initialized capacity of the event buffer, default 1kB
	DefaultEventSize = 1024

	// DefaultMaximumSize is the initialized capacity of the event buffer before it is flushed, default is 1MB.
	DefaultMaximumSize = 1024 * 1024

	// DefaultSendFrequency is the frequency at which payloads are sent at a maximum, default 5s.
	DefaultSendFrequency = 5 * time.Second
)

type splunkLogger struct {
	client   *http.Client
	url      string
	token    string
	source   string
	hostname string

	pool     sync.Pool
	payloads chan []byte
	active   atomic.Bool

	// only modified by tests
	payloadsChannelSize int
	maximumSize         int
	sendFrequency       time.Duration

	stats   Stats
	statsMu sync.Mutex
}

type Stats struct {
	// Total number of events sent to Splunk
	EventCount uint64

	// Total number of requests sent to Splunk
	BatchCount uint64

	// Total number of HTTP retries
	RetryCount uint64

	// Total number of non-200 HTTP responses
	NonHTTP200Count uint64

	// Total number of events enqueued (EventsEnqueued <= EventCount)
	EventsEnqueued uint64

	// Last request duration
	LastRequestDuration time.Duration
}

func newSplunkLogger(ctx context.Context, url, token, source, hostname string) *splunkLogger {
	rcl := retryablehttp.NewClient()

	sl := &splunkLogger{
		client:              rcl.StandardClient(),
		url:                 url,
		token:               token,
		source:              source,
		hostname:            hostname,
		payloadsChannelSize: DefaultPayloadsChannelSize,
		maximumSize:         DefaultMaximumSize,
		sendFrequency:       DefaultSendFrequency,
		pool: sync.Pool{
			New: func() any {
				buf := &bytes.Buffer{}
				buf.Grow(1024)
				return buf
			},
		},
	}

	rcl.RetryWaitMin = 300 * time.Millisecond
	rcl.RetryWaitMax = 3 * time.Second
	rcl.RetryMax = 5
	rcl.Logger = log.New(io.Discard, "", 0)
	rcl.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		sl.statsMu.Lock()
		sl.stats.RetryCount++
		sl.statsMu.Unlock()

		if err != nil && strings.Contains(err.Error(), "nonretryable") {
			return false, nil
		}
		return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
	}

	ticker := time.NewTicker(sl.sendFrequency)
	sl.payloads = make(chan []byte, sl.payloadsChannelSize)

	sl.active.Store(true)
	go sl.flushPayloads(ctx, ticker.C)

	return sl
}

// Statistics returns a copy the current statistics of the logger. It is safe to call
// this method concurrently with other goroutines.
func (sl *splunkLogger) Statistics() Stats {
	sl.statsMu.Lock()
	defer sl.statsMu.Unlock()

	return sl.stats
}

// ErrFullOrClosed is returned when the payloads channel is full or closed via close().
var ErrFullOrClosed = errors.New("cannot create new splunk event: channel full or closed")

// ErrInvalidEvent is returned when the event is not a valid JSON object with a trailing newline.
var ErrInvalidEvent = errors.New("invalid event: must be a JSON object with trailing newline")

// ErrResponseNotOK is returned when the response from Splunk is not 200 OK.
var ErrResponseNotOK = errors.New("unexpected response from Splunk")

func (sl *splunkLogger) flushPayloads(ctx context.Context, ticker <-chan time.Time) {
	defer sl.active.Store(false)

	buf := &bytes.Buffer{}
	buf.Grow(sl.maximumSize + DefaultEventSize)

	sendPayloads := func() {
		err := sl.sendPayloads(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "splunk logger unable to send payloads: %v\n", err)
		}
	}

	for {
		select {
		case <-ctx.Done():
			sendPayloads()
			return
		case event, ok := <-sl.payloads:
			// close call
			if !ok {
				sendPayloads()
				return
			}

			// flush call
			if len(event) == 0 {
				sendPayloads()
				continue
			}

			buf.Write(event)
			sl.statsMu.Lock()
			sl.stats.EventCount++
			sl.statsMu.Unlock()
			if buf.Len() >= sl.maximumSize {
				sendPayloads()
			}
		case <-ticker:
			sendPayloads()
		}
	}
}

func (sl *splunkLogger) sendPayloads(buf *bytes.Buffer) error {
	if buf.Len() == 0 {
		return nil
	}
	defer buf.Truncate(0)

	req, err := http.NewRequest("POST", sl.url, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Splunk "+sl.token)

	start := time.Now()
	res, err := sl.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		sl.statsMu.Lock()
		sl.stats.NonHTTP200Count++
		sl.statsMu.Unlock()
		return ErrResponseNotOK
	}

	dur := time.Since(start)
	sl.statsMu.Lock()
	sl.stats.LastRequestDuration = dur
	sl.stats.BatchCount++
	sl.statsMu.Unlock()
	return nil
}

// flush will cause the logger to flush the current buffer. It does not block, there is no
// guarantee that the buffer will be flushed immediately.
func (sl *splunkLogger) flush() {
	sl.payloads <- []byte("")
}

// close will flush the buffer, close the channel and wait until all payloads are sent,
// not longer than 2 seconds. It is safe to call close multiple times. After close is called
// the client will not accept any new events, all attemtps to send new events will return
// ErrFullOrClosed.
func (sl *splunkLogger) close() {
	close(sl.payloads)

	timeout := time.Now().Add(2 * time.Second)
	for sl.active.Load() {
		time.Sleep(100 * time.Millisecond)

		if time.Now().After(timeout) {
			break
		}
	}
}

// event will create a new event and send it to the payloads channel. It will return
// length of the event or an error if the event is invalid or the channel is full.
// Event must be a valid JSON object with a trailing newline.
func (sl *splunkLogger) event(b []byte) (int, error) {
	buf := sl.pool.Get().(*bytes.Buffer)
	buf.Truncate(0)
	defer sl.pool.Put(buf)

	if len(b) < 2 || b[0] != '{' || b[len(b)-2] != '}' || b[len(b)-1] != '\n' {
		return 0, ErrInvalidEvent
	}

	// remove the trailing newline
	if b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}

	// Construct the event JSON. It uses Go string quoting to escape hostname and stream,
	// it is not ideal but it is fast, simple and we only use alphanumeric characters anyway.
	// Note the message payload itself is delivered through byte buffer, it is not escaped
	// as it is generated by the standard library JSON handler.
	buf.WriteByte('{')
	buf.WriteString(`"time":`)
	buf.WriteString(strconv.FormatInt(time.Now().Unix(), 10))
	buf.WriteByte(',')
	buf.WriteString(`"host":`)
	buf.WriteString(strconv.Quote(sl.hostname))
	buf.WriteString(`,"event":{`)
	buf.WriteString(`"ident":`)
	buf.WriteString(strconv.Quote(sl.source))
	buf.WriteString(`,"host":`)
	buf.WriteString(strconv.Quote(sl.hostname))
	buf.WriteString(`,"message":`)
	buf.Write(b)
	buf.WriteByte('}')
	buf.WriteString("}\n")

	event := make([]byte, buf.Len())
	copy(event, buf.Bytes())

	sl.statsMu.Lock()
	sl.stats.EventsEnqueued++
	sl.statsMu.Unlock()

	select {
	case sl.payloads <- event:
	default:
		return 0, ErrFullOrClosed
	}
	return len(event), nil
}
