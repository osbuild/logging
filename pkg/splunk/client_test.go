package splunk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func decodeBody(t *testing.T, r *http.Request) map[string]any {
	m := map[string]any{}
	err := json.NewDecoder(r.Body).Decode(&m)
	if err != nil {
		t.Error(err)
	}
	m["time"] = 0
	return m
}

func TestSplunkLoggerRetry(t *testing.T) {
	var internalErrorOnce sync.Once
	ch := make(chan bool)
	time.AfterFunc(time.Second*8, func() {
		ch <- false
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalErrorOnce.Do(func() {
			// make sure the logger retries requests
			w.WriteHeader(http.StatusInternalServerError)
		})
		if r.Header.Get("Authorization") != "Splunk token" {
			t.Errorf("got %v, want Splunk token", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("got %v, want application/json", r.Header.Get("Content-Type"))
		}
		m := decodeBody(t, r)
		want := map[string]any{
			"time": 0,
			"host": "hostname",
			"event": map[string]any{
				"message": map[string]any{},
				"ident":   "source",
				"host":    "hostname",
			},
		}
		if !reflect.DeepEqual(want, m) {
			t.Errorf("got %v, want %v", m, want)
		}
		ch <- true
	}))
	defer srv.Close()

	sl := newSplunkLogger(context.Background(), srv.URL, "token", "source", "hostname", 0)
	_, err := sl.event([]byte("{}\n"))
	if err != nil {
		t.Error(err)
	}
	sl.close(100 * time.Millisecond)

	if !<-ch {
		t.Error("timeout")
	}
}

func TestSplunkLoggerContext(t *testing.T) {
	ch := make(chan bool)
	time.AfterFunc(time.Second*10, func() {
		ch <- false
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Splunk token" {
			t.Errorf("got %v, want Splunk token", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("got %v, want application/json", r.Header.Get("Content-Type"))
		}
		m := decodeBody(t, r)
		want := map[string]any{
			"time": 0,
			"host": "hostname",
			"event": map[string]any{
				"message": map[string]any{},
				"ident":   "source",
				"host":    "hostname",
			},
		}
		if !reflect.DeepEqual(want, m) {
			t.Errorf("got %v, want %v", m, want)
		}
		ch <- true
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	sl := newSplunkLogger(ctx, srv.URL, "token", "source", "hostname", 0)
	_, err := sl.event([]byte("{}\n"))
	if err != nil {
		t.Error(err)
	}
	sl.close(100 * time.Millisecond)

	if !<-ch {
		t.Error("timeout")
	}
}

func TestSplunkLoggerPayloads(t *testing.T) {
	var url string

	tests := []struct {
		name string
		f    func() error
		want map[string]any
	}{
		{
			name: "empty",
			f: func() error {
				sl := newSplunkLogger(context.Background(), url, "token", "source", "hostname", 0)
				defer sl.close(100 * time.Millisecond)
				_, err := sl.event([]byte("{}\n"))
				if err != nil {
					return err
				}
				return nil
			},
			want: map[string]any{
				"time": 0,
				"host": "hostname",
				"event": map[string]any{
					"message": map[string]any{},
					"ident":   "source",
					"host":    "hostname",
				},
			},
		},
		{
			name: "json",
			f: func() error {
				sl := newSplunkLogger(context.Background(), url, "token", "source", "hostname", 0)
				defer sl.close(100 * time.Millisecond)
				_, err := sl.event([]byte(`{"a": "b"}` + "\n"))
				if err != nil {
					return err
				}
				return nil
			},
			want: map[string]any{
				"time": 0,
				"host": "hostname",
				"event": map[string]any{
					"ident": "source",
					"host":  "hostname",
					"message": map[string]any{
						"a": "b",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Splunk token" {
					t.Errorf("got %v, want Splunk token", r.Header.Get("Authorization"))
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("got %v, want application/json", r.Header.Get("Content-Type"))
				}
				m := decodeBody(t, r)
				if !reflect.DeepEqual(tt.want, m) {
					t.Errorf("got %+v, want %+v", m, tt.want)
				}
			}))
			url = srv.URL
			defer srv.Close()

			err := tt.f()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestSplunkLoggerCloseTimeout(t *testing.T) {
	stopCh := make(chan struct{})

	var handlerCalls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalls.Add(1)
		select {
		case <-time.After(10 * time.Second):
			return
		case <-stopCh:
			return
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	sl := newSplunkLogger(ctx, srv.URL, "token", "source", "hostname", 0)
	_, err := sl.event([]byte("{}\n"))
	if err != nil {
		t.Error(err)
	}
	start := time.Now()
	closeTimeout := 100 * time.Millisecond

	err = sl.close(closeTimeout)
	assert.Equal(t, ErrCloseTimeout, err)

	// add a lot of slack to the closeTimeout to account for slow  GH VMs that may run the tests
	assert.True(t, time.Since(start) < 5*closeTimeout)

	close(stopCh)
	assert.Equal(t, int32(1), handlerCalls.Load())
}
