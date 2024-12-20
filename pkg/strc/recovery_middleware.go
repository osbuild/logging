package strc

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
)

// RecoverPanicMiddleware is a middleware that recovers from panics and logs them using slog.
func RecoverPanicMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				buf := make([]byte, 2048)
				n := runtime.Stack(buf, false)
				buf = buf[:n]

				slog.ErrorContext(r.Context(), "panic recovered",
					slog.String("error", fmt.Sprintf("%v", err)),
					slog.String("stack", string(buf)),
				)
				w.WriteHeader(500)
				_, _ = w.Write([]byte{})
			}
		}()

		next.ServeHTTP(w, r)
	})
}
