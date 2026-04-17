package strc

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTracingDoerBodyDumping(t *testing.T) {
	tests := []struct {
		name         string
		config       TracingDoerConfig
		reqBody      string
		respBody     string
		wantReqBody  string
		wantRespBody string
	}{
		{
			name:         "request body only",
			config:       TracingDoerConfig{WithRequestBody: true},
			reqBody:      "request data",
			respBody:     "response data",
			wantReqBody:  "request data",
			wantRespBody: "response data",
		},
		{
			name:         "response body only",
			config:       TracingDoerConfig{WithResponseBody: true},
			reqBody:      "request data",
			respBody:     "response data",
			wantReqBody:  "request data",
			wantRespBody: "response data",
		},
		{
			name: "both bodies",
			config: TracingDoerConfig{
				WithRequestBody:  true,
				WithResponseBody: true,
			},
			reqBody:      "request data",
			respBody:     "response data",
			wantReqBody:  "request data",
			wantRespBody: "response data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if string(body) != tt.wantReqBody {
					t.Errorf("server received body = %q, want %q", string(body), tt.wantReqBody)
				}
				w.Write([]byte(tt.respBody))
			}))
			defer server.Close()

			logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
			slog.SetDefault(logger)
			SetLogger(logger)

			doer := NewTracingDoerWithConfig(server.Client(), tt.config)
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, strings.NewReader(tt.reqBody))

			resp, err := doer.Do(req)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			if string(body) != tt.wantRespBody {
				t.Errorf("response body = %q, want %q", string(body), tt.wantRespBody)
			}
		})
	}
}

func TestTracingDoerHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(TraceHTTPHeaderName) == "" {
			t.Error("trace header not set")
		}
		if r.Header.Get(SpanHTTPHeaderName) == "" {
			t.Error("span header not set")
		}
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
	SetLogger(logger)

	ctx := WithTraceID(context.Background(), NewTraceID())
	doer := NewTracingDoer(server.Client())

	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	resp, err := doer.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.Header.Get("X-Custom") != "value" {
		t.Error("custom header not found in response")
	}
}

func TestTracingDoerError(t *testing.T) {
	doer := NewTracingDoer(&errorDoer{})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com", nil)

	_, err := doer.Do(req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var doerErr *DoerErr
	if !errors.As(err, &doerErr) {
		t.Errorf("error should be DoerErr, got %T", err)
	}

	if !errors.Is(err, http.ErrAbortHandler) {
		t.Error("should be able to unwrap to original error")
	}
}

type errorDoer struct{}

func (d *errorDoer) Do(req *http.Request) (*http.Response, error) {
	return nil, http.ErrAbortHandler
}
