package strc_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
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

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if strc.TraceIDFromContext(ctx) != strc.EmptyTraceID {
			t.Error("TraceID not set in context")
		}
	})

	middleware := strc.NewMiddlewareWithConfig(slog.New(logHandler), strc.MiddlewareConfig{
		Filters: []strc.Filter{strc.IgnorePathPrefix("/metrics")},
	})
	handlerToTest := middleware(nextHandler)

	req := httptest.NewRequest("GET", "http://example.com/metrics", nil)
	handlerToTest.ServeHTTP(httptest.NewRecorder(), req)
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

	if !slices.ContainsFunc(logHandler.All(), func(e map[string]any) bool {
		return e["msg"] == "Test log"
	}) {
		t.Error("Log message not found")
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

	if !slices.ContainsFunc(logHandler.All(), func(e map[string]any) bool {
		if e["msg"] != "200: OK" {
			return false
		}
		if r, ok := e["request"]; ok {
			r := r.(map[string]any)
			return r["method"] == "GET" && r["path"] == "/x" && r["host"] == "example.com" && r["length"] == int64(0)
		}
		if r, ok := e["response"]; !ok {
			r := r.(map[string]any)
			return r["status"] == 200 && r["length"] == int64(0)
		}
		return true
	}) {
		t.Error("Log message not found")
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

	if !slices.ContainsFunc(logHandler.All(), func(e map[string]any) bool {
		return e["msg"] == "span test span event test event"
	}) {
		t.Error("Span event not logged")
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

	if !slices.ContainsFunc(logHandler.All(), func(e map[string]any) bool {
		return e["msg"] == "panic: test panic"
	}) {
		t.Error("Log message not found")
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

	if !slices.ContainsFunc(logHandler.All(), func(e map[string]any) bool {
		return e["msg"] == "Test log" && e["request_id"] == "VIPEcES"
	}) {
		t.Error("Log message not found")
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
