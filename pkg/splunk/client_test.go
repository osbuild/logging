package splunk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"
	"time"
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
	time.AfterFunc(time.Second*10, func() {
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
			"host": "test-host",
			"event": map[string]any{
				"message": "message",
				"ident":   "image-builder",
				"host":    "test-host",
			},
		}
		if !reflect.DeepEqual(want, m) {
			t.Errorf("got %v, want %v", m, want)
		}
		ch <- true
	}))

	sl := newSplunkLogger(context.Background(), srv.URL, "token", "image-builder", "test-host")
	err := sl.event([]byte("message"))
	if err != nil {
		t.Error(err)
	}

	sl.close()
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
			"host": "test-host",
			"event": map[string]any{
				"message": "message",
				"ident":   "image-builder",
				"host":    "test-host",
			},
		}
		if !reflect.DeepEqual(want, m) {
			t.Errorf("got %v, want %v", m, want)
		}
		ch <- true
	}))

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	sl := newSplunkLogger(ctx, srv.URL, "token", "image-builder", "test-host")
	err := sl.event([]byte("message"))
	if err != nil {
		t.Error(err)
	}

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
			name: "just string",
			f: func() error {
				sl := newSplunkLogger(context.Background(), url, "token", "osbuild", "hostname")
				defer sl.close()
				err := sl.event([]byte("message"))
				if err != nil {
					return err
				}
				return nil
			},
			want: map[string]any{
				"time": 0,
				"host": "hostname",
				"event": map[string]any{
					"message": "message",
					"ident":   "osbuild",
					"host":    "hostname",
				},
			},
		},
		{
			name: "some json",
			f: func() error {
				sl := newSplunkLogger(context.Background(), url, "token", "osbuild", "hostname")
				defer sl.close()
				err := sl.event([]byte(`{"one": 1, "message": "test"}`))
				if err != nil {
					return err
				}
				return nil
			},
			want: map[string]any{
				"time": 0,
				"host": "hostname",
				"event": map[string]any{
					"message": map[string]any{
						"message": "message",
					},
					"ident": "osbuild",
					"host":  "hostname",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wg := sync.WaitGroup{}
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") != "Splunk token" {
					t.Errorf("got %v, want Splunk token", r.Header.Get("Authorization"))
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("got %v, want application/json", r.Header.Get("Content-Type"))
				}
				m := decodeBody(t, r)
				if !reflect.DeepEqual(tt.want, m) {
					t.Errorf("got %v, want %v", m, tt.want)
				}
				wg.Done()
			}))
			url = srv.URL
			defer srv.Close()

			wg.Add(1)
			err := tt.f()
			if err != nil {
				t.Error(err)
			}
			wg.Wait()
		})
	}
}
