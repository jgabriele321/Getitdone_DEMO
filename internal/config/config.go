package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the application
type Config struct {
	// Telegram configuration
	TelegramToken string

	// OpenRouter configuration
	OpenRouterAPIKey string

	// Google Sheets configuration
	GoogleScriptURL string

	// SendGrid configuration
	SendGridKey string

	// Admin configuration
	AdminTelegramID string

	// Test mode configuration
	TestMode  bool
	TestEmail string

	// Database configuration
	DatabasePath string

	// Server configuration
	Port        string
	Environment string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Debug().Msg("No .env file found, using system environment variables")
	}

	cfg := &Config{
		TelegramToken:    getEnv("TELEGRAM_TOKEN", ""),
		OpenRouterAPIKey: getEnvRequired("OPENROUTER_API_KEY"),
		GoogleScriptURL:  getEnvRequired("GOOGLE_SCRIPT_URL"),
		SendGridKey:      getEnv("SENDGRID_KEY", ""),
		AdminTelegramID:  getEnv("ADMIN_TELEGRAM_ID", "@defibeats"),
		TestMode:         getEnvBool("TEST_MODE", false),
		TestEmail:        getEnv("TEST_EMAIL", ""),
		DatabasePath:     getEnv("DATABASE_PATH", "./cache.db"),
		Port:             getEnv("PORT", "8080"),
		Environment:      getEnv("ENVIRONMENT", "development"),
	}

	// Validate configuration
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	// Log non-sensitive configuration
	log.Info().
		Str("environment", cfg.Environment).
		Str("port", cfg.Port).
		Bool("test_mode", cfg.TestMode).
		Str("database_path", cfg.DatabasePath).
		Msg("Configuration loaded")

	return cfg, nil
}

// validate checks that all required configuration is present
func (c *Config) validate() error {
	if c.OpenRouterAPIKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY is required")
	}
	if c.GoogleScriptURL == "" {
		return fmt.Errorf("GOOGLE_SCRIPT_URL is required")
	}
	// TELEGRAM_TOKEN and SENDGRID_KEY are now optional
	if c.TestMode && c.TestEmail == "" {
		return fmt.Errorf("TEST_EMAIL is required when TEST_MODE is enabled")
	}
	return nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvRequired gets a required environment variable
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Warn().Str("key", key).Msg("Required environment variable not set")
	}
	return value
}

// getEnvBool gets a boolean environment variable
func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}
