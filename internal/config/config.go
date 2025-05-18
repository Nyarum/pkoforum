package config

import (
	"fmt"
	"os"
)

// Config holds all configuration for the application
type Config struct {
	DeepseekAPIKey string
	DeepseekURL    string
	UploadsPath    string
	Port           string
}

// Load returns a Config struct populated with values from environment variables
func Load() (*Config, error) {
	apiKey := os.Getenv("DEEPSEEK_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DEEPSEEK_API_KEY environment variable is required")
	}

	config := &Config{
		DeepseekAPIKey: apiKey,
		DeepseekURL:    getEnvWithDefault("DEEPSEEK_URL", "https://api.deepseek.com"),
		UploadsPath:    getEnvWithDefault("UPLOADS_PATH", "static/uploads"),
		Port:           getEnvWithDefault("PORT", "8080"),
	}

	return config, nil
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
