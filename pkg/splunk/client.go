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
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	// DefaultPayloadsChannelSize is the size of the channel that holds payloads, default 1k.
	DefaultPayloadsChannelSize = 4096

	// DefaultMaximumSize is the maximum size of the event buffer before it is flushed, default is 1MB.
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

	payloadsChannelSize int
	maximumSize         int
	sendFrequency       time.Duration

	payloads chan []byte
	active   atomic.Bool
}

func newSplunkLogger(ctx context.Context, url, token, source, hostname string) *splunkLogger {
	rcl := retryablehttp.NewClient()
	rcl.RetryWaitMin = 300 * time.Millisecond
	rcl.RetryWaitMax = 3 * time.Second
	rcl.RetryMax = 5
	rcl.Logger = log.New(io.Discard, "", 0)

	sl := &splunkLogger{
		client:              rcl.StandardClient(),
		url:                 url,
		token:               token,
		source:              source,
		hostname:            hostname,
		payloadsChannelSize: DefaultPayloadsChannelSize,
		maximumSize:         DefaultMaximumSize,
		sendFrequency:       DefaultSendFrequency,
	}

	ticker := time.NewTicker(sl.sendFrequency)
	sl.payloads = make(chan []byte, sl.payloadsChannelSize)

	sl.active.Store(true)
	go sl.flushPayloads(ctx, ticker.C)

	return sl
}

func (sl *splunkLogger) flushPayloads(ctx context.Context, ticker <-chan time.Time) {
	defer sl.active.Store(false)

	buf := &bytes.Buffer{}
	// pre-allocate with a 1kB tail if the soft limit is reached
	buf.Grow(sl.maximumSize + 1024)

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
		case p, ok := <-sl.payloads:
			// close call
			if !ok {
				sendPayloads()
				return
			}

			// flush call
			if len(p) == 0 {
				sendPayloads()
				continue
			}

			buf.Write(p)
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

	res, err := sl.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			return fmt.Errorf("error forwarding to splunk: parsing response failed: %v\n", err)
		}

		return fmt.Errorf("error forwarding to splunk: %s\n", buf.String())
	}
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

	returnAfter := time.Now().Add(2 * time.Second)
	for sl.active.Load() {
		time.Sleep(100 * time.Millisecond)

		if time.Now().After(returnAfter) {
			break
		}
	}
}

// ErrFullOrClosed is returned when the payloads channel is full or closed via close().
var ErrFullOrClosed = errors.New("cannot create new splunk event: channel full or closed")

// ErrInvalidEvent is returned when the event is not a valid JSON object with a trailing newline.
var ErrInvalidEvent = errors.New("invalid event: must be a JSON object with trailing newline")

// an example of a Splunk event for reference and pre-allocation estimate
var exampleRec = `{
	"time": 1234567890,
	"host": "hostname.example.com",
	"event": {
		"indent": "source-ident",
		"host": "hostname.example.com",
		"message": {}
	}
}`

// event will create a new event and send it to the payloads channel. It will return
// length of the event or an error if the event is invalid or the channel is full.
func (sl *splunkLogger) event(b []byte) (int, error) {
	buf := &bytes.Buffer{}
	buf.Grow(len(b) + len(exampleRec)*2)

	if len(b) < 2 || b[0] != '{' || b[len(b)-2] != '}' || b[len(b)-1] != '\n' {
		return 0, ErrInvalidEvent
	}

	// remove the trailing newline
	if b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}

	// Construct the event JSON. It uses Go string quoting to escape hostname and stream,
	// it is not ideal but it is fast, simple and we only use alphanumeric characters anyway.
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

	select {
	case sl.payloads <- event:
	default:
		return 0, ErrFullOrClosed
	}
	return len(event), nil
}
