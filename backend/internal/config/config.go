// Package config loads Sentinel server configuration.
package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/spf13/viper"
)

// Config contains the runtime configuration required by the Sentinel server.
type Config struct {
	Port           string
	Environment    string
	APIToken       string
	AllowedOrigins []string
	DatabaseURL    string
	RedisAddress   string
	RedisPassword  string
	AIEnabled      bool
	AIBaseURL      string
	AIAPIKey       string
	AIModel        string
}

// Load reads server configuration from environment variables and applies safe local defaults.
func Load() (*Config, error) {
	v := viper.New()
	v.SetDefault("port", "8080")
	v.SetDefault("app_env", "development")
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
		"port":                     {"APP_PORT", "PORT"},
		"app_env":                  {"APP_ENV"},
		"sentinel_api_token":       {"SENTINEL_API_TOKEN"},
		"sentinel_allowed_origins": {"SENTINEL_ALLOWED_ORIGINS"},
		"database_url":             {"DATABASE_URL"},
		"postgres_host":            {"POSTGRES_HOST"},
		"postgres_port":            {"POSTGRES_PORT"},
		"postgres_user":            {"POSTGRES_USER"},
		"postgres_password":        {"POSTGRES_PASSWORD"},
		"postgres_db":              {"POSTGRES_DB"},
		"postgres_sslmode":         {"POSTGRES_SSLMODE"},
		"redis_host":               {"REDIS_HOST"},
		"redis_port":               {"REDIS_PORT"},
		"redis_password":           {"REDIS_PASSWORD"},
		"ai_enabled":               {"AI_ENABLED"},
		"ai_base_url":              {"AI_BASE_URL"},
		"ai_api_key":               {"AI_API_KEY"},
		"ai_model":                 {"AI_MODEL"},
	} {
		if err := v.BindEnv(append([]string{key}, names...)...); err != nil {
			return nil, fmt.Errorf("bind configuration for %q: %w", key, err)
		}
	}

	environment := strings.ToLower(strings.TrimSpace(v.GetString("app_env")))
	origins, err := parseAllowedOrigins(v.GetString("sentinel_allowed_origins"))
	if err != nil {
		return nil, fmt.Errorf("invalid CORS configuration: %w", err)
	}
	if len(origins) == 0 && environment != "production" {
		origins = []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:8080"}
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

	cfg := &Config{
		Port:           v.GetString("port"),
		Environment:    environment,
		APIToken:       v.GetString("sentinel_api_token"),
		AllowedOrigins: origins,
		DatabaseURL:    databaseURL,
		RedisAddress:   net.JoinHostPort(v.GetString("redis_host"), v.GetString("redis_port")),
		RedisPassword:  v.GetString("redis_password"),
		AIEnabled:      v.GetBool("ai_enabled"),
		AIBaseURL:      v.GetString("ai_base_url"),
		AIAPIKey:       v.GetString("ai_api_key"),
		AIModel:        v.GetString("ai_model"),
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return cfg, nil
}

// Validate ensures the configuration contains all values required to start the server.
func (c *Config) Validate() error {
	if strings.TrimSpace(c.Port) == "" {
		return errors.New("server port is required")
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return errors.New("database URL is required")
	}
	if strings.TrimSpace(c.RedisAddress) == "" {
		return errors.New("Redis address is required")
	}
	if len(c.APIToken) < 32 || strings.TrimSpace(c.APIToken) != c.APIToken {
		return errors.New("SENTINEL_API_TOKEN must be at least 32 characters and contain no surrounding whitespace")
	}
	if strings.EqualFold(strings.TrimSpace(c.Environment), "production") && len(c.AllowedOrigins) == 0 {
		return errors.New("SENTINEL_ALLOWED_ORIGINS must be configured in production")
	}
	if strings.TrimSpace(c.Environment) == "" {
		return errors.New("APP_ENV is required")
	}
	if _, err := parseAllowedOrigins(strings.Join(c.AllowedOrigins, ",")); err != nil {
		return fmt.Errorf("invalid allowed origins: %w", err)
	}
	if c.AIEnabled {
		if strings.TrimSpace(c.AIAPIKey) == "" {
			return errors.New("AI API key is required when AI is enabled")
		}
		if strings.TrimSpace(c.AIBaseURL) == "" {
			return errors.New("AI base URL is required when AI is enabled")
		}
		if strings.TrimSpace(c.AIModel) == "" {
			return errors.New("AI model is required when AI is enabled")
		}
	}
	return nil
}

func parseAllowedOrigins(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" || strings.Contains(parsed.Host, "*") || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") || origin == "*" {
			return nil, fmt.Errorf("%q is not a valid HTTP(S) origin", origin)
		}
		origin = parsed.Scheme + "://" + parsed.Host
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}
	return origins, nil
}
