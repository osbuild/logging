package splunk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	// PayloadsChannelSize is the size of the channel that holds payloads
	PayloadsChannelSize = 1024

	// SendFrequency is the frequency at which payloads are sent
	SendFrequency = 5
)

type splunkLogger struct {
	client   *http.Client
	url      string
	token    string
	source   string
	hostname string

	payloads chan *splunkPayload
}

type splunkPayload struct {
	// splunk expects unix time in seconds
	Time  int64       `json:"time"`
	Host  string      `json:"host"`
	Event splunkEvent `json:"event"`
}

type splunkEvent struct {
	Message string `json:"message"`
	Ident   string `json:"ident"`
	Host    string `json:"host"`
}

func newSplunkLogger(ctx context.Context, url, token, source, hostname string) *splunkLogger {
	rcl := retryablehttp.NewClient()
	rcl.RetryWaitMin = 50 * time.Millisecond
	rcl.RetryWaitMax = 2 * time.Second
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
	sl.payloads = make(chan *splunkPayload, PayloadsChannelSize)

	go sl.flushPayloads(ctx, ticker.C)

	return sl
}

func (sl *splunkLogger) flushPayloads(context context.Context, ticker <-chan time.Time) {
	// TODO optimize for allocations: https://pkg.go.dev/sync#Pool
	payloads := make([]*splunkPayload, 0, PayloadsChannelSize)

	sendPayloads := func(payloads []*splunkPayload) {
		err := sl.sendPayloads(payloads)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Splunk logger unable to send payloads: %v", err)
		}

		payloads = make([]*splunkPayload, 0, PayloadsChannelSize)
	}

	for {
		select {
		case <-context.Done():
			sendPayloads(payloads)
			return
		case p := <-sl.payloads:
			if p == nil {
				sendPayloads(payloads)
				return
			}

			payloads = append(payloads, p)
			if len(payloads) >= PayloadsChannelSize {
				sendPayloads(payloads)
			}
		case <-ticker:
			sendPayloads(payloads)
		}
	}
}

func (sl *splunkLogger) sendPayloads(payloads []*splunkPayload) error {
	if len(payloads) == 0 {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	for _, pl := range payloads {
		// TODO this is slow, we should be able to write directly to the buffer
		b, err := json.Marshal(pl)
		if err != nil {
			return err
		}

		_, err = buf.Write(b)
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest("POST", sl.url, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Splunk " + sl.token)

	res, err := sl.client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Unable to close response body when sending payloads")
		}
	}()

	if res.StatusCode != http.StatusOK {
		buf := bytes.Buffer{}
		_, err = buf.ReadFrom(res.Body)
		if err != nil {
			return fmt.Errorf("Error forwarding to splunk: parsing response failed: %v", err)
		}
		return fmt.Errorf("Error forwarding to splunk: %s", buf.String())
	}
	return nil
}

func (sl *splunkLogger) close() {
	close(sl.payloads)
}

// ErrFullOrClosed is returned when the payloads channel is full or closed via close().
var ErrFullOrClosed = errors.New("cannot create new splunk event: channel full or closed")

func (sl *splunkLogger) logWithTime(t time.Time, msg string) error {
	sp := splunkPayload{
		Time: t.Unix(),
		Host: sl.hostname,
		Event: splunkEvent{
			Message: msg,
			Ident:   sl.source,
			Host:    sl.hostname,
		},
	}
	select {
	case sl.payloads <- &sp:
	default:
		return ErrFullOrClosed
	}
	return nil
}

func (sl *splunkLogger) event(b []byte) error {
	sp := splunkPayload{
		Time: time.Now().Unix(),
		Host: sl.hostname,
		Event: splunkEvent{
			Message: string(b),
			Ident:   sl.source,
			Host:    sl.hostname,
		},
	}
	select {
	case sl.payloads <- &sp:
	default:
		return ErrFullOrClosed
	}
	return nil
}
