package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

const requestTimeout = 15 * time.Second

// Chain returns the ordered Sentinel middleware stack applied to every request:
// RequestID, RealIP, Recovery, Logging, CORS, and request Timeout.
func Chain(logger *zap.Logger) []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.RequestID,
		middleware.RealIP,
		Recovery(logger),
		Logging(logger),
		CORS(nil),
		middleware.Timeout(requestTimeout),
	}
}
