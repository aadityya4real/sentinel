package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const requestTimeout = 15 * time.Second

// Chain returns the ordered Sentinel middleware stack applied to every request:
// RequestID, RealIP, Recovery, Logging, CORS, and request Timeout (except WebSocket upgrades).
func Chain(logger *zap.Logger, allowedOrigins []string) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		Recovery(logger),
		Logging(logger),
		CORS(allowedOrigins),
		RequestTimeout,
	}
}

// RequestTimeout applies the HTTP deadline to ordinary requests while keeping
// the original ResponseWriter for WebSocket upgrades, which need Hijacker.
func RequestTimeout(next http.Handler) http.Handler {
	timed := middleware.Timeout(requestTimeout)(next)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method == http.MethodGet && headerHasToken(request.Header.Get("Connection"), "upgrade") && strings.EqualFold(request.Header.Get("Upgrade"), "websocket") {
			next.ServeHTTP(writer, request)
			return
		}
		timed.ServeHTTP(writer, request)
	})
}

func headerHasToken(value, token string) bool {
	for _, part := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(part), token) {
			return true
		}
	}
	return false
}
