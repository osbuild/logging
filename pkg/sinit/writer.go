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

	// Write the log message using the provided logger with call depth of 2
	w.dest.Output(2, string(p))
	return len(p), nil
}
