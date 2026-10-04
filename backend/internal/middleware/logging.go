package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// statusRecorder captures the HTTP status code written by downstream handlers
// so it can be included in the request log.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Hijack keeps WebSocket upgrades working through request logging middleware.
func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("underlying response writer does not support hijacking")
	}
	connection, buffered, err := hijacker.Hijack()
	if err == nil {
		r.status = http.StatusSwitchingProtocols
	}
	return connection, buffered, err
}

// Unwrap exposes the underlying writer to http.ResponseController.
func (r *statusRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// Logging records each HTTP request with its method, path, status, and duration.
func Logging(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			start := time.Now()
			recorder := &statusRecorder{ResponseWriter: writer, status: http.StatusOK}
			next.ServeHTTP(recorder, request)
			logger.Info("http request",
				zap.String("method", request.Method),
				zap.String("path", request.URL.Path),
				zap.Int("status", recorder.status),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}
