package config

import (
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port            string `mapstructure:"port"`
	LogLevel        string `mapstructure:"log_level"`
	GoogleAPIKey    string `mapstructure:"google_api_key"`
	GoogleSearchID  string `mapstructure:"google_search_id"`
	MaxResults      int    `mapstructure:"max_results"`
	RateLimitPerMin int    `mapstructure:"rate_limit_per_min"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/mcp-google-search/")
	viper.AddConfigPath("$HOME/.mcp-google-search/")

	// Set defaults
	viper.SetDefault("port", "8080")
	viper.SetDefault("log_level", "info")
	viper.SetDefault("max_results", 10)
	viper.SetDefault("rate_limit_per_min", 100)

	// Read from environment variables
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	// Override with environment variables
	if val := os.Getenv("GOOGLE_API_KEY"); val != "" {
		config.GoogleAPIKey = val
	}
	if val := os.Getenv("GOOGLE_SEARCH_ID"); val != "" {
		config.GoogleSearchID = val
	}
	if val := os.Getenv("PORT"); val != "" {
		config.Port = val
	}

	return &config, nil
}
