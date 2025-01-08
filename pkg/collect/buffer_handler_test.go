package collect

import (
	"encoding/json"
	"log/slog"
	"testing"
	"testing/slogtest"
)

func TestStandardLibraryHelper(t *testing.T) {
	var result []map[string]any
	th := NewTestHandler(slog.LevelDebug, true, true, true)
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
