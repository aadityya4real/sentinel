package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Port string
}

func Load() (*Config, error) {
	viper.SetDefault("PORT", "8080")
	viper.AutomaticEnv()

	return &Config{
		Port: viper.GetString("PORT"),
	}, nil
}
