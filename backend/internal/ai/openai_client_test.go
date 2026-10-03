package ai

import "testing"

func TestNewOpenAICompatibleClientRequiresHTTPSForRemoteBearerEndpoints(t *testing.T) {
	if _, err := NewOpenAICompatibleClient("http://provider.example/v1", "secret", "model", true); err == nil {
		t.Fatal("accepted remote HTTP endpoint with a bearer key")
	}
	if _, err := NewOpenAICompatibleClient("http://localhost:8081/v1", "secret", "model", true); err != nil {
		t.Fatalf("local development HTTP endpoint rejected: %v", err)
	}
	if _, err := NewOpenAICompatibleClient("http://127.0.0.1:8081/v1", "secret", "model", false); err == nil {
		t.Fatal("accepted HTTP endpoint when local HTTP is disabled")
	}
	if _, err := NewOpenAICompatibleClient("https://provider.example/v1", "secret", "model", false); err != nil {
		t.Fatalf("HTTPS endpoint rejected: %v", err)
	}
}
