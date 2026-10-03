package middleware

import (
	"net/http"
)

// CORS permits only configured browser origins and short-circuits valid preflight requests.
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	allow := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allow[origin] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			origin := request.Header.Get("Origin")
			if origin != "" {
				if _, ok := allow[origin]; !ok {
					http.Error(writer, "origin is not allowed", http.StatusForbidden)
					return
				}
				writer.Header().Add("Vary", "Origin")
				writer.Header().Set("Access-Control-Allow-Origin", origin)
			}
			if request.Method == http.MethodOptions {
				if origin == "" {
					http.Error(writer, "origin is required", http.StatusForbidden)
					return
				}
				writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Upgrade, Connection")
				writer.Header().Set("Access-Control-Max-Age", "300")
				writer.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}
