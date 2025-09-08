package main

import (
	"log/slog"
	"net/http"
	"testing"

	"github.com/osbuild/logging/pkg/collect"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	logs := collect.NewTestHandler(slog.LevelDebug, false, false, true)

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:8131/", nil)
	req.Header.Set("X-Request-Id", "12345")
	request(req, logs)

	require.Equal(t, 18, logs.Count())

	require.Equal(t, 18, logs.CountWith("build_id"))
	require.Equal(t, 5, logs.CountWith("request_id"))
	require.Equal(t, 6, logs.CountWith("trace_id"))
	require.Equal(t, 1, logs.CountWith("logrus"))
	require.Equal(t, 2, logs.CountWith("echo"))

	require.Equal(t, 12, logs.CountWith("span", "trace"))
	require.Equal(t, 12, logs.CountWith("span", "id"))

	require.Equal(t, 2, logs.CountWith("request", "method"))
	require.Equal(t, 2, logs.CountWith("request", "path"))
	require.Equal(t, 2, logs.CountWith("request", "length"))

	require.Equal(t, 2, logs.CountWith("response", "status"))
	require.Equal(t, 2, logs.CountWith("response", "length"))

	tids := logs.CollectWith("trace_id")
	tids = append(tids, logs.CollectWith("span", "trace")...)
	require.Len(t, tids, 18)
	for _, v := range tids {
		require.Equal(t, tids[0], v)
	}
}
