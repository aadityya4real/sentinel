package config

import (
	"strings"
	"testing"
)

func TestValidateAcceptsCompleteConfig(t *testing.T) {
	cfg := &Config{
		Port:         "8080",
		Environment:  "development",
		APIToken:     strings.Repeat("t", 32),
		DatabaseURL:  "postgres://sentinel:sentinel@localhost:5432/sentinel?sslmode=disable",
		RedisAddress: "localhost:6379",
		AIEnabled:    false,
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestValidateAcceptsAIEnabledWhenAllFieldsSet(t *testing.T) {
	cfg := &Config{
		Port:         "8080",
		Environment:  "development",
		APIToken:     strings.Repeat("t", 32),
		DatabaseURL:  "postgres://sentinel:sentinel@localhost:5432/sentinel",
		RedisAddress: "localhost:6379",
		AIEnabled:    true,
		AIBaseURL:    "https://api.openai.com/v1",
		AIAPIKey:     "sk-test",
		AIModel:      "gpt-5-mini",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error = %v, want nil", err)
	}
}

func TestValidateRejectsMissingPort(t *testing.T) {
	cfg := &Config{
		Port:         "",
		Environment:  "development",
		APIToken:     strings.Repeat("t", 32),
		DatabaseURL:  "postgres://localhost/sentinel",
		RedisAddress: "localhost:6379",
	}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "port") {
		t.Fatalf("Validate() error = %v, want port error", err)
	}
}

func TestValidateRejectsMissingDatabaseURL(t *testing.T) {
	cfg := &Config{
		Port:         "8080",
		Environment:  "development",
		APIToken:     strings.Repeat("t", 32),
		DatabaseURL:  "",
		RedisAddress: "localhost:6379",
	}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "database URL") {
		t.Fatalf("Validate() error = %v, want database URL error", err)
	}
}

func TestValidateRejectsMissingRedisAddress(t *testing.T) {
	cfg := &Config{
		Port:         "8080",
		Environment:  "development",
		APIToken:     strings.Repeat("t", 32),
		DatabaseURL:  "postgres://localhost/sentinel",
		RedisAddress: "",
	}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "Redis") {
		t.Fatalf("Validate() error = %v, want Redis address error", err)
	}
}

func TestValidateRejectsAIEnabledWithoutAPIKey(t *testing.T) {
	cfg := &Config{
		Port:         "8080",
		Environment:  "development",
		APIToken:     strings.Repeat("t", 32),
		DatabaseURL:  "postgres://localhost/sentinel",
		RedisAddress: "localhost:6379",
		AIEnabled:    true,
		AIBaseURL:    "https://api.openai.com/v1",
		AIModel:      "gpt-5-mini",
	}
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "AI API key") {
		t.Fatalf("Validate() error = %v, want AI API key error", err)
	}
}

func TestValidateRequiresConfiguredOriginsInProduction(t *testing.T) {
	cfg := &Config{Port: "8080", Environment: "production", APIToken: strings.Repeat("x", 32), DatabaseURL: "postgres://localhost/sentinel", RedisAddress: "localhost:6379"}
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "SENTINEL_ALLOWED_ORIGINS") {
		t.Fatalf("Validate() error=%v, want production origins error", err)
	}
	cfg.AllowedOrigins = []string{"https://dashboard.example"}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() error=%v, want nil", err)
	}
}

func TestParseAllowedOrigins(t *testing.T) {
	origins, err := parseAllowedOrigins(" https://dashboard.example, http://localhost:3000,https://dashboard.example ")
	if err != nil {
		t.Fatal(err)
	}
	if len(origins) != 2 || origins[0] != "https://dashboard.example" || origins[1] != "http://localhost:3000" {
		t.Fatalf("origins=%v", origins)
	}
	for _, invalid := range []string{"*", "https://dashboard.example/path", "javascript:alert(1)"} {
		if _, err := parseAllowedOrigins(invalid); err == nil {
			t.Errorf("accepted invalid origin %q", invalid)
		}
	}
}
