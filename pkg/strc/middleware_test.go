package strc_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/osbuild/logging/pkg/collect"
	"github.com/osbuild/logging/pkg/strc"
)

func TestMiddlewareGenerateContext(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if strc.TraceIDFromContext(ctx) == strc.EmptyTraceID {
			t.Error("TraceID not set in context")
		}

		if strc.SpanIDFromContext(ctx) != strc.EmptySpanID {
			t.Error("SpanID was set in context")
		}
	})

	middleware := strc.NewMiddlewareWithConfig(slog.New(logHandler), strc.MiddlewareConfig{})
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
}

func TestMiddlewareFilter(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	middleware := strc.NewMiddlewareWithConfig(slog.New(logHandler), strc.MiddlewareConfig{
		Filters: []strc.Filter{strc.IgnorePathPrefix("/metrics")},
	})
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com/metrics", nil)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
	if logHandler.Count() != 0 {
		t.Errorf("Log entries found, expected none: %s", logHandler.All())
	}
}

func TestMiddlewareParseContext(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strc.TraceIDFromContext(r.Context()) != "LOlIxiHprrrvHqD" {
			t.Error("TraceID not parsed into context")
		}

		if strc.SpanIDFromContext(r.Context()) != "VIPEcES.yuufaHI" {
			t.Error("SpanID not parsed into context")
		}
	})

	middleware := strc.NewMiddlewareWithConfig(slog.New(logHandler), strc.MiddlewareConfig{})
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Add("X-Strc-Trace-ID", "LOlIxiHprrrvHqD")
	req.Header.Add("X-Strc-Span-Id", "VIPEcES.yuufaHI")
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
}

func TestMiddlewareCustomLogging(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(logHandler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.DebugContext(r.Context(), "Test log")
	})

	middleware := strc.NewMiddlewareWithConfig(logger, strc.MiddlewareConfig{})
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Add("X-Strc-Trace-ID", "LOlIxiHprrrvHqD")
	req.Header.Add("X-Strc-Span-Id", "VIPEcES.yuufaHI")
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)

	if !logHandler.Contains("Test log", slog.MessageKey) {
		t.Errorf("Message not found: %s", logHandler.All())
	}
}

func TestMiddlewareLogging(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(logHandler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})

	middleware := strc.NewMiddlewareWithConfig(logger, strc.MiddlewareConfig{})
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com/x", nil)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)

	if !logHandler.Contains("200: OK", slog.MessageKey) {
		t.Errorf("Message not found: %s", logHandler.All())
	}

	if !logHandler.Contains("GET", "request", "method") {
		t.Errorf("Request method not found: %s", logHandler.All())
	}

	if !logHandler.Contains("/x", "request", "path") {
		t.Errorf("Request path not found: %s", logHandler.All())
	}

	if !logHandler.Contains("example.com", "request", "host") {
		t.Errorf("Request host not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(0), "request", "length") {
		t.Errorf("Request length not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(200), "response", "status") {
		t.Errorf("Response status not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(0), "response", "length") {
		t.Errorf("Response length not found: %s", logHandler.All())
	}
}

func TestMiddlewareErrorLogging(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(logHandler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "test error", http.StatusInternalServerError)
	})

	middleware := strc.NewMiddlewareWithConfig(logger, strc.MiddlewareConfig{})
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com/x", nil)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)

	if !logHandler.Contains("500: Internal Server Error", slog.MessageKey) {
		t.Errorf("Message not found: %s", logHandler.All())
	}

	if !logHandler.Contains("GET", "request", "method") {
		t.Errorf("Request method not found: %s", logHandler.All())
	}

	if !logHandler.Contains("/x", "request", "path") {
		t.Errorf("Request path not found: %s", logHandler.All())
	}

	if !logHandler.Contains("example.com", "request", "host") {
		t.Errorf("Request host not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(0), "request", "length") {
		t.Errorf("Request length not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(500), "response", "status") {
		t.Errorf("Response status not found: %s", logHandler.All())
	}

	if !logHandler.Contains(int64(11), "response", "length") {
		t.Errorf("Response length not found: %s", logHandler.All())
	}
}

func TestMiddlewareSpanEventing(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	strc.SetLogger(slog.New(logHandler))
	defer strc.SetNoopLogger()

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span, _ := strc.Start(r.Context(), "test span")
		defer span.End()

		span.Event("test event")
	})

	middleware := strc.NewMiddlewareWithConfig(slog.New(logHandler), strc.MiddlewareConfig{})
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)

	if !logHandler.Contains("span test span event test event", slog.MessageKey) {
		t.Errorf("Span event not found: %s", logHandler.All())
	}
}

func TestMiddlewarePanic(t *testing.T) {
	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(logHandler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := strc.RecoverPanicMiddleware(logger)
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Add("X-Strc-Trace-ID", "LOlIxiHprrrvHqD")
	req.Header.Add("X-Strc-Span-Id", "VIPEcES.yuufaHI")
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)

	if !logHandler.Contains("panic: test panic", slog.MessageKey) {
		t.Errorf("Panic event not found: %s", logHandler.All())
	}
}

func TestMiddlewarePairs(t *testing.T) {
	var pair = strc.HeadfieldPair{
		HeaderName: "X-Request-Id",
		FieldName:  "request_id",
	}

	logHandler := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	multiHandler := strc.NewMultiHandlerCustom(
		[]slog.Attr{slog.String("a", "b")},
		strc.HeadfieldPairCallback([]strc.HeadfieldPair{pair}),
		logHandler,
	)
	logger := slog.New(multiHandler)

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.DebugContext(r.Context(), "Test log")

		if strc.FetchValueContext(r.Context(), pair) != "VIPEcES" {
			t.Error("Value not found in context")
		}
	})

	middleware := strc.HeadfieldPairMiddleware([]strc.HeadfieldPair{pair})
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Add(pair.HeaderName, "VIPEcES")
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)

	if !logHandler.Contains("Test log", slog.MessageKey) {
		t.Errorf("Message not found: %s", logHandler.All())
	}

	if !logHandler.Contains("VIPEcES", "request_id") {
		t.Errorf("Request ID not found: %s", logHandler.All())
	}
}

func TestMiddlewareAddsToContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	middleware := strc.NewMiddleware(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid := strc.TraceIDFromContext(r.Context())
		sid := strc.SpanIDFromContext(r.Context())

		if tid == "" || tid == strc.EmptyTraceID {
			t.Errorf("expected trace id to be set")
		}

		if sid != strc.EmptySpanID {
			t.Errorf("expected span id to be empty")
		}
	})

	middleware(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
}

func TestMiddlewareDoesNotAddToContext(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	middleware := strc.NewMiddlewareWithConfig(logger, strc.MiddlewareConfig{
		NoTraceContext: true,
	})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid := strc.TraceIDFromContext(r.Context())
		sid := strc.SpanIDFromContext(r.Context())

		if tid != strc.EmptyTraceID {
			t.Errorf("expected trace id to by empty")
		}

		if sid != strc.EmptySpanID {
			t.Errorf("expected span id to be empty")
		}
	})

	middleware(handler).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
}

func TestMiddlewareAddsTraceCtxAndHeaders(t *testing.T) {
	var tid strc.TraceID

	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	middleware := strc.NewMiddleware(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tid = strc.TraceIDFromContext(r.Context())
	})

	rec := httptest.NewRecorder()
	middleware(handler).ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
	if rec.Header().Get(strc.TraceHTTPHeaderName) != tid.String() {
		t.Errorf("expected trace header to be set to %s", tid.String())
	}
}

func TestMiddlewareSkipsTraceHeader(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
	middleware := strc.NewMiddleware(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// noop
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Add(strc.TraceHTTPHeaderName, strc.NewTraceID().String())
	middleware(handler).ServeHTTP(rec, req)
	if rec.Header().Get(strc.TraceHTTPHeaderName) != "" {
		t.Errorf("expected trace header to be empty")
	}
}

func TestMiddlewareAttributes(t *testing.T) {
	tests := []struct {
		name     string
		config   strc.MiddlewareConfig
		expected []string
	}{
		{
			name:   "default",
			config: strc.MiddlewareConfig{},
			expected: []string{
				`msg="200: OK"`,
				`level=INFO`,
				`request.length=12`,  // "test request" = 12 runes
				`response.length=13`, // "test response" = 13 runes
			},
		},
		{
			name:   "with user agent",
			config: strc.MiddlewareConfig{WithUserAgent: true},
			expected: []string{
				`request.user-agent=`,
			},
		},
		{
			name:   "with request body",
			config: strc.MiddlewareConfig{WithRequestBody: true},
			expected: []string{
				`request.body="test request"`,
				`request.length=12`,
			},
		},
		{
			name:   "with response body",
			config: strc.MiddlewareConfig{WithResponseBody: true},
			expected: []string{
				`response.body="test response"`,
				`response.length=13`,
			},
		},
		{
			name:   "with request header",
			config: strc.MiddlewareConfig{WithRequestHeader: true},
			expected: []string{
				`request.host=example.com`,
				`request.method=POST`,
				`request.path=/`,
			},
		},
		{
			name:   "with response headers",
			config: strc.MiddlewareConfig{WithResponseHeader: true},
			expected: []string{
				`response.status=200`,
				`response.latency=`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sb := strings.Builder{}
			logger := slog.New(slog.NewTextHandler(&sb, &slog.HandlerOptions{}))
			middleware := strc.NewMiddlewareWithConfig(logger, tt.config)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				io.ReadAll(r.Body)

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test response"))
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", strings.NewReader("test request"))
			middleware(handler).ServeHTTP(rec, req)

			for _, line := range tt.expected {
				if !strings.Contains(sb.String(), line) {
					t.Log(sb.String())
					t.Errorf("expected log message to contain '%s'", line)
				}
			}
		})
	}
}
