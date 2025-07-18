package sinit

import (
	"errors"
	"log"
)

type logLoggerWriter struct {
	dest *log.Logger
}

var ErrLoggerNotInitialized = errors.New("log logger is not initialized")

func (w *logLoggerWriter) Write(p []byte) (n int, err error) {
	if w.dest == nil {
		return 0, ErrLoggerNotInitialized
	}

	w.dest.Output(2, string(p))
	return len(p), nil
}
