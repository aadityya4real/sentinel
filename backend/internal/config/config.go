// Package config loads Sentinel server configuration.
package config

import (
	"fmt"
	"net"
	"net/url"

	"github.com/spf13/viper"
)

// Config contains the runtime configuration required by the Sentinel server.
type Config struct {
	Port          string
	DatabaseURL   string
	RedisAddress  string
	RedisPassword string
	AIEnabled     bool
	AIBaseURL     string
	AIAPIKey      string
	AIModel       string
}

// Load reads server configuration from environment variables and applies safe local defaults.
func Load() (*Config, error) {
	v := viper.New()
	v.SetDefault("port", "8080")
	v.SetDefault("postgres_host", "localhost")
	v.SetDefault("postgres_port", "5432")
	v.SetDefault("postgres_user", "sentinel")
	v.SetDefault("postgres_password", "sentinel")
	v.SetDefault("postgres_db", "sentinel")
	v.SetDefault("postgres_sslmode", "disable")
	v.SetDefault("redis_host", "localhost")
	v.SetDefault("redis_port", "6379")
	v.SetDefault("redis_password", "")
	v.SetDefault("ai_enabled", false)
	v.SetDefault("ai_base_url", "https://api.openai.com/v1")
	v.SetDefault("ai_model", "gpt-5-mini")
	v.AutomaticEnv()

	for key, names := range map[string][]string{
		"port":              {"APP_PORT", "PORT"},
		"database_url":      {"DATABASE_URL"},
		"postgres_host":     {"POSTGRES_HOST"},
		"postgres_port":     {"POSTGRES_PORT"},
		"postgres_user":     {"POSTGRES_USER"},
		"postgres_password": {"POSTGRES_PASSWORD"},
		"postgres_db":       {"POSTGRES_DB"},
		"postgres_sslmode":  {"POSTGRES_SSLMODE"},
		"redis_host":        {"REDIS_HOST"},
		"redis_port":        {"REDIS_PORT"},
		"redis_password":    {"REDIS_PASSWORD"},
		"ai_enabled":        {"AI_ENABLED"},
		"ai_base_url":       {"AI_BASE_URL"},
		"ai_api_key":        {"AI_API_KEY"},
		"ai_model":          {"AI_MODEL"},
	} {
		if err := v.BindEnv(append([]string{key}, names...)...); err != nil {
			return nil, fmt.Errorf("bind configuration for %q: %w", key, err)
		}
	}

	databaseURL := v.GetString("database_url")
	if databaseURL == "" {
		databaseURL = (&url.URL{
			Scheme: "postgres",
			User:   url.UserPassword(v.GetString("postgres_user"), v.GetString("postgres_password")),
			Host:   net.JoinHostPort(v.GetString("postgres_host"), v.GetString("postgres_port")),
			Path:   v.GetString("postgres_db"),
			RawQuery: url.Values{
				"sslmode": {v.GetString("postgres_sslmode")},
			}.Encode(),
		}).String()
	}

	return &Config{
		Port:          v.GetString("port"),
		DatabaseURL:   databaseURL,
		RedisAddress:  net.JoinHostPort(v.GetString("redis_host"), v.GetString("redis_port")),
		RedisPassword: v.GetString("redis_password"),
		AIEnabled:     v.GetBool("ai_enabled"),
		AIBaseURL:     v.GetString("ai_base_url"),
		AIAPIKey:      v.GetString("ai_api_key"),
		AIModel:       v.GetString("ai_model"),
	}, nil
}
