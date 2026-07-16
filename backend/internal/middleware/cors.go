package middleware

import (
	"net/http"
	"strings"
)

// CORS sets Cross-Origin Resource Sharing headers on responses and short-circuits
// OPTIONS preflight requests. An empty allowedOrigins slice permits all origins.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	origin := "*"
	if len(allowedOrigins) > 0 {
		origin = strings.Join(allowedOrigins, ", ")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			writer.Header().Set("Access-Control-Max-Age", "300")
			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}
