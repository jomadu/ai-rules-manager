package config

import (
	"github.com/spf13/viper"
)

// Config holds the application configuration
type Config struct {
	Targets      []string          `mapstructure:"targets"`
	Dependencies map[string]string `mapstructure:"dependencies"`
}

// Load loads configuration from files and environment
func Load() (*Config, error) {
	viper.SetConfigName("rules")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	var config Config
	if err := viper.ReadInConfig(); err != nil {
		return &config, err
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
