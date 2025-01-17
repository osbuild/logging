package strc

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
)

// RecoverPanicMiddleware is a middleware that recovers from panics and logs them using slog
// as errors with status code 500. No body is returned.
func RecoverPanicMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					buf := make([]byte, 2048)
					n := runtime.Stack(buf, false)
					buf = buf[:n]
					msg := fmt.Sprintf("%v", err)

					logger.ErrorContext(r.Context(), "panic: "+msg,
						slog.String("error", msg),
						slog.String("stack", string(buf)),
					)
					w.WriteHeader(500)
					_, _ = w.Write([]byte{})
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
