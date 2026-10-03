package websocket

import (
	"net/http/httptest"
	"testing"
)

func TestOriginAllowedUsesConfiguredOrigins(t *testing.T) {
	allowed := []string{"http://localhost:3000", "https://dashboard.example"}
	for _, test := range []struct {
		origin string
		want   bool
	}{
		{origin: "http://localhost:3000", want: true},
		{origin: "https://dashboard.example", want: true},
		{origin: "https://evil.example", want: false},
	} {
		request := httptest.NewRequest("GET", "http://sentinel.example/ws/v1/metrics", nil)
		request.Header.Set("Origin", test.origin)
		if got := OriginAllowed(request, allowed); got != test.want {
			t.Errorf("origin %q allowed=%t, want %t", test.origin, got, test.want)
		}
	}
}
