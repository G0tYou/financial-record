package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	ServerPort        string
	GoogleCredentials string
	SpreadsheetID     string
	SheetName         string
	DefaultPhone      string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Try to load .env file, but don't fail if it doesn't exist (common in production)
	_ = godotenv.Load()
	return &Config{
		ServerPort:        getEnv("SERVER_PORT", "8080"),
		GoogleCredentials: getEnv("GOOGLE_CREDENTIALS", ""),
		SpreadsheetID:     getEnv("SPREADSHEET_ID", ""),
		DefaultPhone:      getEnv("DEFAULT_PHONE", ""),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		// Strip surrounding quotes if present
		if len(value) >= 2 && (value[0] == '\'' || value[0] == '"') && value[0] == value[len(value)-1] {
			return value[1 : len(value)-1]
		}
		return value
	}
	return defaultValue
}
