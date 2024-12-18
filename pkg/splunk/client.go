package splunk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

var (
	// PayloadsChannelSize is the size of the channel that holds payloads
	PayloadsChannelSize = 1024

	// MaximumSize is the maximum size of a payload, default is 1MB
	MaximumSize = 1024 * 1024

	// SendFrequency is the frequency at which payloads are sent
	SendFrequency = 5
)

type splunkLogger struct {
	client   *http.Client
	url      string
	token    string
	source   string
	hostname string

	payloads chan []byte
}

type splunkPayload struct {
	Time  int64
	Host  string
	Event splunkEvent
}

type splunkEvent struct {
	Message string
	Ident   string
	Host    string
}

func newSplunkLogger(ctx context.Context, url, token, source, hostname string) *splunkLogger {
	rcl := retryablehttp.NewClient()
	rcl.RetryWaitMin = 300 * time.Millisecond
	rcl.RetryWaitMax = 3 * time.Second
	rcl.RetryMax = 5
	// TODO set rcl.Logger

	sl := &splunkLogger{
		client:   rcl.StandardClient(),
		url:      url,
		token:    token,
		source:   source,
		hostname: hostname,
	}

	ticker := time.NewTicker(time.Second * SendFrequency)
	sl.payloads = make(chan []byte, PayloadsChannelSize)

	go sl.flushPayloads(ctx, ticker.C)

	return sl
}

func (sl *splunkLogger) flushPayloads(context context.Context, ticker <-chan time.Time) {
	buf := &bytes.Buffer{}
	// pre-allocate with a 1kB tail if the soft limit is reached
	buf.Grow(MaximumSize + 1024)

	sendPayloads := func() {
		err := sl.sendPayloads(buf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "splunk logger unable to send payloads: %v\n", err)
		}
	}

	for {
		select {
		case <-context.Done():
			sendPayloads()
			return
		case p, ok := <-sl.payloads:
			if !ok {
				sendPayloads()
				return
			}

			if len(p) == 0 {
				sendPayloads()
				continue
			}

			buf.Write(p)
			if buf.Len() >= MaximumSize {
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

func (sl *splunkLogger) flush() {
	sl.payloads <- []byte("")
}

func (sl *splunkLogger) close() {
	close(sl.payloads)
}

// ErrFullOrClosed is returned when the payloads channel is full or closed via close().
var ErrFullOrClosed = errors.New("cannot create new splunk event: channel full or closed")

var recordTemplateStr = `
{"time": {{.Time}}, "host": "{{.Host}}", "event": {
"message": {{.Event.Message}},
"ident": "{{.Event.Ident}}",
"host": "{{.Event.Host}}"}}`

var recordTemplate = template.Must(template.New("record").Parse(recordTemplateStr))

func (sl *splunkLogger) event(b []byte) error {
	buf := &bytes.Buffer{}
	buf.Grow(len(b) + len(recordTemplateStr))

	// remove possible newline
	if b[len(b)-1] == '\n' {
		b = b[:len(b)-1]
	}

	if len(b) < 2 || b[0] != '{' || b[len(b)-1] != '}' {
		// Input is string (only used in tests, no JSON quoting needed)
		buf.Write([]byte(`"`))
		buf.Write(b)
		buf.Write([]byte(`"`))
	} else {
		// Input is JSON (append as-is)
		buf.Write(b)
	}

	msg := strings.Builder{}
	msg.Grow(buf.Len())
	msg.Write(buf.Bytes())

	sp := splunkPayload{
		Time: time.Now().Unix(),
		Host: sl.hostname,
		Event: splunkEvent{
			Message: msg.String(),
			Ident:   sl.source,
			Host:    sl.hostname,
		},
	}

	// Text template is likely faster than JSON marshalling, however, not as fast as
	// building the string manually. I do not have time for this though and it will look
	// a bit ugly.

	buf.Truncate(0) // keep the pre-allocated capacity
	err := recordTemplate.Execute(buf, sp)
	if err != nil {
		return err
	}

	payload := make([]byte, buf.Len())
	copy(payload, buf.Bytes())

	select {
	case sl.payloads <- payload:
	default:
		return ErrFullOrClosed
	}
	return nil
}
