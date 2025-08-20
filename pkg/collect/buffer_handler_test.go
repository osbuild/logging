package collect_test

import (
	"encoding/json"
	"log/slog"
	"testing"
	"testing/slogtest"

	"github.com/google/go-cmp/cmp"
	"github.com/osbuild/logging/pkg/collect"
)

func TestStandardLibraryHelper(t *testing.T) {
	var result []map[string]any
	th := collect.NewTestHandler(slog.LevelDebug, true, true, true)
	err := slogtest.TestHandler(th, func() []map[string]any {
		result = th.All()
		return parseLogEntries(t, result)
	})
	if err != nil {
		t.Error(err)
		json, _ := json.MarshalIndent(result, "", " ")
		t.Log(string(json))
	}
}

func parseLogEntries(_ *testing.T, ms []map[string]any) []map[string]any {
	return ms
}

func TestLast(t *testing.T) {
	h := collect.NewTestHandler(slog.LevelDebug, false, false, false)
	logger := slog.New(h)

	tests := []struct {
		f    func()
		want map[string]any
	}{
		{
			f: func() {
				logger.Debug("test", "key", "value")
			},
			want: map[string]any{
				"msg": "test",
				"key": "value",
			},
		},
		{
			f: func() {
				logger.Debug("test", slog.Group("g", "key", "value"))
			},
			want: map[string]any{
				"msg": "test",
				"g": map[string]any{
					"key": "value",
				},
			},
		},
		{
			f: func() {
				logger.WithGroup("g").Debug("test", "key", "value")
			},
			want: map[string]any{
				"msg": "test",
				"g": map[string]any{
					"key": "value",
				},
			},
		},
	}

	for _, tt := range tests {
		tt.f()
		got := h.Last()

		if !cmp.Equal(got, tt.want) {
			t.Errorf("Got: %v, want: %v", got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	h := collect.NewTestHandler(slog.LevelDebug, false, false, true)
	logger := slog.New(h)

	logger.Info("test message", "key1", "value1", "group1", slog.GroupValue(slog.String("key2", "value2")))

	if !h.Contains("test message", slog.MessageKey) {
		t.Errorf("Expected to find msg with value test message")
	}

	if !h.Contains("INFO", slog.LevelKey) {
		t.Errorf("Expected to find level with value INFO")
	}

	if !h.Contains("value1", "key1") {
		t.Errorf("Expected to find key1 with value value1")
	}

	if !h.Contains("value2", "group1", "key2") {
		t.Errorf("Expected to find group1.key2 with value value2")
	}

	if h.Contains("nonexistent", "key1") {
		t.Errorf("Did not expect to find key1 with value nonexistent")
	}

	if h.Contains("value1", "nonexistent") {
		t.Errorf("Did not expect to find nonexistent key")
	}
}

func TestContainsPanic(t *testing.T) {
	h := collect.NewTestHandler(slog.LevelDebug, false, false, true)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic, got none")
		}
	}()

	h.Contains("test message")
}
