package config

import (
	"strings"
	"testing"
)

func TestValidateAcceptsCompleteConfig(t *testing.T) {
	cfg := &Config{
		Port:         "8080",
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
