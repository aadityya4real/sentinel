package middleware

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
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

type hijackableRecorder struct{ *httptest.ResponseRecorder }

func (hijackableRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) { return nil, nil, nil }

func TestRequestTimeoutPreservesWebSocketHijacker(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ws/v1/metrics", nil)
	request.Header.Set("Connection", "keep-alive, Upgrade")
	request.Header.Set("Upgrade", "websocket")
	writer := hijackableRecorder{httptest.NewRecorder()}
	called := false
	handler := RequestTimeout(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		called = true
		if _, ok := response.(http.Hijacker); !ok {
			t.Error("WebSocket upgrade writer no longer implements http.Hijacker")
		}
	}))
	handler.ServeHTTP(writer, request)
	if !called {
		t.Fatal("WebSocket upgrade did not reach the handler")
	}
}

func TestRequestTimeoutAddsDeadlineToRegularHTTP(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	called := false
	handler := RequestTimeout(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		called = true
		if _, ok := request.Context().Deadline(); !ok {
			t.Error("regular HTTP request has no timeout deadline")
		}
	}))
	handler.ServeHTTP(httptest.NewRecorder(), request)
	if !called {
		t.Fatal("regular HTTP request did not reach the handler")
	}
}

func TestLoggingPreservesWebSocketHijacker(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/ws/v1/metrics", nil)
	writer := hijackableRecorder{httptest.NewRecorder()}
	called := false
	handler := Logging(zap.NewNop())(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		called = true
		if _, ok := response.(http.Hijacker); !ok {
			t.Error("logging middleware removed http.Hijacker")
		}
	}))
	handler.ServeHTTP(writer, request)
	if !called {
		t.Fatal("WebSocket request did not reach handler")
	}
}
