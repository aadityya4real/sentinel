// Package middleware provides HTTP middleware for the Sentinel API server.
package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
)

// Recovery recovers from panics in downstream handlers, logs the stack trace,
// and returns a 500 response so a single failing request cannot crash the process.
func Recovery(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered",
						zap.Any("panic", rec),
						zap.ByteString("stack", debug.Stack()),
					)
					http.Error(writer, fmt.Sprintf("internal server error: %v", rec), http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(writer, request)
		})
	}
}
