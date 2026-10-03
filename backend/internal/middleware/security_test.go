package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireBearerToken(t *testing.T) {
	const secret = "sentinel-test-token-which-is-at-least-32-chars"
	handler := RequireBearerToken(secret)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	tests := []struct {
		name, authorization string
		want                int
	}{
		{name: "missing", want: http.StatusUnauthorized},
		{name: "wrong", authorization: "Bearer not-the-secret", want: http.StatusUnauthorized},
		{name: "malformed", authorization: "Basic " + secret, want: http.StatusUnauthorized},
		{name: "correct", authorization: "Bearer " + secret, want: http.StatusNoContent},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if test.authorization != "" {
				request.Header.Set("Authorization", test.authorization)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != test.want {
				t.Fatalf("status=%d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestCORSAllowsConfiguredOriginAndRejectsUnknownOrigin(t *testing.T) {
	handler := CORS([]string{"https://dashboard.example"})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for _, test := range []struct {
		origin      string
		want        int
		allowHeader string
	}{
		{origin: "https://dashboard.example", want: http.StatusNoContent, allowHeader: "https://dashboard.example"},
		{origin: "https://evil.example", want: http.StatusForbidden},
	} {
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Header.Set("Origin", test.origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != test.want {
			t.Fatalf("origin %q status=%d, want %d", test.origin, response.Code, test.want)
		}
		if got := response.Header().Get("Access-Control-Allow-Origin"); got != test.allowHeader {
			t.Fatalf("allow origin=%q, want %q", got, test.allowHeader)
		}
	}
}
