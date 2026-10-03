package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"strings"
)

// RequireBearerToken protects an API operation with the configured server token.
// Hashing both values ensures the constant-time comparison always uses equal-length inputs.
func RequireBearerToken(expected string) func(http.Handler) http.Handler {
	expectedHash := sha256.Sum256([]byte(expected))
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			scheme, token, ok := strings.Cut(request.Header.Get("Authorization"), " ")
			providedHash := sha256.Sum256([]byte(token))
			matches := subtle.ConstantTimeCompare(expectedHash[:], providedHash[:]) == 1
			valid := ok && strings.EqualFold(scheme, "Bearer") && token != "" && !strings.ContainsAny(token, " \t\r\n") && matches && expected != ""
			if !valid {
				writer.Header().Set("WWW-Authenticate", `Bearer realm="sentinel"`)
				writer.Header().Set("Content-Type", "application/json; charset=utf-8")
				writer.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(writer).Encode(map[string]any{"error": map[string]string{"code": "unauthorized", "message": "bearer token is required"}})
				return
			}
			next.ServeHTTP(writer, request)
		})
	}
}
